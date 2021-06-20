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

package binding

import (
	"fmt"
	"github.com/unrolled/render"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

var formTestCases = []formTestCase{
	{
		description:   "Happy path",
		shouldSucceed: true,
		payload:       `title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet`,
		contentType:   formContentType,
		expected:      Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:   "Happy path with interface",
		shouldSucceed: true,
		withInterface: true,
		payload:       `title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet`,
		contentType:   formContentType,
		expected:      Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:   "Empty payload",
		shouldSucceed: false,
		payload:       ``,
		contentType:   formContentType,
		expected:      Post{},
	},
	{
		description:   "Empty content type",
		shouldSucceed: false,
		payload:       `title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet`,
		contentType:   ``,
		expected:      Post{},
	},
	{
		description:   "Malformed form body",
		shouldSucceed: false,
		payload:       `title=%2`,
		contentType:   formContentType,
		expected:      Post{},
	},
	{
		description:   "With nested and embedded structs",
		shouldSucceed: true,
		payload:       `title=Glorious+Post+Title&id=1&name=Matt+Holt`,
		contentType:   formContentType,
		expected:      BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:   "Required embedded struct field not specified",
		shouldSucceed: false,
		payload:       `id=1&name=Matt+Holt`,
		contentType:   formContentType,
		expected:      BlogPost{Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:   "Required nested struct field not specified",
		shouldSucceed: false,
		payload:       `title=Glorious+Post+Title&id=1`,
		contentType:   formContentType,
		expected:      BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1},
	},
	{
		description:   "Multiple values into slice",
		shouldSucceed: true,
		payload:       `title=Glorious+Post+Title&id=1&name=Matt+Holt&rating=4&rating=3&rating=5`,
		contentType:   formContentType,
		expected:      BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}, Ratings: []int{4, 3, 5}},
	},
	{
		description:   "Unexported field",
		shouldSucceed: true,
		payload:       `title=Glorious+Post+Title&id=1&name=Matt+Holt&unexported=foo`,
		contentType:   formContentType,
		expected:      BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:   "Query string POST",
		shouldSucceed: true,
		payload:       `title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet`,
		contentType:   formContentType,
		expected:      Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:   "Query string with Content-Type (POST request)",
		shouldSucceed: true,
		queryString:   "?title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet",
		payload:       ``,
		contentType:   formContentType,
		expected:      Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:   "Query string without Content-Type (GET request)",
		shouldSucceed: true,
		method:        "GET",
		queryString:   "?title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet",
		payload:       ``,
		expected:      Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:   "Embed struct pointer",
		shouldSucceed: true,
		deepEqual:     true,
		method:        "GET",
		queryString:   "?name=Glorious+Post+Title&email=Lorem+ipsum+dolor+sit+amet",
		payload:       ``,
		expected:      EmbedPerson{&Person{Name: "Glorious Post Title", Email: "Lorem ipsum dolor sit amet"}},
	},
	{
		description:   "Embed struct pointer remain nil if not binded",
		shouldSucceed: true,
		deepEqual:     true,
		method:        "GET",
		queryString:   "?",
		payload:       ``,
		expected:      EmbedPerson{nil},
	},
	{
		description:   "Custom error handler",
		shouldSucceed: true,
		deepEqual:     true,
		method:        "GET",
		queryString:   "?",
		payload:       ``,
		expected:      CustomErrorHandle{},
	},
}

func Test_Form(t *testing.T) {
	for _, testCase := range formTestCases {
		t.Run(testCase.description, func(t *testing.T) {
			performFormTest(t, Form, testCase)
		})
	}
}

func performFormTest(t *testing.T, binder handlerFunc, testCase formTestCase) {
	resp := httptest.NewRecorder()
	m := chi.NewRouter()
	m.Use(Injector(render.New()))

	formTestHandler := func(actual interface{}, errs Errors) {
		if testCase.shouldSucceed {
			assert.Empty(t, errs)
		} else if !testCase.shouldSucceed {
			assert.NotEmpty(t, errs)
		}
		expString := fmt.Sprintf("%+v", testCase.expected)
		actString := fmt.Sprintf("%+v", actual)
		if actString != expString && !(testCase.deepEqual && reflect.DeepEqual(testCase.expected, actual)) {
			assert.EqualValues(t, expString, actString)
		}
	}

	switch testCase.expected.(type) {
	case Post:
		if testCase.withInterface {
			m.Post(testRoute, func(resp http.ResponseWriter, req *http.Request) {
				var actual Post
				errs := binder(req, &actual)
				p := testCase.expected.(Post)
				assert.EqualValues(t, p.Title, actual.Title)
				formTestHandler(actual, errs)
			})
		} else {
			m.Post(testRoute, func(resp http.ResponseWriter, req *http.Request) {
				var actual Post
				errs := binder(req, &actual)
				formTestHandler(actual, errs)
			})
			m.Get(testRoute, func(resp http.ResponseWriter, req *http.Request) {
				var actual Post
				errs := binder(req, &actual)
				formTestHandler(actual, errs)
			})
		}

	case BlogPost:
		if testCase.withInterface {
			m.Post(testRoute, func(resp http.ResponseWriter, req *http.Request) {
				var actual BlogPost
				errs := binder(req, &actual)
				p := testCase.expected.(BlogPost)
				assert.EqualValues(t, p.Title, actual.Title)
				formTestHandler(actual, errs)
			})
		} else {
			m.Post(testRoute, func(resp http.ResponseWriter, req *http.Request) {
				var actual BlogPost
				errs := binder(req, &actual)
				formTestHandler(actual, errs)
			})
		}

	case EmbedPerson:
		m.Post(testRoute, func(resp http.ResponseWriter, req *http.Request) {
			var actual EmbedPerson
			errs := binder(req, &actual)
			formTestHandler(actual, errs)
		})
		m.Get(testRoute, func(resp http.ResponseWriter, req *http.Request) {
			var actual EmbedPerson
			errs := binder(req, &actual)
			formTestHandler(actual, errs)
		})
	case CustomErrorHandle:
		m.Get(testRoute, func(resp http.ResponseWriter, req *http.Request) {
			var actual CustomErrorHandle
			errs := binder(req, &actual)
			formTestHandler(actual, errs)
		})
	}

	if len(testCase.method) == 0 {
		testCase.method = "POST"
	}

	req, err := http.NewRequest(testCase.method, testRoute+testCase.queryString, strings.NewReader(testCase.payload))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", testCase.contentType)

	m.ServeHTTP(resp, req)

	switch resp.Code {
	case http.StatusNotFound:
		panic("Routing is messed up in test fixture (got 404): check methods and paths")
	case http.StatusInternalServerError:
		panic("Something bad happened on '" + testCase.description + "'")
	}
}

type (
	formTestCase struct {
		description   string
		shouldSucceed bool
		deepEqual     bool
		withInterface bool
		queryString   string
		payload       string
		contentType   string
		expected      interface{}
		method        string
	}
)

type defaultForm struct {
	Default string `binding:"Default(hello world)"`
}

func Test_Default(t *testing.T) {
	m := chi.NewRouter()
	m.Use(Injector(render.New()))

	m.Get("/", func(resp http.ResponseWriter, req *http.Request) {
		var f defaultForm
		Bind(req, &f)
		assert.EqualValues(t, "hello world", f.Default)
	})
	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	assert.Nil(t, err)

	m.ServeHTTP(resp, req)
}
