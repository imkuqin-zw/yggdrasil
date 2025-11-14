package metadata

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMD_New(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected MD
	}{
		{
			name:     "empty map",
			input:    map[string]string{},
			expected: MD{},
		},
		{
			name: "single key-value pair",
			input: map[string]string{
				"key1": "value1",
			},
			expected: MD{
				"key1": []string{"value1"},
			},
		},
		{
			name: "multiple key-value pairs",
			input: map[string]string{
				"Key1": "value1",
				"KEY2": "value2",
				"key3": "value3",
			},
			expected: MD{
				"key1": []string{"value1"},
				"key2": []string{"value2"},
				"key3": []string{"value3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := New(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMD_Pairs(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected MD
	}{
		{
			name:     "empty pairs",
			input:    []string{},
			expected: MD{},
		},
		{
			name:     "single pair",
			input:    []string{"key1", "value1"},
			expected: MD{"key1": []string{"value1"}},
		},
		{
			name: "multiple pairs",
			input: []string{
				"Key1", "value1",
				"KEY2", "value2",
				"key3", "value3",
			},
			expected: MD{
				"key1": []string{"value1"},
				"key2": []string{"value2"},
				"key3": []string{"value3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Pairs(tt.input...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMD_Pairs_Panic(t *testing.T) {
	assert.Panics(t, func() {
		Pairs("key1", "value1", "key2") // Odd number of arguments
	})
}

func TestMD_Len(t *testing.T) {
	tests := []struct {
		name     string
		md       MD
		expected int
	}{
		{
			name:     "empty metadata",
			md:       MD{},
			expected: 0,
		},
		{
			name: "single key",
			md: MD{
				"key1": []string{"value1"},
			},
			expected: 1,
		},
		{
			name: "multiple keys",
			md: MD{
				"key1": []string{"value1"},
				"key2": []string{"value2", "value3"},
				"key3": []string{"value4"},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.md.Len())
		})
	}
}

func TestMD_Get(t *testing.T) {
	md := MD{
		"key1": []string{"value1"},
		"key2": []string{"value2", "value3"},
		"key3": []string{"value4"}, // Should be accessible with lowercase
	}

	tests := []struct {
		name     string
		key      string
		expected []string
	}{
		{
			name:     "existing key lowercase",
			key:      "key1",
			expected: []string{"value1"},
		},
		{
			name:     "existing key uppercase",
			key:      "KEY1",
			expected: []string{"value1"},
		},
		{
			name:     "existing key with multiple values",
			key:      "key2",
			expected: []string{"value2", "value3"},
		},
		{
			name:     "existing key mixed case",
			key:      "key3",
			expected: []string{"value4"},
		},
		{
			name:     "non-existing key",
			key:      "nonexistent",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := md.Get(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMD_Set(t *testing.T) {
	md := MD{
		"key1": []string{"value1"},
	}

	tests := []struct {
		name     string
		key      string
		values   []string
		expected MD
	}{
		{
			name:   "set existing key",
			key:    "key1",
			values: []string{"newvalue1", "newvalue2"},
			expected: MD{
				"key1": []string{"newvalue1", "newvalue2"},
			},
		},
		{
			name:   "set new key",
			key:    "key2",
			values: []string{"value2"},
			expected: MD{
				"key1": []string{"value1"},
				"key2": []string{"value2"},
			},
		},
		{
			name:   "set with empty values",
			key:    "key3",
			values: []string{},
			expected: MD{
				"key1": []string{"value1"},
			},
		},
		{
			name:   "set with mixed case key",
			key:    "KEY2",
			values: []string{"value2"},
			expected: MD{
				"key1": []string{"value1"},
				"key2": []string{"value2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mdCopy := md.Copy()
			mdCopy.Set(tt.key, tt.values...)
			assert.Equal(t, tt.expected, mdCopy)
		})
	}
}

func TestMD_Append(t *testing.T) {
	md := MD{
		"key1": []string{"value1"},
	}

	tests := []struct {
		name     string
		key      string
		values   []string
		expected MD
	}{
		{
			name:   "append to existing key",
			key:    "key1",
			values: []string{"value2", "value3"},
			expected: MD{
				"key1": []string{"value1", "value2", "value3"},
			},
		},
		{
			name:   "append to new key",
			key:    "key2",
			values: []string{"value2"},
			expected: MD{
				"key1": []string{"value1"},
				"key2": []string{"value2"},
			},
		},
		{
			name:   "append with empty values",
			key:    "key3",
			values: []string{},
			expected: MD{
				"key1": []string{"value1"},
			},
		},
		{
			name:   "append with mixed case key",
			key:    "KEY1",
			values: []string{"value2"},
			expected: MD{
				"key1": []string{"value1", "value2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mdCopy := md.Copy()
			mdCopy.Append(tt.key, tt.values...)
			assert.Equal(t, tt.expected, mdCopy)
		})
	}
}

func TestMD_Copy(t *testing.T) {
	original := MD{
		"key1": []string{"value1"},
		"key2": []string{"value2", "value3"},
	}

	copy := original.Copy()
	assert.Equal(t, original, copy)

	// Modify the copy
	copy.Set("key1", "newvalue1")

	// Original should be unchanged
	assert.Equal(t, []string{"value1"}, original.Get("key1"))
	assert.Equal(t, []string{"newvalue1"}, copy.Get("key1"))
}

func TestMD_Join(t *testing.T) {
	tests := []struct {
		name     string
		mds      []MD
		expected MD
	}{
		{
			name:     "no metadata",
			mds:      []MD{},
			expected: MD{},
		},
		{
			name: "single metadata",
			mds: []MD{
				{"key1": []string{"value1"}},
			},
			expected: MD{
				"key1": []string{"value1"},
			},
		},
		{
			name: "multiple metadata with different keys",
			mds: []MD{
				{"key1": []string{"value1"}},
				{"key2": []string{"value2"}},
				{"key3": []string{"value3"}},
			},
			expected: MD{
				"key1": []string{"value1"},
				"key2": []string{"value2"},
				"key3": []string{"value3"},
			},
		},
		{
			name: "multiple metadata with overlapping keys",
			mds: []MD{
				{"key1": []string{"value1"}},
				{"key1": []string{"value2"}},
				{"key2": []string{"value3"}},
				{"key1": []string{"value4"}},
			},
			expected: MD{
				"key1": []string{"value1", "value2", "value4"},
				"key2": []string{"value3"},
			},
		},
		{
			name: "metadata with empty values",
			mds: []MD{
				{},
				{"key1": []string{"value1"}},
				{},
			},
			expected: MD{
				"key1": []string{"value1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Join(tt.mds...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContext_WithInContext(t *testing.T) {
	ctx := context.Background()
	md := MD{"key1": []string{"value1"}}

	// Test adding metadata to empty context
	newCtx := WithInContext(ctx, md)
	retrievedMd, ok := FromInContext(newCtx)
	require.True(t, ok)
	assert.Equal(t, md, retrievedMd)

	// Test adding metadata to context with existing metadata
	md2 := MD{"key2": []string{"value2"}}
	ctxWithMd := WithInContext(newCtx, md2)
	retrievedMd2, ok := FromInContext(ctxWithMd)
	require.True(t, ok)
	expected := Join(md, md2)
	assert.Equal(t, expected, retrievedMd2)
}

func TestContext_FromInContext(t *testing.T) {
	ctx := context.Background()

	// Test from context without metadata
	_, ok := FromInContext(ctx)
	assert.False(t, ok)

	// Test from context with metadata
	md := MD{"key1": []string{"value1"}}
	ctxWithMd := context.WithValue(ctx, inKey{}, md)
	retrievedMd, ok := FromInContext(ctxWithMd)
	require.True(t, ok)
	assert.Equal(t, md, retrievedMd)

	// Test that returned metadata is a copy
	retrievedMd.Set("key1", "newvalue1")
	originalMd, _ := FromInContext(ctxWithMd)
	assert.Equal(t, []string{"value1"}, originalMd.Get("key1"))
}

func TestContext_WithOutContext(t *testing.T) {
	ctx := context.Background()
	md := MD{"key1": []string{"value1"}}

	// Test adding metadata to empty context
	newCtx := WithOutContext(ctx, md)
	retrievedMd, ok := FromOutContext(newCtx)
	require.True(t, ok)
	assert.Equal(t, md, retrievedMd)

	// Test adding metadata to context with existing metadata
	md2 := MD{"key2": []string{"value2"}}
	ctxWithMd := WithOutContext(newCtx, md2)
	retrievedMd2, ok := FromOutContext(ctxWithMd)
	require.True(t, ok)
	expected := Join(md, md2)
	assert.Equal(t, expected, retrievedMd2)
}

func TestContext_FromOutContext(t *testing.T) {
	ctx := context.Background()

	// Test from context without metadata
	_, ok := FromOutContext(ctx)
	assert.False(t, ok)

	// Test from context with metadata
	md := MD{"key1": []string{"value1"}}
	ctxWithMd := context.WithValue(ctx, outKey{}, md)
	retrievedMd, ok := FromOutContext(ctxWithMd)
	require.True(t, ok)
	assert.Equal(t, md, retrievedMd)
}

func TestContext_WithStreamContext(t *testing.T) {
	ctx := context.Background()

	// Test creating stream context
	streamCtx := WithStreamContext(ctx)
	_, ok := streamCtx.Value(streamKey{}).(*stream)
	assert.True(t, ok)

	// Test getting existing stream context
	streamCtx2 := WithStreamContext(streamCtx)
	_, ok2 := streamCtx2.Value(streamKey{}).(*stream)
	assert.True(t, ok2)

	// Should be the same stream object
	stream1, _ := streamCtx.Value(streamKey{}).(*stream)
	stream2, _ := streamCtx2.Value(streamKey{}).(*stream)
	assert.Same(t, stream1, stream2)
}

func TestContext_SetHeader(t *testing.T) {
	ctx := WithStreamContext(context.Background())
	md := MD{"key1": []string{"value1"}}

	// Test setting header
	err := SetHeader(ctx, md)
	assert.NoError(t, err)

	// Test retrieving header
	retrievedMd, ok := FromHeaderCtx(ctx)
	require.True(t, ok)
	assert.Equal(t, md, retrievedMd)

	// Test setting additional header
	md2 := MD{"key2": []string{"value2"}}
	err = SetHeader(ctx, md2)
	assert.NoError(t, err)

	retrievedMd2, ok := FromHeaderCtx(ctx)
	require.True(t, ok)
	expected := Join(md, md2)
	assert.Equal(t, expected, retrievedMd2)
}

func TestContext_SetHeader_Error(t *testing.T) {
	ctx := context.Background()
	md := MD{"key1": []string{"value1"}}

	// Test setting header on context without stream
	err := SetHeader(ctx, md)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch the stream")
}

func TestContext_FromHeaderCtx(t *testing.T) {
	ctx := context.Background()

	// Test from context without stream
	_, ok := FromHeaderCtx(ctx)
	assert.False(t, ok)

	// Test from context with stream but no header
	streamCtx := WithStreamContext(ctx)
	_, ok = FromHeaderCtx(streamCtx)
	assert.False(t, ok)

	// Test from context with header
	md := MD{"key1": []string{"value1"}}
	streamValue, _ := streamCtx.Value(streamKey{}).(*stream)
	streamValue.header = md
	streamCtxWithHeader := context.WithValue(streamCtx, streamKey{}, streamValue)

	retrievedMd, ok := FromHeaderCtx(streamCtxWithHeader)
	require.True(t, ok)
	assert.Equal(t, md, retrievedMd)
}

func TestContext_SetTrailer(t *testing.T) {
	ctx := WithStreamContext(context.Background())
	md := MD{"key1": []string{"value1"}}

	// Test setting trailer
	err := SetTrailer(ctx, md)
	assert.NoError(t, err)

	// Test retrieving trailer
	retrievedMd, ok := FromTrailerCtx(ctx)
	require.True(t, ok)
	assert.Equal(t, md, retrievedMd)

	// Test setting additional trailer
	md2 := MD{"key2": []string{"value2"}}
	err = SetTrailer(ctx, md2)
	assert.NoError(t, err)

	retrievedMd2, ok := FromTrailerCtx(ctx)
	require.True(t, ok)
	expected := Join(md, md2)
	assert.Equal(t, expected, retrievedMd2)
}

func TestContext_SetTrailer_Error(t *testing.T) {
	ctx := context.Background()
	md := MD{"key1": []string{"value1"}}

	// Test setting trailer on context without stream
	err := SetTrailer(ctx, md)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch the stream")
}

func TestContext_FromTrailerCtx(t *testing.T) {
	ctx := context.Background()

	// Test from context without stream
	_, ok := FromTrailerCtx(ctx)
	assert.False(t, ok)

	// Test from context with stream but no trailer
	streamCtx := WithStreamContext(ctx)
	_, ok = FromTrailerCtx(streamCtx)
	assert.False(t, ok)

	// Test from context with trailer
	md := MD{"key1": []string{"value1"}}
	streamValue, _ := streamCtx.Value(streamKey{}).(*stream)
	streamValue.trailer = md
	streamCtxWithTrailer := context.WithValue(streamCtx, streamKey{}, streamValue)

	retrievedMd, ok := FromTrailerCtx(streamCtxWithTrailer)
	require.True(t, ok)
	assert.Equal(t, md, retrievedMd)
}

func TestContext_StreamConcurrency(t *testing.T) {
	ctx := WithStreamContext(context.Background())
	md1 := MD{"key1": []string{"value1"}}
	md2 := MD{"key2": []string{"value2"}}

	// Test concurrent header and trailer setting
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			_ = SetHeader(ctx, md1)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			_ = SetTrailer(ctx, md2)
		}
		done <- true
	}()

	<-done
	<-done

	// Verify both header and trailer are set
	header, ok := FromHeaderCtx(ctx)
	require.True(t, ok)
	assert.NotEmpty(t, header)

	trailer, ok := FromTrailerCtx(ctx)
	require.True(t, ok)
	assert.NotEmpty(t, trailer)
}
