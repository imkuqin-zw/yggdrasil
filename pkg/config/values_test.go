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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValues(t *testing.T) {
	// Test with nil map
	v := newValues(".", nil)
	require.NotNil(t, v)
	assert.NotNil(t, v.val)

	// Test with empty map
	emptyMap := map[string]interface{}{}
	v = newValues(".", emptyMap)
	require.NotNil(t, v)
	assert.Equal(t, emptyMap, v.val)

	// Test with data
	testMap := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	v = newValues(".", testMap)
	require.NotNil(t, v)
	assert.Equal(t, testMap, v.val)

	// Verify that the map is cloned (modifying original shouldn't affect values)
	testMap["key3"] = "new_value"
	assert.NotContains(t, v.val, "key3")
}

func TestValues_Get(t *testing.T) {
	testData := map[string]interface{}{
		"simple":  "value",
		"nested":  map[string]interface{}{"deep": "deep_value"},
		"number":  42,
		"boolean": true,
	}

	v := newValues(".", testData)

	// Test getting simple key
	val := v.Get("simple")
	assert.Equal(t, "value", val.String())

	// Test getting non-existent key
	val = v.Get("nonexistent")
	assert.Equal(t, "", val.String())

	// Test getting empty key (should return entire map)
	val = v.Get("")
	assert.Equal(t, testData, val.Map())

	// Test getting nested value
	val = v.Get("nested.deep")
	assert.Equal(t, "deep_value", val.String())

	// Test getting numeric value
	val = v.Get("number")
	assert.Equal(t, 42, val.Int())

	// Test getting boolean value
	val = v.Get("boolean")
	assert.Equal(t, true, val.Bool())
}

func TestValues_GetMulti(t *testing.T) {
	testData := map[string]interface{}{
		"section1": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
		"section2": map[string]interface{}{
			"key3": "value3",
			"key4": "value4",
		},
		"standalone": "standalone_value",
	}

	v := newValues(".", testData)

	// Test getting multiple sections
	val := v.GetMulti("section1", "section2")
	result := val.Map()

	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, "value2", result["key2"])
	assert.Equal(t, "value3", result["key3"])
	assert.Equal(t, "value4", result["key4"])
	assert.NotContains(t, result, "standalone")

	// Test getting non-existent keys
	val = v.GetMulti("nonexistent1", "nonexistent2")
	result = val.Map()
	assert.Empty(t, result)

	// Test getting mix of existent and non-existent keys
	val = v.GetMulti("section1", "nonexistent")
	result = val.Map()
	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, "value2", result["key2"])
}

func TestValues_Set(t *testing.T) {
	testData := map[string]interface{}{
		"existing": map[string]interface{}{
			"nested": "old_value",
		},
	}

	v := newValues(".", testData)

	// Test setting existing key in existing nested map
	err := v.Set("existing.nested", "new_value")
	require.NoError(t, err)
	assert.Equal(t, "new_value", v.Get("existing.nested").String())

	// Test setting new key in existing nested map
	err = v.Set("existing.new_key", "new_value")
	require.NoError(t, err)
	assert.Equal(t, "new_value", v.Get("existing.new_key").String())

	// Test setting new nested structure - this won't work because path doesn't exist
	err = v.Set("new.nested.deep.key", "deep_value")
	require.NoError(t, err)
	// Since the path doesn't exist, the value won't actually be set
	assert.Equal(t, "", v.Get("new.nested.deep.key").String())

	// Test setting non-existent path (should return nil but no error)
	err = v.Set("nonexistent.path.key", "value")
	require.NoError(t, err)
	// Since the path doesn't exist, the value won't actually be set
	assert.Equal(t, "", v.Get("nonexistent.path.key").String())
}

func TestValues_SetMulti(t *testing.T) {
	testData := map[string]interface{}{
		"existing": map[string]interface{}{
			"old": "old_value",
		},
	}

	v := newValues(".", testData)

	// Test setting multiple existing keys
	keys := []string{"existing.old", "existing.new"}
	values := []interface{}{"new_value", "another_value"}

	err := v.SetMulti(keys, values)
	require.NoError(t, err)

	assert.Equal(t, "new_value", v.Get("existing.old").String())
	assert.Equal(t, "another_value", v.Get("existing.new").String())

	// Test with mismatched keys and values length
	err = v.SetMulti([]string{"key1"}, []interface{}{"value1", "value2"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quantity of key and value does not match")

	// Test with empty arrays
	err = v.SetMulti([]string{}, []interface{}{})
	require.NoError(t, err)
}

func TestValues_Del(t *testing.T) {
	testData := map[string]interface{}{
		"delete_me": map[string]interface{}{
			"nested": "value_to_delete",
		},
		"keep_me": "keep_this_value",
	}

	v := newValues(".", testData)

	// Verify initial state
	assert.Equal(t, "value_to_delete", v.Get("delete_me.nested").String())
	assert.Equal(t, "keep_this_value", v.Get("keep_me").String())

	// Delete nested key
	err := v.Del("delete_me.nested")
	require.NoError(t, err)
	assert.Equal(t, "", v.Get("delete_me.nested").String()) // Should return default (empty string)

	// Verify other key is still there
	assert.Equal(t, "keep_this_value", v.Get("keep_me").String())

	// Delete non-existent key (should not error)
	err = v.Del("nonexistent.path.key")
	require.NoError(t, err)

	// Delete key from non-existent path (should not error)
	err = v.Del("nonexistent.nested.key")
	require.NoError(t, err)
}

func TestValues_Map(t *testing.T) {
	// Test with data
	testData := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": map[string]interface{}{
			"nested": "value",
		},
	}

	v := newValues(".", testData)
	result := v.Map()

	assert.Equal(t, testData, result)

	// Verify it's a clone (modifying result shouldn't affect original)
	result["new_key"] = "new_value"
	assert.NotContains(t, v.val, "new_key")

	// Test with nil data
	v = newValues(".", nil)
	result = v.Map()
	assert.NotNil(t, result)
}

func TestValues_Scan(t *testing.T) {
	type TestStruct struct {
		StringField string  `yaml:"stringField"`
		IntField    int     `yaml:"intField"`
		BoolField   bool    `yaml:"boolField"`
		FloatField  float64 `yaml:"floatField"`
		Nested      struct {
			Inner string `yaml:"inner"`
		} `yaml:"nested"`
		WithDefault string `yaml:"withDefault" default:"default_value"`
	}

	testData := map[string]interface{}{
		"stringField": "test_string",
		"intField":    123,
		"boolField":   true,
		"floatField":  3.14,
		"nested": map[string]interface{}{
			"inner": "inner_value",
		},
	}

	v := newValues(".", testData)

	// Test scanning to struct
	var result TestStruct
	err := v.Scan(&result)
	require.NoError(t, err)

	assert.Equal(t, "test_string", result.StringField)
	assert.Equal(t, 123, result.IntField)
	assert.Equal(t, true, result.BoolField)
	assert.Equal(t, 3.14, result.FloatField)
	assert.Equal(t, "inner_value", result.Nested.Inner)
	assert.Equal(t, "default_value", result.WithDefault) // Should be set by defaults

	// Test scanning to map
	var mapResult map[string]interface{}
	err = v.Scan(&mapResult)
	require.NoError(t, err)

	assert.Equal(t, "test_string", mapResult["stringField"])
	assert.Equal(t, 123, mapResult["intField"])

	// Test scanning to slice
	var sliceResult []string
	err = v.Scan(&sliceResult)
	assert.Error(t, err) // Should error, because data is not a slice
}

func TestValues_Bytes(t *testing.T) {
	// Test with data
	testData := map[string]interface{}{
		"string": "value",
		"number": 42,
		"bool":   true,
		"nested": map[string]interface{}{
			"inner": "inner_value",
		},
	}

	v := newValues(".", testData)
	result := v.Bytes()

	assert.NotNil(t, result)
	assert.Contains(t, string(result), "value")
	assert.Contains(t, string(result), "42")
	assert.Contains(t, string(result), "true")
	assert.Contains(t, string(result), "inner_value")

	// Test with nil data
	v = newValues(".", nil)
	result = v.Bytes()
	assert.Equal(t, []byte("{}"), result)
}

func TestValues_get(t *testing.T) {
	testData := map[string]interface{}{
		"direct": "direct_value",
		"nested": map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": "deep_value",
			},
		},
	}

	v := newValues(".", testData)

	// Test getting direct key
	result := v.get("direct")
	assert.Equal(t, "direct_value", result)

	// Test getting nested key
	result = v.get("nested.level1.level2")
	assert.Equal(t, "deep_value", result)

	// Test getting non-existent key
	result = v.get("nonexistent")
	assert.Nil(t, result)

	// Test getting intermediate nested key
	result = v.get("nested.level1")
	expected := map[string]interface{}{"level2": "deep_value"}
	assert.Equal(t, expected, result)
}

func TestValues_deepSearchInMap(t *testing.T) {
	v := newValues(".", nil)

	testMap := map[string]interface{}{
		"key1": "value1",
		"nested": map[string]interface{}{
			"key2": "value2",
			"deep": map[string]interface{}{
				"key3": "value3",
			},
		},
	}

	// Test finding direct key
	result := v.deepSearchInMap(testMap, "key1", ".")
	assert.Equal(t, "value1", result)

	// Test finding nested key
	result = v.deepSearchInMap(testMap, "nested.key2", ".")
	assert.Equal(t, "value2", result)

	// Test finding deeply nested key
	result = v.deepSearchInMap(testMap, "nested.deep.key3", ".")
	assert.Equal(t, "value3", result)

	// Test non-existent key
	result = v.deepSearchInMap(testMap, "nonexistent", ".")
	assert.Nil(t, result)

	// Test non-existent nested key
	result = v.deepSearchInMap(testMap, "nested.nonexistent", ".")
	assert.Nil(t, result)
}

func TestValues_WithDifferentDelimiter(t *testing.T) {
	testData := map[string]interface{}{
		"section1": map[string]interface{}{
			"subsection": map[string]interface{}{
				"key": "value",
			},
		},
	}

	v := newValues("_", testData)

	// Test getting with underscore delimiter
	val := v.Get("section1_subsection_key")
	assert.Equal(t, "value", val.String())

	// Test setting with underscore delimiter - this won't work because path doesn't exist
	err := v.Set("section1_new_newKey", "newValue")
	require.NoError(t, err)
	// Since the path doesn't exist, the value won't actually be set
	assert.Equal(t, "", v.Get("section1_new_newKey").String())
}

func TestValues_EmptyAndEdgeCases(t *testing.T) {
	// Test with completely empty data
	v := newValues(".", map[string]interface{}{})

	val := v.Get("any_key")
	assert.Equal(t, "", val.String())

	err := v.Set("new.key", "value")
	require.NoError(t, err)
	// Path doesn't exist, so value won't be set
	assert.Equal(t, "", v.Get("new.key").String())

	// Test with empty string key
	testData := map[string]interface{}{
		"": "empty_key_value",
	}
	v = newValues(".", testData)

	val = v.Get("")
	assert.Equal(t, testData, val.Map()) // Empty key returns whole map

	// Test setting empty key
	err = v.Set("", "new_value")
	require.NoError(t, err)
}

func TestValues_ComplexNestedStructures(t *testing.T) {
	complexData := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"final": "final_value",
				},
				"sibling": "sibling_value",
			},
		},
		"other": "other_value",
	}

	v := newValues(".", complexData)

	// Test deep nesting access
	val := v.Get("level1.level2.level3.final")
	assert.Equal(t, "final_value", val.String())

	// Test sibling access
	val = v.Get("level1.level2.sibling")
	assert.Equal(t, "sibling_value", val.String())

	// Test modifying deep nested value
	err := v.Set("level1.level2.level3.final", "modified_value")
	require.NoError(t, err)
	assert.Equal(t, "modified_value", v.Get("level1.level2.level3.final").String())

	// Verify other values are unchanged
	assert.Equal(t, "sibling_value", v.Get("level1.level2.sibling").String())
	assert.Equal(t, "other_value", v.Get("other").String())
}
