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
	"errors"
	"testing"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/config/source"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig_Instance(t *testing.T) {
	// Test that the default config instance is properly initialized
	assert.NotNil(t, cfg)
	assert.Implements(t, (*Config)(nil), cfg)
}

func TestDefaultFunctions(t *testing.T) {
	// Test Set and Get
	err := Set("test.key", "test_value")
	require.NoError(t, err)
	assert.Equal(t, "test_value", Get("test.key").String())

	// Test GetMulti
	err = Set("section1.key1", "value1")
	err = Set("section2.key2", "value2")
	require.NoError(t, err)

	multiValue := GetMulti("section1", "section2")
	assert.NotNil(t, multiValue.Map())

	// Test SetMulti
	keys := []string{"multi1", "multi2"}
	values := []interface{}{"val1", 123}
	err = SetMulti(keys, values)
	require.NoError(t, err)

	assert.Equal(t, "val1", Get("multi1").String())
	assert.Equal(t, 123, Get("multi2").Int())

	// Test ValueToValues
	testValue := Get("test.key")
	resultValues := ValueToValues(testValue)
	assert.NotNil(t, resultValues)
}

func TestDefaultGetterFunctions(t *testing.T) {
	// Set up test data
	testData := map[string]interface{}{
		"bool_key":     true,
		"int_key":      42,
		"int64_key":    int64(1234567890),
		"string_key":   "hello world",
		"float_key":    3.14159,
		"duration_key": "1h30m",
		"bytes_key":    []byte("test bytes"),
		"slice_key":    []string{"a", "b", "c"},
		"map_key":      map[string]string{"key1": "val1", "key2": "val2"},
		"big_map_key":  map[string]interface{}{"nested": "value"},
	}

	for key, value := range testData {
		err := Set(key, value)
		require.NoError(t, err)
	}

	// Test GetBool
	assert.Equal(t, true, GetBool("bool_key"))
	assert.Equal(t, false, GetBool("nonexistent_bool"))
	assert.Equal(t, true, GetBool("nonexistent_bool", true))
	assert.Equal(t, false, GetBool("nonexistent_bool", false))

	// Test GetInt
	assert.Equal(t, 42, GetInt("int_key"))
	assert.Equal(t, 0, GetInt("nonexistent_int"))
	assert.Equal(t, 999, GetInt("nonexistent_int", 999))

	// Test GetInt64
	assert.Equal(t, int64(1234567890), GetInt64("int64_key"))
	assert.Equal(t, int64(0), GetInt64("nonexistent_int64"))
	assert.Equal(t, int64(888), GetInt64("nonexistent_int64", 888))

	// Test GetString
	assert.Equal(t, "hello world", GetString("string_key"))
	assert.Equal(t, "", GetString("nonexistent_string"))
	assert.Equal(t, "default", GetString("nonexistent_string", "default"))

	// Test GetBytes
	assert.Equal(t, []byte("test bytes"), GetBytes("bytes_key"))
	assert.Nil(t, GetBytes("nonexistent_bytes"))
	assert.Equal(t, []byte("default"), GetBytes("nonexistent_bytes", []byte("default")))

	// Test GetStringSlice
	assert.Equal(t, []string{"a", "b", "c"}, GetStringSlice("slice_key"))
	assert.Nil(t, GetStringSlice("nonexistent_slice"))
	assert.Equal(t, []string{"default"}, GetStringSlice("nonexistent_slice", []string{"default"}))

	// Test GetStringMap
	expectedMap := map[string]string{"key1": "val1", "key2": "val2"}
	assert.Equal(t, expectedMap, GetStringMap("map_key"))
	assert.Equal(t, map[string]string{}, GetStringMap("nonexistent_map"))
	assert.Equal(t, map[string]string{"default": "value"}, GetStringMap("nonexistent_map", map[string]string{"default": "value"}))

	// Test GetMap
	expectedBigMap := map[string]interface{}{"nested": "value"}
	assert.Equal(t, expectedBigMap, GetMap("big_map_key"))
	assert.Equal(t, map[string]interface{}{}, GetMap("nonexistent_big_map"))
	assert.Equal(t, map[string]interface{}{"default": "value"}, GetMap("nonexistent_big_map", map[string]interface{}{"default": "value"}))

	// Test GetFloat64
	assert.Equal(t, 3.14159, GetFloat64("float_key"))
	assert.Equal(t, 0.0, GetFloat64("nonexistent_float"))
	assert.Equal(t, 9.99, GetFloat64("nonexistent_float", 9.99))

	// Test GetDuration
	expectedDuration, err := time.ParseDuration("1h30m")
	require.NoError(t, err)
	assert.Equal(t, expectedDuration, GetDuration("duration_key"))
	assert.Equal(t, time.Duration(0), GetDuration("nonexistent_duration"))
	assert.Equal(t, time.Hour, GetDuration("nonexistent_duration", time.Hour))
}

func TestDefaultScan(t *testing.T) {
	type TestConfig struct {
		StringField string `yaml:"stringField"`
		IntField    int    `yaml:"intField"`
		BoolField   bool   `yaml:"boolField"`
	}

	// Set test data
	err := Set("scan_test.stringField", "test_value")
	require.NoError(t, err)
	err = Set("scan_test.intField", 123)
	require.NoError(t, err)
	err = Set("scan_test.boolField", true)
	require.NoError(t, err)

	// Test scanning
	var config TestConfig
	err = Scan("scan_test", &config)
	require.NoError(t, err)

	assert.Equal(t, "test_value", config.StringField)
	assert.Equal(t, 123, config.IntField)
	assert.Equal(t, true, config.BoolField)
}

func TestDefaultLoadSource(t *testing.T) {
	// Create a mock source
	mockData := map[string]interface{}{
		"source": map[string]interface{}{
			"loaded": "value",
		},
	}

	mockSource := &mockSource{
		data: mockData,
	}

	// Test LoadSource
	err := LoadSource(mockSource)
	require.NoError(t, err)

	assert.Equal(t, "value", Get("source.loaded").String())
}

func TestDefaultWatchers(t *testing.T) {
	var eventReceived WatchEvent
	eventHandler := func(event WatchEvent) {
		eventReceived = event
	}

	// Test AddWatcher
	err := AddWatcher("test_watch", eventHandler)
	require.NoError(t, err)

	// Give some time for the initial event to be processed
	time.Sleep(10 * time.Millisecond)

	// Verify initial event was received
	require.NotNil(t, eventReceived)
	assert.Equal(t, WatchEventUpd, eventReceived.Type())

	// Test DelWatcher
	err = DelWatcher("test_watch", eventHandler)
	require.NoError(t, err)
}

func TestDefaultValueConversionFunctions(t *testing.T) {
	// Test with string representations
	err := Set("string.bool", "true")
	require.NoError(t, err)
	assert.Equal(t, true, GetBool("string.bool"))

	err = Set("string.int", "42")
	require.NoError(t, err)
	assert.Equal(t, 42, GetInt("string.int"))

	err = Set("string.int64", "1234567890")
	require.NoError(t, err)
	assert.Equal(t, int64(1234567890), GetInt64("string.int64"))

	err = Set("string.float", "3.14")
	require.NoError(t, err)
	assert.Equal(t, 3.14, GetFloat64("string.float"))

	err = Set("string.duration", "1h")
	require.NoError(t, err)
	assert.Equal(t, time.Hour, GetDuration("string.duration"))

	// Test with invalid values
	err = Set("invalid.bool", "not_a_bool")
	require.NoError(t, err)
	assert.Equal(t, false, GetBool("invalid.bool"))
	assert.Equal(t, true, GetBool("invalid.bool", true)) // with default

	err = Set("invalid.int", "not_an_int")
	require.NoError(t, err)
	assert.Equal(t, 0, GetInt("invalid.int"))
	assert.Equal(t, 999, GetInt("invalid.int", 999)) // with default
}

func TestDefaultEdgeCases(t *testing.T) {
	// Test with nil values
	err := Set("nil_key", nil)
	require.NoError(t, err)
	assert.Equal(t, "", GetString("nil_key"))
	assert.Equal(t, "default", GetString("nil_key", "default"))

	// Test with complex nested structures
	complexData := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"value": "deep_value",
			},
		},
	}
	err = Set("complex.nested", complexData)
	require.NoError(t, err)
	assert.Equal(t, "deep_value", Get("complex.nested.level1.level2.value").String())

	// Test Bytes function
	assert.NotEmpty(t, Bytes())
	assert.Contains(t, string(Bytes()), "complex")
}

func TestDefaultConcurrentAccess(t *testing.T) {
	// Test concurrent read/write access
	done := make(chan bool, 2)

	// Goroutine 1: Write
	go func() {
		for i := 0; i < 100; i++ {
			Set("concurrent.key", i)
		}
		done <- true
	}()

	// Goroutine 2: Read
	go func() {
		for i := 0; i < 100; i++ {
			_ = Get("concurrent.key").Int()
		}
		done <- true
	}()

	// Wait for both to complete
	<-done
	<-done
}

func TestDefaultTypeConversions(t *testing.T) {
	// Test various type conversions
	testCases := []struct {
		name     string
		setKey   string
		setValue interface{}
		testFunc func(t *testing.T)
	}{
		{
			name:     "float to int",
			setKey:   "float_to_int",
			setValue: 3.14,
			testFunc: func(t *testing.T) {
				assert.Equal(t, 0, GetInt("float_to_int")) // float should convert to 0
			},
		},
		{
			name:     "bool to string",
			setKey:   "bool_to_string",
			setValue: true,
			testFunc: func(t *testing.T) {
				assert.Equal(t, "", GetString("bool_to_string")) // bool should convert to empty string
			},
		},
		{
			name:     "slice to string",
			setKey:   "slice_to_string",
			setValue: []string{"a", "b"},
			testFunc: func(t *testing.T) {
				assert.Equal(t, "", GetString("slice_to_string")) // slice should convert to empty string
			},
		},
		{
			name:     "map to bool",
			setKey:   "map_to_bool",
			setValue: map[string]interface{}{"key": "value"},
			testFunc: func(t *testing.T) {
				assert.Equal(t, false, GetBool("map_to_bool")) // map should convert to false
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Set(tc.setKey, tc.setValue)
			require.NoError(t, err)
			tc.testFunc(t)
		})
	}
}

func TestDefaultLoadSourceError(t *testing.T) {
	// Create a mock source that returns error
	mockSource := &mockSource{
		returnError: true,
	}

	err := LoadSource(mockSource)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock error")
}

func TestDefaultValueToValues(t *testing.T) {
	// Create a test value
	testMap := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"nested": map[string]interface{}{
			"inner": "inner_value",
		},
	}

	// Set the test map
	err := Set("value_to_values_test", testMap)
	require.NoError(t, err)

	// Get the value and convert to values
	val := Get("value_to_values_test")
	values := ValueToValues(val)

	// Test the converted values
	assert.Equal(t, "value1", values.Get("key1").String())
	assert.Equal(t, 123, values.Get("key2").Int())
	assert.Equal(t, "inner_value", values.Get("nested.inner").String())
}

// Mock source for testing
type mockSource struct {
	data        map[string]interface{}
	returnError bool
	changeable  bool
}

func (m *mockSource) Name() string {
	return "mock"
}

func (m *mockSource) Read() (source.SourceData, error) {
	if m.returnError {
		return nil, errors.New("mock error")
	}
	return source.NewMapSourceData(source.PriorityFile, m.data), nil
}

func (m *mockSource) Changeable() bool {
	return m.changeable
}

func (m *mockSource) Watch() (<-chan source.SourceData, error) {
	return nil, nil
}

func (m *mockSource) Close() error {
	return nil
}
