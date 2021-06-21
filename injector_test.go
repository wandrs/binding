package binding_test

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.wandrs.dev/binding"
	httpw "go.wandrs.dev/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/unrolled/render"
)

var returnErr = errors.New("err")

type errorWrapper interface {
	error
}

func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func toJSONForVal(v interface{}, _ error) string {
	return toJSON(v)
}

func toJSONForErr(_ interface{}, err error) string {
	return toJSON(httpw.ErrorToAPIStatus(err))
}

func toStringForVal(v []byte, _ error) string {
	return string(v)
}

func h0_no_return(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

func h1_returns_error() error {
	return nil
}

func h1_returns_custom_error() *fs.PathError {
	return nil
}

func h1_returns_custom_error_ptr() fs.PathError {
	return fs.PathError{}
}

func h1_returns_custom_error_interface() errorWrapper {
	return nil
}

func h1_returns_string() string {
	return "handler"
}

func h1_returns_int() int {
	return 69
}

func h1_returns_bool() bool {
	return true
}

func h1_returns_byte_array() []byte {
	return []byte("handler")
}

func h1_returns_byte_array_empty() []byte {
	return []byte{}
}

func h1_returns_byte_array_nil() []byte {
	return nil
}

func h1_returns_string_array() []string {
	return []string{"handler"}
}

func h1_returns_string_array_empty() []string {
	return []string{}
}

func h1_returns_string_array_nil() []string {
	return nil
}

func h1_returns_int_array() []int {
	return []int{69, 420}
}

func h1_returns_int_array_empty() []int {
	return []int{}
}

func h1_returns_int_array_nil() []int {
	return nil
}

func h1_returns_bool_array() []bool {
	return []bool{true, false}
}

func h1_returns_bool_array_empty() []bool {
	return []bool{}
}

func h1_returns_bool_array_nil() []bool {
	return nil
}

func h1_returns_struct() Person {
	return Person{Name: "John"}
}

func h1_returns_struct_ptr() *Person {
	return &Person{Name: "John"}
}

func h1_returns_struct_ptr_nil() *Person {
	return nil
}

func h1_returns_array() []Person {
	return []Person{
		{Name: "John"},
		{Name: "Jane"},
	}
}

func h1_returns_array_empty() []Person {
	return []Person{}
}

func h1_returns_array_nil() []Person {
	return nil
}

func h2_returns_custom_error() (*Person, *fs.PathError) {
	return &Person{Name: "John"}, nil
}

func h2_returns_custom_error_err() (*Person, *fs.PathError) {
	return nil, &fs.PathError{Err: returnErr}
}

func h2_returns_custom_error_ptr() (Person, fs.PathError) {
	return Person{}, fs.PathError{}
}

func h2_returns_custom_error_interface() (*Person, errorWrapper) {
	return &Person{Name: "John"}, nil
}

func h2_returns_string() (string, error) {
	return "handler", nil
}

func h2_returns_string_err() (string, error) {
	return "", returnErr
}

func h2_returns_int() (int, error) {
	return 69, nil
}

func h2_returns_int_err() (int, error) {
	return -1, returnErr
}

func h2_returns_bool() (bool, error) {
	return true, nil
}

func h2_returns_bool_err() (bool, error) {
	return false, returnErr
}

func h2_returns_byte_array() ([]byte, error) {
	return []byte("handler"), nil
}

func h2_returns_byte_array_empty() ([]byte, error) {
	return []byte{}, nil
}

func h2_returns_byte_array_nil() ([]byte, error) {
	return nil, nil
}

func h2_returns_byte_array_err() ([]byte, error) {
	return nil, returnErr
}

func h2_returns_string_array() ([]string, error) {
	return []string{"handler"}, nil
}

func h2_returns_string_array_empty() ([]string, error) {
	return []string{}, nil
}

func h2_returns_string_array_nil() ([]string, error) {
	return nil, nil
}

func h2_returns_string_array_err() ([]string, error) {
	return nil, returnErr
}

func h2_returns_int_array() ([]int, error) {
	return []int{69, 420}, nil
}

func h2_returns_int_array_empty() ([]int, error) {
	return []int{}, nil
}

func h2_returns_int_array_nil() ([]int, error) {
	return nil, nil
}

func h2_returns_int_array_err() ([]int, error) {
	return nil, returnErr
}

func h2_returns_bool_array() ([]bool, error) {
	return []bool{true, false}, nil
}

func h2_returns_bool_array_empty() ([]bool, error) {
	return []bool{}, nil
}

func h2_returns_bool_array_nil() ([]bool, error) {
	return nil, nil
}

func h2_returns_bool_array_err() ([]bool, error) {
	return nil, returnErr
}

func h2_returns_struct() (Person, error) {
	return Person{Name: "John"}, nil
}

func h2_returns_struct_err() (Person, error) {
	return Person{}, returnErr
}

func h2_returns_struct_ptr() (*Person, error) {
	return &Person{Name: "John"}, nil
}

func h2_returns_struct_ptr_nil() (*Person, error) {
	return nil, nil
}

func h2_returns_struct_ptr_err() (*Person, error) {
	return nil, returnErr
}

func h2_returns_array() ([]Person, error) {
	return []Person{
		{Name: "John"},
		{Name: "Jane"},
	}, nil
}

func h2_returns_array_empty() ([]Person, error) {
	return []Person{}, nil
}

func h2_returns_array_nil() ([]Person, error) {
	return nil, nil
}

func h2_returns_array_err() ([]Person, error) {
	return nil, returnErr
}

func h_returns_too_many_returns() (int, Person, error) {
	return http.StatusOK, Person{Name: "John"}, nil
}

func TestHandlerFunc(t *testing.T) {
	tests := []struct {
		name   string
		args   interface{}
		want   string
		panics bool
	}{
		{
			name:   "h0_no_return",
			args:   h0_no_return,
			want:   "",
			panics: false,
		},
		{
			name:   "h1_returns_error",
			args:   h1_returns_error,
			want:   toJSON(httpw.ErrorToAPIStatus(h1_returns_error())),
			panics: false,
		},
		{
			name:   "h1_returns_custom_error",
			args:   h1_returns_custom_error,
			want:   toJSON(httpw.ErrorToAPIStatus(h1_returns_custom_error())),
			panics: false,
		},
		{
			name:   "h1_returns_custom_error_ptr",
			args:   h1_returns_custom_error_ptr,
			want:   "",
			panics: true,
		},
		{
			name:   "h1_returns_custom_error_interface",
			args:   h1_returns_custom_error_interface,
			want:   toJSON(httpw.ErrorToAPIStatus(h1_returns_custom_error_interface())),
			panics: false,
		},
		{
			name:   "h1_returns_string",
			args:   h1_returns_string,
			want:   toJSON(h1_returns_string()),
			panics: false,
		},
		{
			name:   "h1_returns_int",
			args:   h1_returns_int,
			want:   toJSON(h1_returns_int()),
			panics: false,
		},
		{
			name:   "h1_returns_bool",
			args:   h1_returns_bool,
			want:   toJSON(h1_returns_bool()),
			panics: false,
		},
		{
			name:   "h1_returns_byte_array",
			args:   h1_returns_byte_array,
			want:   string(h1_returns_byte_array()),
			panics: false,
		},
		{
			name:   "h1_returns_byte_array_empty",
			args:   h1_returns_byte_array_empty,
			want:   string(h1_returns_byte_array_empty()),
			panics: false,
		},
		{
			name:   "h1_returns_byte_array_nil",
			args:   h1_returns_byte_array_nil,
			want:   string(h1_returns_byte_array_nil()),
			panics: false,
		},

		{
			name:   "h1_returns_string_array",
			args:   h1_returns_string_array,
			want:   toJSON(h1_returns_string_array()),
			panics: false,
		},
		{
			name:   "h1_returns_string_array_empty",
			args:   h1_returns_string_array_empty,
			want:   toJSON(h1_returns_string_array_empty()),
			panics: false,
		},
		{
			name:   "h1_returns_string_array_nil",
			args:   h1_returns_string_array_nil,
			want:   toJSON(h1_returns_string_array_nil()),
			panics: false,
		},
		{
			name:   "h1_returns_int_array",
			args:   h1_returns_int_array,
			want:   toJSON(h1_returns_int_array()),
			panics: false,
		},
		{
			name:   "h1_returns_int_array_empty",
			args:   h1_returns_int_array_empty,
			want:   toJSON(h1_returns_int_array_empty()),
			panics: false,
		},
		{
			name:   "h1_returns_int_array_nil",
			args:   h1_returns_int_array_nil,
			want:   toJSON(h1_returns_int_array_nil()),
			panics: false,
		},
		{
			name:   "h1_returns_bool_array",
			args:   h1_returns_bool_array,
			want:   toJSON(h1_returns_bool_array()),
			panics: false,
		},
		{
			name:   "h1_returns_bool_array_empty",
			args:   h1_returns_bool_array_empty,
			want:   toJSON(h1_returns_bool_array_empty()),
			panics: false,
		},
		{
			name:   "h1_returns_bool_array_nil",
			args:   h1_returns_bool_array_nil,
			want:   toJSON(h1_returns_bool_array_nil()),
			panics: false,
		},
		{
			name:   "h1_returns_struct",
			args:   h1_returns_struct,
			want:   toJSON(h1_returns_struct()),
			panics: false,
		},
		{
			name:   "h1_returns_struct_ptr",
			args:   h1_returns_struct_ptr,
			want:   toJSON(h1_returns_struct_ptr()),
			panics: false,
		},
		{
			name:   "h1_returns_struct_ptr_nil",
			args:   h1_returns_struct_ptr_nil,
			want:   toJSON(h1_returns_struct_ptr_nil()),
			panics: false,
		},
		{
			name:   "h1_returns_array",
			args:   h1_returns_array,
			want:   toJSON(h1_returns_array()),
			panics: false,
		},
		{
			name:   "h1_returns_array_empty",
			args:   h1_returns_array_empty,
			want:   toJSON(h1_returns_array_empty()),
			panics: false,
		},
		{
			name:   "h1_returns_array_nil",
			args:   h1_returns_array_nil,
			want:   toJSON(h1_returns_array_nil()),
			panics: false,
		},

		{
			name:   "h2_returns_custom_error",
			args:   h2_returns_custom_error,
			want:   toJSONForVal(h2_returns_custom_error()),
			panics: false,
		},
		{
			name:   "h2_returns_custom_error_err",
			args:   h2_returns_custom_error_err,
			want:   toJSONForErr(h2_returns_custom_error_err()),
			panics: false,
		},
		{
			name:   "h2_returns_custom_error_ptr",
			args:   h2_returns_custom_error_ptr,
			want:   "",
			panics: true,
		},
		{
			name:   "h2_returns_custom_error_interface",
			args:   h2_returns_custom_error_interface,
			want:   toJSONForVal(h2_returns_custom_error_interface()),
			panics: false,
		},
		{
			name:   "h2_returns_string",
			args:   h2_returns_string,
			want:   toJSONForVal(h2_returns_string()),
			panics: false,
		},
		{
			name:   "h2_returns_string_err",
			args:   h2_returns_string_err,
			want:   toJSONForErr(h2_returns_string_err()),
			panics: false,
		},
		{
			name:   "h2_returns_int",
			args:   h2_returns_int,
			want:   toJSONForVal(h2_returns_int()),
			panics: false,
		},
		{
			name:   "h2_returns_int_err",
			args:   h2_returns_int_err,
			want:   toJSONForErr(h2_returns_int_err()),
			panics: false,
		},
		{
			name:   "h2_returns_bool",
			args:   h2_returns_bool,
			want:   toJSONForVal(h2_returns_bool()),
			panics: false,
		},
		{
			name:   "h2_returns_bool_err",
			args:   h2_returns_bool_err,
			want:   toJSONForErr(h2_returns_bool_err()),
			panics: false,
		},
		{
			name:   "h2_returns_byte_array",
			args:   h2_returns_byte_array,
			want:   toStringForVal(h2_returns_byte_array()),
			panics: false,
		},
		{
			name:   "h2_returns_byte_array_empty",
			args:   h2_returns_byte_array_empty,
			want:   toStringForVal(h2_returns_byte_array_empty()),
			panics: false,
		},
		{
			name:   "h2_returns_byte_array_nil",
			args:   h2_returns_byte_array_nil,
			want:   toStringForVal(h2_returns_byte_array_nil()),
			panics: false,
		},
		{
			name:   "h2_returns_byte_array_err",
			args:   h2_returns_byte_array_err,
			want:   toJSONForErr(h2_returns_byte_array_err()),
			panics: false,
		},
		{
			name:   "h2_returns_string_array",
			args:   h2_returns_string_array,
			want:   toJSONForVal(h2_returns_string_array()),
			panics: false,
		},
		{
			name:   "h2_returns_string_array_empty",
			args:   h2_returns_string_array_empty,
			want:   toJSONForVal(h2_returns_string_array_empty()),
			panics: false,
		},
		{
			name:   "h2_returns_string_array_nil",
			args:   h2_returns_string_array_nil,
			want:   toJSONForVal(h2_returns_string_array_nil()),
			panics: false,
		},
		{
			name:   "h2_returns_string_array_err",
			args:   h2_returns_string_array_err,
			want:   toJSONForErr(h2_returns_string_array_err()),
			panics: false,
		},
		{
			name:   "h2_returns_int_array",
			args:   h2_returns_int_array,
			want:   toJSONForVal(h2_returns_int_array()),
			panics: false,
		},
		{
			name:   "h2_returns_int_array_empty",
			args:   h2_returns_int_array_empty,
			want:   toJSONForVal(h2_returns_int_array_empty()),
			panics: false,
		},
		{
			name:   "h2_returns_int_array_nil",
			args:   h2_returns_int_array_nil,
			want:   toJSONForVal(h2_returns_int_array_nil()),
			panics: false,
		},
		{
			name:   "h2_returns_int_array_err",
			args:   h2_returns_int_array_err,
			want:   toJSONForErr(h2_returns_int_array_err()),
			panics: false,
		},
		{
			name:   "h2_returns_bool_array",
			args:   h2_returns_bool_array,
			want:   toJSONForVal(h2_returns_bool_array()),
			panics: false,
		},
		{
			name:   "h2_returns_bool_array_empty",
			args:   h2_returns_bool_array_empty,
			want:   toJSONForVal(h2_returns_bool_array_empty()),
			panics: false,
		},
		{
			name:   "h2_returns_bool_array_nil",
			args:   h2_returns_bool_array_nil,
			want:   toJSONForVal(h2_returns_bool_array_nil()),
			panics: false,
		},
		{
			name:   "h2_returns_bool_array_err",
			args:   h2_returns_bool_array_err,
			want:   toJSONForErr(h2_returns_bool_array_err()),
			panics: false,
		},
		{
			name:   "h2_returns_struct",
			args:   h2_returns_struct,
			want:   toJSONForVal(h2_returns_struct()),
			panics: false,
		},
		{
			name:   "h2_returns_struct_err",
			args:   h2_returns_struct_err,
			want:   toJSONForErr(h2_returns_struct_err()),
			panics: false,
		},
		{
			name:   "h2_returns_struct_ptr",
			args:   h2_returns_struct_ptr,
			want:   toJSONForVal(h2_returns_struct_ptr()),
			panics: false,
		},
		{
			name:   "h2_returns_struct_ptr_nil",
			args:   h2_returns_struct_ptr_nil,
			want:   toJSONForVal(h2_returns_struct_ptr_nil()),
			panics: false,
		},
		{
			name:   "h2_returns_struct_ptr_err",
			args:   h2_returns_struct_ptr_err,
			want:   toJSONForErr(h2_returns_struct_ptr_err()),
			panics: false,
		},
		{
			name:   "h2_returns_array",
			args:   h2_returns_array,
			want:   toJSONForVal(h2_returns_array()),
			panics: false,
		},
		{
			name:   "h2_returns_array_empty",
			args:   h2_returns_array_empty,
			want:   toJSONForVal(h2_returns_array_empty()),
			panics: false,
		},
		{
			name:   "h2_returns_array_nil",
			args:   h2_returns_array_nil,
			want:   toJSONForVal(h2_returns_array_nil()),
			panics: false,
		},
		{
			name:   "h2_returns_array_err",
			args:   h2_returns_array_err,
			want:   toJSONForErr(h2_returns_array_err()),
			panics: false,
		},

		{
			name:   "h_returns_too_many_returns",
			args:   h_returns_too_many_returns,
			want:   "",
			panics: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			func() {
				defer func() {
					if r := recover(); r != nil && !tt.panics {
						t.Errorf("unexpected panic %v", r)
					}
				}()

				w := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/", nil)

				m := chi.NewRouter()
				m.Use(middleware.Logger)
				m.Use(binding.Injector(render.New()))
				m.Get("/", binding.HandlerFunc(tt.args))

				m.ServeHTTP(w, req)

				resp := w.Result()
				if got, _ := io.ReadAll(resp.Body); string(got) != tt.want {
					t.Errorf("body = %v, want %v", string(got), tt.want)
				}
			}()
		})
	}
}
