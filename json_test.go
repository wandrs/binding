// Copyright 2014 Martini Authors
// Copyright 2014 The Macaron Authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package binding_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/unrolled/render"
	"go.wandrs.dev/binding"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

const (
	_JSON_CONTENT_TYPE = "application/json; charset=utf-8"
)

type (
	jsonTestCase struct {
		description        string
		withInterface      bool
		shouldFailOnBind   bool
		payload            string
		contentType        string
		expected           interface{}
		expectedStatusCode int
	}
)

var jsonTestCases = []jsonTestCase{
	{
		description:        "Happy path",
		expectedStatusCode: http.StatusOK,
		payload:            `{"title": "Glorious Post Title", "content": "Lorem ipsum dolor sit amet"}`,
		contentType:        _JSON_CONTENT_TYPE,
		expected:           Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:        "Happy path with interface",
		expectedStatusCode: http.StatusOK,
		withInterface:      true,
		payload:            `{"title": "Glorious Post Title", "content": "Lorem ipsum dolor sit amet"}`,
		contentType:        _JSON_CONTENT_TYPE,
		expected:           Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:        "Nil payload",
		expectedStatusCode: http.StatusUnprocessableEntity,
		payload:            `-nil-`,
		contentType:        _JSON_CONTENT_TYPE,
		expected:           Post{},
	},
	{
		description:        "Empty payload",
		expectedStatusCode: http.StatusUnprocessableEntity,
		payload:            ``,
		contentType:        _JSON_CONTENT_TYPE,
		expected:           Post{},
	},
	{
		description:        "Empty content type",
		expectedStatusCode: http.StatusOK,
		shouldFailOnBind:   true,
		payload:            `{"title": "Glorious Post Title", "content": "Lorem ipsum dolor sit amet"}`,
		contentType:        ``,
		expected:           Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:        "Unsupported content type",
		expectedStatusCode: http.StatusOK,
		shouldFailOnBind:   true,
		payload:            `{"title": "Glorious Post Title", "content": "Lorem ipsum dolor sit amet"}`,
		contentType:        `BoGuS`,
		expected:           Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:        "Malformed JSON",
		expectedStatusCode: http.StatusBadRequest,
		payload:            `{"title":"foo"`,
		contentType:        _JSON_CONTENT_TYPE,
		expected:           Post{Title: "foo"},
	},
	{
		description:        "Deserialization with nested and embedded struct",
		expectedStatusCode: http.StatusOK,
		payload:            `{"title":"Glorious Post Title", "id":1, "author":{"name":"Matt Holt"}}`,
		contentType:        _JSON_CONTENT_TYPE,
		expected:           BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:        "Deserialization with nested and embedded struct with interface",
		expectedStatusCode: http.StatusOK,
		withInterface:      true,
		payload:            `{"title":"Glorious Post Title", "id":1, "author":{"name":"Matt Holt"}}`,
		contentType:        _JSON_CONTENT_TYPE,
		expected:           BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:        "Required nested struct field not specified",
		expectedStatusCode: http.StatusUnprocessableEntity,
		payload:            `{"title":"Glorious Post Title", "id":1, "author":{}}`,
		contentType:        _JSON_CONTENT_TYPE,
		expected:           BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1},
	},
	{
		description:        "Required embedded struct field not specified",
		expectedStatusCode: http.StatusUnprocessableEntity,
		payload:            `{"id":1, "author":{"name":"Matt Holt"}}`,
		contentType:        _JSON_CONTENT_TYPE,
		expected:           BlogPost{Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:        "Slice of Posts",
		expectedStatusCode: http.StatusOK,
		payload:            `[{"title": "First Post"}, {"title": "Second Post"}]`,
		contentType:        _JSON_CONTENT_TYPE,
		expected:           []Post{{Title: "First Post"}, {Title: "Second Post"}},
	},
	{
		description:        "Slice of structs",
		expectedStatusCode: http.StatusOK,
		payload:            `{"name": "group1", "people": [{"name":"awoods"}, {"name": "anthony"}]}`,
		contentType:        _JSON_CONTENT_TYPE,
		expected:           Group{Name: "group1", People: []Person{{Name: "awoods"}, {Name: "anthony"}}},
	},
}

func Test_JSON(t *testing.T) {
	for _, testCase := range jsonTestCases {
		performJSONTest(t, binding.JSON, testCase)
	}
}

func performJSONTest(t *testing.T, binder binderFunc, testCase jsonTestCase) {
	t.Run(testCase.description, func(t *testing.T) {
		m := chi.NewRouter()
		m.Use(middleware.Logger)
		m.Use(binding.Injector(render.New()))

		switch testCase.expected.(type) {
		case []Post:
			if testCase.withInterface {
				m.With(binder([]Post{}, (*modeler)(nil))).
					Post(testRoute, binding.HandlerFunc(func(w http.ResponseWriter, actual []Post, iface modeler) {
						assert.Equal(t, testCase.expected, actual)
						for _, a := range actual {
							assert.Equal(t, a.Title, iface.Model())
						}
						w.WriteHeader(http.StatusOK)
					}))
			} else {
				m.With(binder([]Post{})).
					Post(testRoute, binding.HandlerFunc(func(w http.ResponseWriter, actual []Post) {
						assert.Equal(t, testCase.expected, actual)
						w.WriteHeader(http.StatusOK)
					}))
			}

		case Post:
			if testCase.withInterface {
				m.With(binder(Post{}, (*modeler)(nil))).
					Post(testRoute, binding.HandlerFunc(func(w http.ResponseWriter, actual Post, iface modeler) {
						assert.Equal(t, testCase.expected, actual)
						assert.Equal(t, actual.Title, iface.Model())
						w.WriteHeader(http.StatusOK)
					}))
			} else {
				m.With(binder(Post{})).
					Post(testRoute, binding.HandlerFunc(func(w http.ResponseWriter, actual Post) {
						assert.Equal(t, testCase.expected, actual)
						w.WriteHeader(http.StatusOK)
					}))
			}

		case BlogPost:
			if testCase.withInterface {
				m.With(binder(BlogPost{}, (*modeler)(nil))).
					Post(testRoute, binding.HandlerFunc(func(w http.ResponseWriter, actual BlogPost, iface modeler) {
						assert.Equal(t, testCase.expected, actual)
						assert.Equal(t, actual.Title, iface.Model())
						w.WriteHeader(http.StatusOK)
					}))
			} else {
				m.With(binder(BlogPost{}, (*modeler)(nil))).
					Post(testRoute, binding.HandlerFunc(func(w http.ResponseWriter, actual BlogPost) {
						assert.Equal(t, testCase.expected, actual)
						w.WriteHeader(http.StatusOK)
					}))
			}

		case Group:
			if testCase.withInterface {
				m.With(binder(Group{}, (*modeler)(nil))).
					Post(testRoute, binding.HandlerFunc(func(w http.ResponseWriter, actual Group, iface modeler) {
						assert.Equal(t, testCase.expected, actual)
						assert.Equal(t, actual.Name, iface.Model())
						w.WriteHeader(http.StatusOK)
					}))
			} else {
				m.With(binder(Group{}, (*modeler)(nil))).
					Post(testRoute, binding.HandlerFunc(func(w http.ResponseWriter, actual Group) {
						assert.Equal(t, testCase.expected, actual)
						w.WriteHeader(http.StatusOK)
					}))
			}
		}

		var payload io.Reader
		if testCase.payload == "-nil-" {
			payload = nil
		} else {
			payload = strings.NewReader(testCase.payload)
		}

		req, err := http.NewRequest(http.MethodPost, testRoute, payload)
		if err != nil {
			panic(err)
		}
		req.Header.Set("Content-Type", testCase.contentType)

		w := httptest.NewRecorder()
		m.ServeHTTP(w, req)
		resp := w.Result()

		switch resp.StatusCode {
		case http.StatusNotFound:
			panic("Routing is messed up in test fixture (got 404): check method and path")
		case http.StatusInternalServerError:
			panic("Something bad happened on '" + testCase.description + "'")
		case http.StatusUnsupportedMediaType:
			if !testCase.shouldFailOnBind {
				panic("expected to fail '" + testCase.description + "'")
			}
		default:
			assert.EqualValues(t, testCase.expectedStatusCode, resp.StatusCode)
		}
	})
}
