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
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ErrorsAdd(t *testing.T) {
	var actual Errors
	expected := Errors{
		Error{
			FieldNames:     []string{"Field1", "Field2"},
			Classification: "ErrorClass",
			Message:        "Some message",
		},
	}

	actual.Add(expected[0].FieldNames, expected[0].Classification, expected[0].Message)

	assert.Len(t, actual, 1)
	assert.EqualValues(t, fmt.Sprintf("%#v", expected), fmt.Sprintf("%#v", actual))
}

func Test_ErrorsLen(t *testing.T) {
	assert.EqualValues(t, len(errorsTestSet), errorsTestSet.Len())
}

func Test_ErrorsHas(t *testing.T) {
	assert.True(t, errorsTestSet.Has("ClassA"))
	assert.False(t, errorsTestSet.Has("ClassQ"))
}

func Test_ErrorGetters(t *testing.T) {

	err := Error{
		FieldNames:     []string{"field1", "field2"},
		Classification: "ErrorClass",
		Message:        "The message",
	}

	fieldsActual := err.Fields()

	assert.Len(t, fieldsActual, 2)
	assert.EqualValues(t, "field1", fieldsActual[0])
	assert.EqualValues(t, "field2", fieldsActual[1])

	assert.EqualValues(t, "ErrorClass", err.Kind())
	assert.EqualValues(t, "The message", err.Error())

}

/*
func TestErrorsWithClass(t *testing.T) {
	expected := Errors{
		errorsTestSet[0],
		errorsTestSet[3],
	}
	actualStr := fmt.Sprintf("%#v", errorsTestSet.WithClass("ClassA"))
	expectedStr := fmt.Sprintf("%#v", expected)
	if actualStr != expectedStr {
		t.Errorf("Expected:\n%s\nbut got:\n%s", expectedStr, actualStr)
	}
}
*/

var errorsTestSet = Errors{
	Error{
		FieldNames:     []string{},
		Classification: "ClassA",
		Message:        "Foobar",
	},
	Error{
		FieldNames:     []string{},
		Classification: "ClassB",
		Message:        "Foo",
	},
	Error{
		FieldNames:     []string{"field1", "field2"},
		Classification: "ClassB",
		Message:        "Foobar",
	},
	Error{
		FieldNames:     []string{"field2"},
		Classification: "ClassA",
		Message:        "Foobar",
	},
	Error{
		FieldNames:     []string{"field2"},
		Classification: "ClassB",
		Message:        "Foobar",
	},
}
