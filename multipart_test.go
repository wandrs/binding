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
	"bytes"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/unrolled/render"
	"go.wandrs.dev/binding"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

type (
	multipartFormTestCase struct {
		description         string
		expectedStatusCode  int
		expected            BlogPost
		malformEncoding     bool
		callFormValueBefore bool
	}
)

var multipartFormTestCases = []multipartFormTestCase{
	{
		description:        "Happy multipart form path",
		expectedStatusCode: http.StatusOK,
		expected:           BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:         "FormValue called before req.MultipartReader(); see https://github.com/martini-contrib/csrf/issues/6",
		expectedStatusCode:  http.StatusOK,
		callFormValueBefore: true,
		expected:            BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:        "Empty payload",
		expectedStatusCode: http.StatusUnprocessableEntity,
		expected:           BlogPost{},
	},
	{
		description:        "Missing required field (Id)",
		expectedStatusCode: http.StatusUnprocessableEntity,
		expected:           BlogPost{Post: Post{Title: "Glorious Post Title"}, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:        "Required embedded struct field not specified",
		expectedStatusCode: http.StatusUnprocessableEntity,
		expected:           BlogPost{Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:        "Required nested struct field not specified",
		expectedStatusCode: http.StatusUnprocessableEntity,
		expected:           BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1},
	},
	{
		description:        "Multiple values",
		expectedStatusCode: http.StatusOK,
		expected:           BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}, Ratings: []int{3, 5, 4}},
	},
	//{
	//	description:        "Bad multipart encoding",
	//	expectedStatusCode: http.StatusBadRequest,
	//	malformEncoding:    true,
	//},
}

func Test_MultipartForm(t *testing.T) {
	for _, testCase := range multipartFormTestCases {
		performMultipartFormTest(t, binding.MultipartForm, testCase)
	}
}

func performMultipartFormTest(t *testing.T, binder binderFunc, testCase multipartFormTestCase) {
	m := chi.NewRouter()
	m.Use(middleware.Logger)
	m.Use(binding.Injector(render.New()))

	m.With(binder(BlogPost{})).
		Post(testRoute, binding.HandlerFunc(func(w http.ResponseWriter, actual BlogPost) {
			assert.Equal(t, testCase.expected, actual)
			w.WriteHeader(http.StatusOK)
		}))

	multipartPayload, mpWriter := makeMultipartPayload(testCase)

	req, err := http.NewRequest("POST", testRoute, multipartPayload)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-Type", mpWriter.FormDataContentType())

	err = mpWriter.Close()
	if err != nil {
		panic(err)
	}

	if testCase.callFormValueBefore {
		req.FormValue("foo")
	}

	w := httptest.NewRecorder()
	m.ServeHTTP(w, req)
	resp := w.Result()

	switch resp.StatusCode {
	case http.StatusNotFound:
		panic("Routing is messed up in test fixture (got 404): check methods and paths")
	case http.StatusInternalServerError:
		panic("Something bad happened on '" + testCase.description + "'")
	default:
		assert.EqualValues(t, testCase.expectedStatusCode, resp.StatusCode)
	}
}

// Writes the input from a test case into a buffer using the multipart writer.
func makeMultipartPayload(testCase multipartFormTestCase) (*bytes.Buffer, *multipart.Writer) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if testCase.malformEncoding {
		// TODO: Break the multipart form parser which is apparently impervious!!
		// (Get it to return an error. Trying to get 100% test coverage.)
		body.Write([]byte(`--` + writer.Boundary() + `\nContent-Disposition: form-data; name="foo"\n\n--` + writer.Boundary() + `--`))
		return body, writer
	} else {
		writer.WriteField("title", testCase.expected.Title)
		writer.WriteField("content", testCase.expected.Content)
		writer.WriteField("id", strconv.Itoa(testCase.expected.Id))
		writer.WriteField("ignored", testCase.expected.Ignored)
		for _, value := range testCase.expected.Ratings {
			writer.WriteField("rating", strconv.Itoa(value))
		}
		writer.WriteField("author.name", testCase.expected.Author.Name)
		writer.WriteField("author.email", testCase.expected.Author.Email)
		return body, writer
	}
}
