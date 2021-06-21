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
	"github.com/go-chi/chi/v5/middleware"
	"github.com/unrolled/render"
	"go.wandrs.dev/binding"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type (
	formTestCase struct {
		description        string
		expectedStatusCode int
		shouldFailOnBind   bool
		withInterface      bool
		queryString        string
		payload            string
		contentType        string
		expected           interface{}
		method             string
	}
)

var formTestCases = []formTestCase{
	{
		description:        "Happy path",
		expectedStatusCode: http.StatusOK,
		method:             "POST",
		payload:            `title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet`,
		contentType:        formContentType,
		expected:           Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:        "Happy path with interface",
		expectedStatusCode: http.StatusOK,
		withInterface:      true,
		method:             "POST",
		payload:            `title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet`,
		contentType:        formContentType,
		expected:           Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:        "Empty payload",
		expectedStatusCode: http.StatusUnprocessableEntity,
		method:             "POST",
		payload:            ``,
		contentType:        formContentType,
		expected:           Post{},
	},
	{
		description:        "Empty content type",
		expectedStatusCode: http.StatusUnprocessableEntity,
		shouldFailOnBind:   true,
		method:             "POST",
		payload:            `title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet`,
		contentType:        ``,
		expected:           Post{},
	},
	{
		description:        "Malformed form body",
		expectedStatusCode: http.StatusBadRequest,
		method:             "POST",
		payload:            `title=%2`,
		contentType:        formContentType,
		expected:           Post{},
	},
	{
		description:        "With nested and embedded structs",
		expectedStatusCode: http.StatusOK,
		method:             "POST",
		payload:            `title=Glorious+Post+Title&id=1&author.name=Matt+Holt`,
		contentType:        formContentType,
		expected:           BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:        "Required embedded struct field not specified",
		expectedStatusCode: http.StatusUnprocessableEntity,
		method:             "POST",
		payload:            `id=1&author.name=Matt+Holt`,
		contentType:        formContentType,
		expected:           BlogPost{Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:        "Required nested struct field not specified",
		expectedStatusCode: http.StatusUnprocessableEntity,
		method:             "POST",
		payload:            `title=Glorious+Post+Title&id=1`,
		contentType:        formContentType,
		expected:           BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1},
	},
	{
		description:        "Multiple values into slice",
		expectedStatusCode: http.StatusOK,
		method:             "POST",
		payload:            `title=Glorious+Post+Title&id=1&author.name=Matt+Holt&rating=4&rating=3&rating=5`,
		contentType:        formContentType,
		expected:           BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}, Ratings: []int{4, 3, 5}},
	},
	{
		description:        "Unexported field",
		expectedStatusCode: http.StatusOK,
		method:             "POST",
		payload:            `title=Glorious+Post+Title&id=1&author.name=Matt+Holt&unexported=foo`,
		contentType:        formContentType,
		expected:           BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:        "Query string POST",
		expectedStatusCode: http.StatusOK,
		method:             "POST",
		payload:            `title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet`,
		contentType:        formContentType,
		expected:           Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:        "Query string with Content-Type (POST request)",
		expectedStatusCode: http.StatusOK,
		method:             "POST",
		queryString:        "?title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet",
		payload:            ``,
		contentType:        formContentType,
		expected:           Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:        "Query string without Content-Type (GET request)",
		expectedStatusCode: http.StatusOK,
		method:             "GET",
		queryString:        "?title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet",
		payload:            ``,
		expected:           Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:        "Embed struct pointer",
		expectedStatusCode: http.StatusOK,
		method:             "GET",
		queryString:        "?name=Glorious+Post+Title&email=Lorem+ipsum+dolor+sit+amet",
		payload:            ``,
		expected:           EmbedPerson{&Person{Name: "Glorious Post Title", Email: "Lorem ipsum dolor sit amet"}},
	},
	{
		description:        "Embed struct pointer remain nil if not binded",
		expectedStatusCode: http.StatusOK,
		method:             "GET",
		queryString:        "?",
		payload:            ``,
		expected:           EmbedPerson{nil},
	},
}

func Test_Form(t *testing.T) {
	for _, testCase := range formTestCases {
		t.Run(testCase.description, func(t *testing.T) {
			performFormTest(t, binding.Form, testCase)
		})
	}
}

func performFormTest(t *testing.T, binder binderFunc, testCase formTestCase) {
	m := chi.NewRouter()
	m.Use(middleware.Logger)
	m.Use(binding.Injector(render.New()))

	switch testCase.expected.(type) {
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
			m.With(binder(Post{})).
				Get(testRoute, binding.HandlerFunc(func(w http.ResponseWriter, actual Post) {
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
			m.With(binder(BlogPost{})).
				Post(testRoute, binding.HandlerFunc(func(w http.ResponseWriter, actual BlogPost) {
					assert.Equal(t, testCase.expected, actual)
					w.WriteHeader(http.StatusOK)
				}))
		}

	case EmbedPerson:
		m.With(binder(EmbedPerson{})).
			Post(testRoute, binding.HandlerFunc(func(w http.ResponseWriter, actual EmbedPerson) {
				assert.Equal(t, testCase.expected, actual)
				w.WriteHeader(http.StatusOK)
			}))
		m.With(binder(EmbedPerson{})).
			Get(testRoute, binding.HandlerFunc(func(w http.ResponseWriter, actual EmbedPerson) {
				assert.Equal(t, testCase.expected, actual)
				w.WriteHeader(http.StatusOK)
			}))
	}

	req, err := http.NewRequest(testCase.method, testRoute+testCase.queryString, strings.NewReader(testCase.payload))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", testCase.contentType)

	w := httptest.NewRecorder()
	m.ServeHTTP(w, req)
	resp := w.Result()

	switch resp.StatusCode {
	case http.StatusNotFound:
		panic("Routing is messed up in test fixture (got 404): check methods and paths")
	case http.StatusInternalServerError:
		panic("Something bad happened on '" + testCase.description + "'")
	case http.StatusUnsupportedMediaType:
		if !testCase.shouldFailOnBind {
			panic("expected to fail '" + testCase.description + "'")
		}
	default:
		assert.EqualValues(t, testCase.expectedStatusCode, resp.StatusCode)
	}
}
