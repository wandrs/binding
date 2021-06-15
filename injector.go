package binding

import (
	"context"
	"net/http"
	"reflect"
	"sync"

	"go.wandrs.dev/inject"
)

var pool = sync.Pool{
	New: func() interface{} {
		return inject.New()
	},
}

type injectorKey struct{}

func Injector(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if a routing context already exists from a parent router.
		injector, _ := r.Context().Value(injectorKey{}).(inject.Injector)
		if injector != nil {
			next.ServeHTTP(w, r)
			return
		}

		injector = pool.Get().(inject.Injector)
		injector.Reset()

		// NOTE: r.WithContext() causes 2 allocations and context.WithValue() causes 1 allocation
		ctx := context.WithValue(r.Context(), injectorKey{}, injector)
		r = r.WithContext(ctx)

		injector.Map(ctx)
		injector.Map(r)
		injector.Map(w)

		// Serve the request and once its done, put the request context back in the sync pool
		next.ServeHTTP(w, r)
		pool.Put(injector)
	})
}

func Inject(fn func(inject.Injector)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			injector, _ := r.Context().Value(injectorKey{}).(inject.Injector)
			if injector == nil {
				panic("chi: register Injector middleware")
			}

			fn(injector)
			next.ServeHTTP(w, r)
		})
	}
}

// Maps the interface{} value based on its immediate type from reflect.TypeOf.
func Map(val interface{}) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			injector, _ := r.Context().Value(injectorKey{}).(inject.Injector)
			if injector == nil {
				panic("chi: register Injector middleware")
			}

			injector.Map(val)
			next.ServeHTTP(w, r)
		})
	}
}

// Maps the interface{} value based on the pointer of an Interface provided.
// This is really only useful for mapping a value as an interface, as interfaces
// cannot at this time be referenced directly without a pointer.
func MapTo(val interface{}, ifacePtr interface{}) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			injector, _ := r.Context().Value(injectorKey{}).(inject.Injector)
			if injector == nil {
				panic("chi: register Injector middleware")
			}

			injector.MapTo(val, ifacePtr)
			next.ServeHTTP(w, r)
		})
	}
}

// Provides a possibility to directly insert a mapping based on type and value.
// This makes it possible to directly map type arguments not possible to instantiate
// with reflect like unidirectional channels.
func Set(typ reflect.Type, val reflect.Value) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			injector, _ := r.Context().Value(injectorKey{}).(inject.Injector)
			if injector == nil {
				panic("chi: register Injector middleware")
			}

			injector.Set(typ, val)
			next.ServeHTTP(w, r)
		})
	}
}

// github.com/go-macaron/macaron/return_handler.go
func Handler(fn interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		injector, _ := r.Context().Value(injectorKey{}).(inject.Injector)
		if injector == nil {
			panic("chi: register Injector middleware")
		}
		vals, err := injector.Invoke(fn)
		if err != nil {
			panic(err)
		}

		var respVal reflect.Value
		if len(vals) > 1 && vals[0].Kind() == reflect.Int {
			w.WriteHeader(int(vals[0].Int()))
			respVal = vals[1]
		} else if len(vals) > 0 {
			respVal = vals[0]

			if isError(respVal) {
				err := respVal.Interface().(error)
				if err != nil {
					// ctx.internalServerError(ctx, err)
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(err.Error()))
				}
				return
			} else if canDeref(respVal) {
				if respVal.IsNil() {
					return // Ignore nil error
				}
			}
		}
		if canDeref(respVal) {
			respVal = respVal.Elem()
		}
		if isByteSlice(respVal) {
			_, _ = w.Write(respVal.Bytes())
		} else {
			_, _ = w.Write([]byte(respVal.String()))
		}
	}
}

func canDeref(val reflect.Value) bool {
	return val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr
}

func isError(val reflect.Value) bool {
	_, ok := val.Interface().(error)
	return ok
}

func isByteSlice(val reflect.Value) bool {
	return val.Kind() == reflect.Slice && val.Type().Elem().Kind() == reflect.Uint8
}
