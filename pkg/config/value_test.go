// Copyright 2022 The imkuqin-zw Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValue_Bool(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
		def      bool
	}{
		{
			name:     "true bool",
			input:    true,
			expected: true,
		},
		{
			name:     "false bool",
			input:    false,
			expected: false,
		},
		{
			name:     "true string",
			input:    "true",
			expected: true,
		},
		{
			name:     "false string",
			input:    "false",
			expected: false,
		},
		{
			name:     "1 string",
			input:    "1",
			expected: true,
		},
		{
			name:     "0 string",
			input:    "0",
			expected: false,
		},
		{
			name:     "invalid string with default",
			input:    "invalid",
			expected: true,
			def:      true,
		},
		{
			name:     "invalid string without default",
			input:    "invalid",
			expected: false,
		},
		{
			name:     "nil with default",
			input:    nil,
			expected: true,
			def:      true,
		},
		{
			name:     "nil without default",
			input:    nil,
			expected: false,
		},
		{
			name:     "int type",
			input:    123,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newValue(tt.input)
			if len([]bool{tt.def}) > 0 {
				assert.Equal(t, tt.expected, v.Bool(tt.def))
			} else {
				assert.Equal(t, tt.expected, v.Bool())
			}
		})
	}
}

func TestValue_Int(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int
		def      int
		hasDef   bool
	}{
		{
			name:     "int type",
			input:    42,
			expected: 42,
		},
		{
			name:     "valid string",
			input:    "123",
			expected: 123,
		},
		{
			name:     "invalid string with default",
			input:    "invalid",
			expected: 999,
			hasDef:   true,
			def:      999,
		},
		{
			name:     "invalid string without default",
			input:    "invalid",
			expected: 0,
		},
		{
			name:     "nil with default",
			input:    nil,
			expected: 888,
			hasDef:   true,
			def:      888,
		},
		{
			name:     "nil without default",
			input:    nil,
			expected: 0,
		},
		{
			name:     "float type",
			input:    3.14,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newValue(tt.input)
			if tt.hasDef {
				assert.Equal(t, tt.expected, v.Int(tt.def))
			} else {
				assert.Equal(t, tt.expected, v.Int())
			}
		})
	}
}

func TestValue_Int64(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int64
		def      int64
		hasDef   bool
	}{
		{
			name:     "int64 type",
			input:    int64(42),
			expected: 42,
		},
		{
			name:     "int type",
			input:    42,
			expected: 42,
		},
		{
			name:     "valid string",
			input:    "123",
			expected: 123,
		},
		{
			name:     "large number string",
			input:    "9223372036854775807",
			expected: 9223372036854775807,
		},
		{
			name:     "invalid string with default",
			input:    "invalid",
			expected: 999,
			hasDef:   true,
			def:      999,
		},
		{
			name:     "nil without default",
			input:    nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newValue(tt.input)
			if tt.hasDef {
				assert.Equal(t, tt.expected, v.Int64(tt.def))
			} else {
				assert.Equal(t, tt.expected, v.Int64())
			}
		})
	}
}

func TestValue_String(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
		def      string
		hasDef   bool
	}{
		{
			name:     "string type",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "int type",
			input:    42,
			expected: "",
		},
		{
			name:     "nil with default",
			input:    nil,
			expected: "default",
			hasDef:   true,
			def:      "default",
		},
		{
			name:     "nil without default",
			input:    nil,
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newValue(tt.input)
			if tt.hasDef {
				assert.Equal(t, tt.expected, v.String(tt.def))
			} else {
				assert.Equal(t, tt.expected, v.String())
			}
		})
	}
}

func TestValue_Float64(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
		def      float64
		hasDef   bool
	}{
		{
			name:     "float64 type",
			input:    3.14,
			expected: 3.14,
		},
		{
			name:     "int type",
			input:    42,
			expected: 42.0,
		},
		{
			name:     "valid string",
			input:    "3.14159",
			expected: 3.14159,
		},
		{
			name:     "invalid string with default",
			input:    "invalid",
			expected: 9.99,
			hasDef:   true,
			def:      9.99,
		},
		{
			name:     "invalid string without default",
			input:    "invalid",
			expected: 0.0,
		},
		{
			name:     "nil with default",
			input:    nil,
			expected: 1.23,
			hasDef:   true,
			def:      1.23,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newValue(tt.input)
			if tt.hasDef {
				assert.Equal(t, tt.expected, v.Float64(tt.def))
			} else {
				assert.Equal(t, tt.expected, v.Float64())
			}
		})
	}
}

func TestValue_Duration(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected time.Duration
		def      time.Duration
		hasDef   bool
	}{
		{
			name:     "duration type",
			input:    time.Second,
			expected: time.Second,
		},
		{
			name:     "valid string",
			input:    "1s",
			expected: time.Second,
		},
		{
			name:     "valid string with minutes",
			input:    "5m",
			expected: 5 * time.Minute,
		},
		{
			name:     "valid string with hours",
			input:    "2h",
			expected: 2 * time.Hour,
		},
		{
			name:     "invalid string with default",
			input:    "invalid",
			expected: time.Minute,
			hasDef:   true,
			def:      time.Minute,
		},
		{
			name:     "invalid string without default",
			input:    "invalid",
			expected: 0,
		},
		{
			name:     "nil with default",
			input:    nil,
			expected: time.Hour,
			hasDef:   true,
			def:      time.Hour,
		},
		{
			name:     "int type",
			input:    123,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newValue(tt.input)
			if tt.hasDef {
				assert.Equal(t, tt.expected, v.Duration(tt.def))
			} else {
				assert.Equal(t, tt.expected, v.Duration())
			}
		})
	}
}

func TestValue_StringSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []string
		def      []string
		hasDef   bool
	}{
		{
			name:     "string slice",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "interface slice",
			input:    []interface{}{"a", 1, "c", true},
			expected: []string{"a", "1", "c", "true"},
		},
		{
			name:     "string slice from map",
			input:    map[string]string{"key": "value"},
			expected: nil,
		},
		{
			name:     "nil with default",
			input:    nil,
			expected: []string{"default"},
			hasDef:   true,
			def:      []string{"default"},
		},
		{
			name:     "nil without default",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty interface slice",
			input:    []interface{}{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newValue(tt.input)
			if tt.hasDef {
				assert.Equal(t, tt.expected, v.StringSlice(tt.def))
			} else {
				assert.Equal(t, tt.expected, v.StringSlice())
			}
		})
	}
}

func TestValue_StringMap(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected map[string]string
		def      map[string]string
		hasDef   bool
	}{
		{
			name:     "string map",
			input:    map[string]string{"key1": "value1", "key2": "value2"},
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:     "interface map",
			input:    map[string]interface{}{"key1": "value1", "key2": 123},
			expected: map[string]string{},
		},
		{
			name:     "nil with default",
			input:    nil,
			expected: map[string]string{"default": "value"},
			hasDef:   true,
			def:      map[string]string{"default": "value"},
		},
		{
			name:     "nil without default",
			input:    nil,
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newValue(tt.input)
			if tt.hasDef {
				assert.Equal(t, tt.expected, v.StringMap(tt.def))
			} else {
				assert.Equal(t, tt.expected, v.StringMap())
			}
		})
	}
}

func TestValue_Map(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected map[string]interface{}
		def      map[string]interface{}
		hasDef   bool
	}{
		{
			name:     "interface map",
			input:    map[string]interface{}{"key1": "value1", "key2": 123},
			expected: map[string]interface{}{"key1": "value1", "key2": 123},
		},
		{
			name:     "nil with default",
			input:    nil,
			expected: map[string]interface{}{"default": "value"},
			hasDef:   true,
			def:      map[string]interface{}{"default": "value"},
		},
		{
			name:     "nil without default",
			input:    nil,
			expected: map[string]interface{}{},
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newValue(tt.input)
			if tt.hasDef {
				result := v.Map(tt.def)
				assert.Equal(t, tt.expected, result)
			} else {
				result := v.Map()
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestValue_Scan(t *testing.T) {
	type TestStruct struct {
		StringField string
		IntField    int
		BoolField   bool `yaml:"bool_field"`
	}

	tests := []struct {
		name     string
		input    interface{}
		target   interface{}
		wantErr  bool
		validate func(*testing.T, interface{})
	}{
		{
			name: "map to struct",
			input: map[string]interface{}{
				"stringField": "test",
				"intField":    42,
				"bool_field":  true,
			},
			target:  &TestStruct{},
			wantErr: false,
			validate: func(t *testing.T, v interface{}) {
				s := v.(*TestStruct)
				assert.Equal(t, "test", s.StringField)
				assert.Equal(t, 42, s.IntField)
				assert.Equal(t, true, s.BoolField)
			},
		},
		{
			name:    "array to slice",
			input:   []interface{}{"a", "b", "c"},
			target:  &[]string{},
			wantErr: false,
			validate: func(t *testing.T, v interface{}) {
				s := v.(*[]string)
				assert.Equal(t, []string{"a", "b", "c"}, *s)
			},
		},
		{
			name:    "nil to struct",
			input:   nil,
			target:  &TestStruct{},
			wantErr: false,
			validate: func(t *testing.T, v interface{}) {
				s := v.(*TestStruct)
				assert.Equal(t, "", s.StringField)
				assert.Equal(t, 0, s.IntField)
				assert.Equal(t, false, s.BoolField)
			},
		},
		{
			name:    "string to struct",
			input:   "simple string",
			target:  &TestStruct{},
			wantErr: false,
			validate: func(t *testing.T, v interface{}) {
				s := v.(*TestStruct)
				assert.Equal(t, "", s.StringField)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newValue(tt.input)
			err := v.Scan(tt.target)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tt.target)
				}
			}
		})
	}
}

func TestValue_Bytes(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []byte
		def      []byte
		hasDef   bool
	}{
		{
			name:     "byte slice",
			input:    []byte("hello"),
			expected: []byte("hello"),
		},
		{
			name:     "string",
			input:    "hello",
			expected: []byte("hello"),
		},
		{
			name:     "map",
			input:    map[string]interface{}{"key": "value"},
			expected: []byte(`{"key":"value"}`),
		},
		{
			name:     "nil with default",
			input:    nil,
			expected: []byte("default"),
			hasDef:   true,
			def:      []byte("default"),
		},
		{
			name:     "nil without default",
			input:    nil,
			expected: nil,
		},
		{
			name:     "number",
			input:    42,
			expected: []byte("42"),
		},
		{
			name:     "bool",
			input:    true,
			expected: []byte("true"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newValue(tt.input)
			var result []byte
			if tt.hasDef {
				result = v.Bytes(tt.def)
			} else {
				result = v.Bytes()
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValue_BytesJSONError(t *testing.T) {
	// Create a value that will cause JSON marshaling to fail
	invalidValue := newValue(func() {}) // Function can't be marshaled to JSON
	result := invalidValue.Bytes()
	// Should return nil when JSON marshaling fails
	assert.Nil(t, result)
}

func TestNewValue(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "string",
			input: "test",
		},
		{
			name:  "nil",
			input: nil,
		},
		{
			name:  "number",
			input: 42,
		},
		{
			name:  "map",
			input: map[string]interface{}{"key": "value"},
		},
		{
			name:  "slice",
			input: []interface{}{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newValue(tt.input)
			assert.NotNil(t, v)
			assert.Implements(t, (*Value)(nil), v)
		})
	}
}

func TestValue_JSONMarshalComplex(t *testing.T) {
	complexData := map[string]interface{}{
		"nested": map[string]interface{}{
			"array": []interface{}{1, "two", true},
			"null":  nil,
		},
		"number": 3.14159,
		"bool":   false,
	}

	v := newValue(complexData)
	result := v.Bytes()

	var unmarshaled map[string]interface{}
	err := json.Unmarshal(result, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, 3.14159, unmarshaled["number"])
	assert.Equal(t, false, unmarshaled["bool"])
	assert.NotNil(t, unmarshaled["nested"])
}
