package binding

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	httpw "go.wandrs.dev/http"
	"go.wandrs.dev/inject"

	"github.com/unrolled/render"
)

var pool = sync.Pool{
	New: func() interface{} {
		return inject.New()
	},
}

type injectorKey struct{}

func Injector(r *render.Render) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Check if a routing context already exists from a parent router.
			injector, _ := req.Context().Value(injectorKey{}).(inject.Injector)
			if injector != nil {
				next.ServeHTTP(w, req)
				return
			}

			injector = pool.Get().(inject.Injector)
			injector.Reset()

			// NOTE: req.WithContext() causes 2 allocations and context.WithValue() causes 1 allocation
			ctx := context.WithValue(req.Context(), injectorKey{}, injector)
			req = req.WithContext(ctx)

			injector.Map(ctx)
			injector.Map(req)
			injector.Map(w)
			injector.Map(httpw.NewResponseWriter(w, req, r))

			// Serve the request and once its done, put the request context back in the sync pool
			next.ServeHTTP(w, req)
			pool.Put(injector)
		})
	}
}

func Inject(fn func(inject.Injector)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			injector, _ := req.Context().Value(injectorKey{}).(inject.Injector)
			if injector == nil {
				panic("chi: register Injector middleware")
			}

			fn(injector)
			next.ServeHTTP(w, req)
		})
	}
}

// Maps the interface{} value based on its immediate type from reflect.TypeOf.
func Map(val interface{}) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			injector, _ := req.Context().Value(injectorKey{}).(inject.Injector)
			if injector == nil {
				panic("chi: register Injector middleware")
			}

			injector.Map(val)
			next.ServeHTTP(w, req)
		})
	}
}

// Maps the interface{} value based on the pointer of an Interface provided.
// This is really only useful for mapping a value as an interface, as interfaces
// cannot at this time be referenced directly without a pointer.
func MapTo(val interface{}, ifacePtr interface{}) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			injector, _ := req.Context().Value(injectorKey{}).(inject.Injector)
			if injector == nil {
				panic("chi: register Injector middleware")
			}

			injector.MapTo(val, ifacePtr)
			next.ServeHTTP(w, req)
		})
	}
}

// Provides a possibility to directly insert a mapping based on type and value.
// This makes it possible to directly map type arguments not possible to instantiate
// with reflect like unidirectional channels.
func Set(typ reflect.Type, val reflect.Value) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			injector, _ := req.Context().Value(injectorKey{}).(inject.Injector)
			if injector == nil {
				panic("chi: register Injector middleware")
			}

			injector.Set(typ, val)
			next.ServeHTTP(w, req)
		})
	}
}

var (
	errorType          = reflect.TypeOf((*error)(nil)).Elem()
	responseWriterType = reflect.TypeOf((*httpw.ResponseWriter)(nil)).Elem()
)

// github.com/go-macaron/macaron/return_handler.go
func HandlerFunc(fn interface{}) http.HandlerFunc {
	firstReturnIsErr := false

	typ := reflect.TypeOf(fn)
	if typ.Kind() != reflect.Func {
		panic(fmt.Sprintf("fn %s must be a function, found %s", typ, typ.Kind()))
	}
	switch typ.NumOut() {
	case 0:
		// nothing more to check
	case 1:
		etyp := typ.Out(0)
		if etyp.Implements(errorType) {
			firstReturnIsErr = true
		} else if reflect.New(etyp).Type().Implements(errorType) {
			panic(fmt.Sprintf("fn %s return type should be *%s to be considered an error", typ, etyp.Name()))
		}
	case 2:
		etyp := typ.Out(1)
		if !etyp.Implements(errorType) {
			if reflect.New(etyp).Type().Implements(errorType) {
				panic(fmt.Sprintf("fn %s 2nd return value should be *%s to be considered an error", typ, etyp.Name()))
			}
			panic("2nd return value must implement error")
		}
		vtyp := typ.Out(0)
		if vtyp.Implements(errorType) {
			panic(fmt.Sprintf("fn %s 1st return value must not an error", typ))
		}
	default:
		panic(fmt.Sprintf("fn %s has %d return values, at most 2 are allowed", typ, typ.NumOut()))
	}

	return func(w http.ResponseWriter, req *http.Request) {
		injector, _ := req.Context().Value(injectorKey{}).(inject.Injector)
		if injector == nil {
			panic("chi: register Injector middleware")
		}
		injector.Map(req.Context()) // make sure we have the latest Context

		results, err := injector.Invoke(fn)
		if err != nil {
			panic(fmt.Sprintf("failed to invoke %s, reason: %v", typ.String(), err))
		}

		ww := injector.GetVal(responseWriterType).Interface().(httpw.ResponseWriter)
		switch len(results) {
		case 0:
			return // nothing returned, assuming function directly wrote to http.ResponseWriter
		case 1:
			if firstReturnIsErr {
				err, _ := results[0].Interface().(error)
				ww.APIError(err)
				return
			}

			// ELSE,
			// write val[0] in JSON (use content negotiation in future)
			// nil object
			// nil slice
			// []byte
			// primitive types
			// objects
			// slices

			// isNil() ???

			v := results[0]
			if isByteSlice(v) {
				_, _ = w.Write(v.Bytes())
			} else {
				ww.JSON(http.StatusOK, v.Interface())
			}
			return
		case 2:
			err, _ := results[1].Interface().(error)
			if err != nil {
				ww.APIError(err)
				return
			}

			// if err == nil
			// write val[0] in JSON (use content negotiation in future)
			// nil object
			// nil slice
			// []byte
			// primitive types
			// objects
			// slices

			//if isByteSlice(respVal) {
			//	_, _ = w.Write(respVal.Bytes())
			//} else {

			v := results[0]
			if isByteSlice(v) {
				_, _ = w.Write(v.Bytes())
			} else {
				ww.JSON(http.StatusOK, v.Interface())
			}
			return
		default:
			panic(fmt.Sprintf("received %d return values, can only handle upto 2 return values", len(results)))
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
