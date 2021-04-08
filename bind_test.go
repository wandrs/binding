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
	"testing"
)

func Test_Bind(t *testing.T) {
	t.Run("Bind form", func(t *testing.T) {
		for _, testCase := range formTestCases {
			performFormTest(t, Bind, testCase)
		}
	})

	t.Run("Bind JSON", func(t *testing.T) {
		for _, testCase := range jsonTestCases {
			performJsonTest(t, Bind, testCase)
		}
	})

	t.Run("Bind multipart form", func(t *testing.T) {
		for _, testCase := range multipartFormTestCases {
			performMultipartFormTest(t, Bind, testCase)
		}
	})

	t.Run("Bind with file", func(t *testing.T) {
		for _, testCase := range fileTestCases {
			performFileTest(t, Bind, testCase)
		}
	})
}
