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

package otel

import (
	"testing"

	"go.opentelemetry.io/otel/attribute"

	"github.com/stretchr/testify/assert"
)

func TestParseAttributes_EmptyMap(t *testing.T) {
	attrsMap := map[string]interface{}{}
	attrs := ParseAttributes(attrsMap)
	assert.Empty(t, attrs)
}

func TestParseAttributes_NilMap(t *testing.T) {
	var attrsMap map[string]interface{}
	attrs := ParseAttributes(attrsMap)
	assert.Empty(t, attrs)
}

func TestParseAttributes_Bool(t *testing.T) {
	attrsMap := map[string]interface{}{
		"bool_true":  true,
		"bool_false": false,
	}
	attrs := ParseAttributes(attrsMap)

	assert.Len(t, attrs, 2)

	// Create a map for lookup since order is not guaranteed
	attrMap := make(map[string]attribute.KeyValue)
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr
	}

	assert.Equal(t, attribute.Bool("bool_true", true), attrMap["bool_true"])
	assert.Equal(t, attribute.Bool("bool_false", false), attrMap["bool_false"])
}

func TestParseAttributes_String(t *testing.T) {
	attrsMap := map[string]interface{}{
		"string_empty":       "",
		"string_value":       "test value",
		"string_with_spaces": "  spaced value  ",
	}
	attrs := ParseAttributes(attrsMap)

	assert.Len(t, attrs, 3)

	// Create a map for lookup since order is not guaranteed
	attrMap := make(map[string]attribute.KeyValue)
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr
	}

	assert.Equal(t, attribute.String("string_empty", ""), attrMap["string_empty"])
	assert.Equal(t, attribute.String("string_value", "test value"), attrMap["string_value"])
	assert.Equal(t, attribute.String("string_with_spaces", "  spaced value  "), attrMap["string_with_spaces"])
}

func TestParseAttributes_Int64(t *testing.T) {
	attrsMap := map[string]interface{}{
		"int_positive": int64(42),
		"int_negative": int64(-10),
		"int_zero":     int64(0),
		"int_max":      int64(9223372036854775807),
		"int_min":      int64(-9223372036854775808),
	}
	attrs := ParseAttributes(attrsMap)

	assert.Len(t, attrs, 5)

	// Create a map for lookup since order is not guaranteed
	attrMap := make(map[string]attribute.KeyValue)
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr
	}

	assert.Equal(t, attribute.Int64("int_positive", 42), attrMap["int_positive"])
	assert.Equal(t, attribute.Int64("int_negative", -10), attrMap["int_negative"])
	assert.Equal(t, attribute.Int64("int_zero", 0), attrMap["int_zero"])
	assert.Equal(t, attribute.Int64("int_max", 9223372036854775807), attrMap["int_max"])
	assert.Equal(t, attribute.Int64("int_min", -9223372036854775808), attrMap["int_min"])
}

func TestParseAttributes_Float64(t *testing.T) {
	attrsMap := map[string]interface{}{
		"float_positive":   3.14,
		"float_negative":   -2.71,
		"float_zero":       0.0,
		"float_scientific": 1.5e10,
	}
	attrs := ParseAttributes(attrsMap)

	assert.Len(t, attrs, 4)

	// Create a map for lookup since order is not guaranteed
	attrMap := make(map[string]attribute.KeyValue)
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr
	}

	assert.Equal(t, attribute.Float64("float_positive", 3.14), attrMap["float_positive"])
	assert.Equal(t, attribute.Float64("float_negative", -2.71), attrMap["float_negative"])
	assert.Equal(t, attribute.Float64("float_zero", 0.0), attrMap["float_zero"])
	assert.Equal(t, attribute.Float64("float_scientific", 1.5e10), attrMap["float_scientific"])
}

func TestParseAttributes_Float64Slice(t *testing.T) {
	attrsMap := map[string]interface{}{
		"float_slice_empty":    []float64{},
		"float_slice_single":   []float64{3.14},
		"float_slice_multiple": []float64{1.1, 2.2, 3.3},
		"float_slice_mixed":    []float64{-1.0, 0.0, 1.0},
	}
	attrs := ParseAttributes(attrsMap)

	assert.Len(t, attrs, 4)

	// Create a map for lookup since order is not guaranteed
	attrMap := make(map[string]attribute.KeyValue)
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr
	}

	assert.Equal(t, attribute.Float64Slice("float_slice_empty", []float64{}), attrMap["float_slice_empty"])
	assert.Equal(t, attribute.Float64Slice("float_slice_single", []float64{3.14}), attrMap["float_slice_single"])
	assert.Equal(t, attribute.Float64Slice("float_slice_multiple", []float64{1.1, 2.2, 3.3}), attrMap["float_slice_multiple"])
	assert.Equal(t, attribute.Float64Slice("float_slice_mixed", []float64{-1.0, 0.0, 1.0}), attrMap["float_slice_mixed"])
}

func TestParseAttributes_Int64Slice(t *testing.T) {
	attrsMap := map[string]interface{}{
		"int_slice_empty":    []int64{},
		"int_slice_single":   []int64{42},
		"int_slice_multiple": []int64{1, 2, 3},
		"int_slice_mixed":    []int64{-1, 0, 1},
	}
	attrs := ParseAttributes(attrsMap)

	assert.Len(t, attrs, 4)

	// Create a map for lookup since order is not guaranteed
	attrMap := make(map[string]attribute.KeyValue)
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr
	}

	assert.Equal(t, attribute.Int64Slice("int_slice_empty", []int64{}), attrMap["int_slice_empty"])
	assert.Equal(t, attribute.Int64Slice("int_slice_single", []int64{42}), attrMap["int_slice_single"])
	assert.Equal(t, attribute.Int64Slice("int_slice_multiple", []int64{1, 2, 3}), attrMap["int_slice_multiple"])
	assert.Equal(t, attribute.Int64Slice("int_slice_mixed", []int64{-1, 0, 1}), attrMap["int_slice_mixed"])
}

func TestParseAttributes_StringSlice(t *testing.T) {
	attrsMap := map[string]interface{}{
		"string_slice_empty":    []string{},
		"string_slice_single":   []string{"hello"},
		"string_slice_multiple": []string{"a", "b", "c"},
		"string_slice_mixed":    []string{"", " ", "test"},
	}
	attrs := ParseAttributes(attrsMap)

	assert.Len(t, attrs, 4)

	// Create a map for lookup since order is not guaranteed
	attrMap := make(map[string]attribute.KeyValue)
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr
	}

	assert.Equal(t, attribute.StringSlice("string_slice_empty", []string{}), attrMap["string_slice_empty"])
	assert.Equal(t, attribute.StringSlice("string_slice_single", []string{"hello"}), attrMap["string_slice_single"])
	assert.Equal(t, attribute.StringSlice("string_slice_multiple", []string{"a", "b", "c"}), attrMap["string_slice_multiple"])
	assert.Equal(t, attribute.StringSlice("string_slice_mixed", []string{"", " ", "test"}), attrMap["string_slice_mixed"])
}

func TestParseAttributes_BoolSlice(t *testing.T) {
	attrsMap := map[string]interface{}{
		"bool_slice_empty":     []bool{},
		"bool_slice_single":    []bool{true},
		"bool_slice_multiple":  []bool{true, false, true},
		"bool_slice_all_true":  []bool{true, true, true},
		"bool_slice_all_false": []bool{false, false},
	}
	attrs := ParseAttributes(attrsMap)

	assert.Len(t, attrs, 5)

	// Create a map for lookup since order is not guaranteed
	attrMap := make(map[string]attribute.KeyValue)
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr
	}

	assert.Equal(t, attribute.BoolSlice("bool_slice_empty", []bool{}), attrMap["bool_slice_empty"])
	assert.Equal(t, attribute.BoolSlice("bool_slice_single", []bool{true}), attrMap["bool_slice_single"])
	assert.Equal(t, attribute.BoolSlice("bool_slice_multiple", []bool{true, false, true}), attrMap["bool_slice_multiple"])
	assert.Equal(t, attribute.BoolSlice("bool_slice_all_true", []bool{true, true, true}), attrMap["bool_slice_all_true"])
	assert.Equal(t, attribute.BoolSlice("bool_slice_all_false", []bool{false, false}), attrMap["bool_slice_all_false"])
}

func TestParseAttributes_DefaultType(t *testing.T) {
	attrsMap := map[string]interface{}{
		"pointer":   (*string)(nil),
		"interface": interface{}(nil),
		"channel":   make(chan int),
		"func":      func() {},
		"struct":    struct{ Name string }{Name: "test"},
		"map_val":   map[string]int{"key": 1},
		"array":     [3]int{1, 2, 3},
		"int32":     int32(42),     // This should go to default case
		"uint64":    uint64(100),   // This should go to default case
		"complex":   complex(1, 2), // This should go to default case
	}
	attrs := ParseAttributes(attrsMap)

	assert.Len(t, attrs, 10)

	// All should be converted to string format using fmt.Sprintf
	for _, attr := range attrs {
		// Since these are nil/complex types, they'll be stringified representations
		assert.Equal(t, attribute.STRING, attr.Value.Type())
	}
}

func TestParseAttributes_MixedTypes(t *testing.T) {
	attrsMap := map[string]interface{}{
		"bool_val":     true,
		"string_val":   "hello",
		"int64_val":    int64(42),
		"float64_val":  3.14,
		"int_slice":    []int64{1, 2, 3},
		"float_slice":  []float64{1.1, 2.2},
		"string_slice": []string{"a", "b"},
		"bool_slice":   []bool{true, false},
		"unknown":      struct{}{},
	}
	attrs := ParseAttributes(attrsMap)

	assert.Len(t, attrs, 9)

	// Verify each type was converted correctly
	foundTypes := make(map[string]bool)
	for _, attr := range attrs {
		switch attr.Value.Type() {
		case attribute.BOOL:
			foundTypes["bool"] = true
		case attribute.STRING:
			foundTypes["string"] = true
		case attribute.INT64:
			foundTypes["int64"] = true
		case attribute.FLOAT64:
			foundTypes["float64"] = true
		case attribute.BOOLSLICE:
			foundTypes["bool_slice"] = true
		case attribute.INT64SLICE:
			foundTypes["int64_slice"] = true
		case attribute.FLOAT64SLICE:
			foundTypes["float64_slice"] = true
		case attribute.STRINGSLICE:
			foundTypes["string_slice"] = true
		}
	}

	assert.True(t, foundTypes["bool"])
	assert.True(t, foundTypes["string"])
	assert.True(t, foundTypes["int64"])
	assert.True(t, foundTypes["float64"])
	assert.True(t, foundTypes["bool_slice"])
	assert.True(t, foundTypes["int64_slice"])
	assert.True(t, foundTypes["float64_slice"])
	assert.True(t, foundTypes["string_slice"])
}

func TestParseAttributes_OrderPreservation(t *testing.T) {
	// Test that the order of attributes is preserved (important for some systems)
	attrsMap := map[string]interface{}{
		"first":  "first_value",
		"second": int64(2),
		"third":  true,
	}
	attrs := ParseAttributes(attrsMap)

	assert.Len(t, attrs, 3)
	// Note: Map iteration order in Go is not guaranteed, so we can't test exact order
	// But we can verify that all expected attributes are present
	var foundFirst, foundSecond, foundThird bool
	for _, attr := range attrs {
		switch attr.Key {
		case "first":
			foundFirst = true
			assert.Equal(t, attribute.String("first", "first_value"), attr)
		case "second":
			foundSecond = true
			assert.Equal(t, attribute.Int64("second", 2), attr)
		case "third":
			foundThird = true
			assert.Equal(t, attribute.Bool("third", true), attr)
		}
	}

	assert.True(t, foundFirst)
	assert.True(t, foundSecond)
	assert.True(t, foundThird)
}
