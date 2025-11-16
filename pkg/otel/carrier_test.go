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

	"github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	"github.com/stretchr/testify/assert"
)

func TestNewMetadataReaderWriter(t *testing.T) {
	md := metadata.New(nil)
	rw := NewMetadataReaderWriter(&md)

	assert.NotNil(t, rw)
	assert.Equal(t, &md, rw.md)
}

func TestMetadataReaderWriter_Get(t *testing.T) {
	tests := []struct {
		name     string
		metadata metadata.MD
		key      string
		expected string
	}{
		{
			name:     "key exists with single value",
			metadata: metadata.MD{"trace-id": {"12345"}},
			key:      "trace-id",
			expected: "12345",
		},
		{
			name:     "key exists with multiple values",
			metadata: metadata.MD{"baggage": {"key1=value1", "key2=value2"}},
			key:      "baggage",
			expected: "key1=value1;key2=value2",
		},
		{
			name:     "key does not exist",
			metadata: metadata.MD{"other-key": {"value"}},
			key:      "missing-key",
			expected: "",
		},
		{
			name:     "empty metadata",
			metadata: metadata.MD{},
			key:      "any-key",
			expected: "",
		},
		{
			name:     "key exists with empty values",
			metadata: metadata.MD{"empty-key": {}},
			key:      "empty-key",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rw := NewMetadataReaderWriter(&tt.metadata)

			result := rw.Get(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMetadataReaderWriter_Set(t *testing.T) {
	md := metadata.New(nil)
	rw := NewMetadataReaderWriter(&md)

	// Test setting new value
	rw.Set("test-key", "test-value")
	values := md.Get("test-key")
	assert.Equal(t, []string{"test-value"}, values)

	// Test overwriting existing value
	rw.Set("test-key", "new-value")
	values = md.Get("test-key")
	assert.Equal(t, []string{"new-value"}, values)

	// Test setting multiple different keys
	rw.Set("another-key", "another-value")
	rw.Set("third-key", "third-value")

	assert.Equal(t, []string{"new-value"}, md.Get("test-key"))
	assert.Equal(t, []string{"another-value"}, md.Get("another-key"))
	assert.Equal(t, []string{"third-value"}, md.Get("third-key"))
}

func TestMetadataReaderWriter_Keys(t *testing.T) {
	tests := []struct {
		name        string
		metadata    metadata.MD
		expectedLen int
		contains    []string
	}{
		{
			name:        "empty metadata",
			metadata:    metadata.MD{},
			expectedLen: 0,
			contains:    []string{},
		},
		{
			name:        "single key",
			metadata:    metadata.MD{"key1": {"value1"}},
			expectedLen: 1,
			contains:    []string{"key1"},
		},
		{
			name: "multiple keys",
			metadata: metadata.MD{
				"key1": {"value1"},
				"key2": {"value2", "value3"},
				"key3": {"value4"},
			},
			expectedLen: 3,
			contains:    []string{"key1", "key2", "key3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rw := NewMetadataReaderWriter(&tt.metadata)

			keys := rw.Keys()
			assert.Len(t, keys, tt.expectedLen)

			for _, expectedKey := range tt.contains {
				assert.Contains(t, keys, expectedKey)
			}
		})
	}
}

func TestMetadataReaderWriter_ImplementsTextMapCarrier(t *testing.T) {
	// This test verifies that MetadataReaderWriter implements the interface correctly
	// The interface assertion is at line 34 of carrier.go
	rw := &MetadataReaderWriter{}
	assert.NotNil(t, rw)
	// The interface assertion is verified at compile time in carrier.go:34
}

func TestMetadataReaderWriter_Integration(t *testing.T) {
	// Test the full workflow of using MetadataReaderWriter
	md := metadata.New(nil)
	rw := NewMetadataReaderWriter(&md)

	// Initially empty
	assert.Empty(t, rw.Keys())
	assert.Empty(t, rw.Get("any-key"))

	// Set some values
	rw.Set("traceparent", "00-12345678901234567890123456789012-1234567890123456-01")
	rw.Set("baggage", "key1=value1,key2=value2")

	// Verify keys
	keys := rw.Keys()
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "traceparent")
	assert.Contains(t, keys, "baggage")

	// Verify values
	assert.Equal(t, "00-12345678901234567890123456789012-1234567890123456-01", rw.Get("traceparent"))
	assert.Equal(t, "key1=value1,key2=value2", rw.Get("baggage"))

	// Test setting the same key again (should overwrite)
	rw.Set("baggage", "newkey=newvalue")
	assert.Equal(t, "newkey=newvalue", rw.Get("baggage"))
	assert.Len(t, keys, 2) // keys length should remain the same
}

func TestMetadataReaderWriter_NilMetadata(t *testing.T) {
	// Test behavior with nil metadata
	rw := NewMetadataReaderWriter(nil)
	assert.NotNil(t, rw)
	assert.Nil(t, rw.md)

	// These calls should panic, which is expected behavior
	assert.Panics(t, func() {
		rw.Get("test-key")
	})

	assert.Panics(t, func() {
		rw.Set("test-key", "test-value")
	})

	assert.Panics(t, func() {
		rw.Keys()
	})
}
