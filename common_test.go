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
	"net/http"
)

const (
	testRoute       = "/test"
	formContentType = "application/x-www-form-urlencoded"
)

type binderFunc func(obj interface{}, ifacePtr ...interface{}) func(next http.Handler) http.Handler

// These types are mostly contrived examples, but they're used
// across many test cases. The idea is to cover all the scenarios
// that this binding package might encounter in actual use.
type (
	// For basic test cases with a required field
	Post struct {
		Title   string `json:"title" form:"title" validate:"required"`
		Content string `json:"content" form:"content"`
	}

	// To be used as a nested struct (with a required field)
	Person struct {
		Name  string `json:"name" form:"name" validate:"required"`
		Email string `json:"email,omitempty" form:"email"`
	}

	// For advanced test cases: multiple values, embedded
	// and nested structs, an ignored field, and single
	// and multiple file uploads
	BlogPost struct {
		Post
		Id         int     `form:"id" validate:"required"` // JSON not specified here for test coverage
		Ignored    string  `json:"-" form:"-"`
		Ratings    []int   `json:"ratings" form:"rating"`
		Author     Person  `json:"author" form:"author"`
		Coauthor   *Person `json:"coauthor"`
		unexported string  `form:"unexported"` //nolint
	}

	EmbedPerson struct {
		*Person
	}

	Group struct {
		Name   string   `json:"name" validate:"required"`
		People []Person `json:"people" validate:"min=1"`
	}

	// Used for testing mapping an interface to the context
	// If used (withInterface = true in the testCases), a modeler
	// should be mapped to the context as well as BlogPost, meaning
	// you can receive a modeler in your application instead of a
	// concrete BlogPost.
	modeler interface {
		Model() string
	}
)

func (p Post) Model() string {
	return p.Title
}

func (g Group) Model() string {
	return g.Name
}
