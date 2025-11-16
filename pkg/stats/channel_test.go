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

package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test ChanTagInfoBase functionality
func TestChanTagInfoBase(t *testing.T) {
	tests := []struct {
		name           string
		remoteEndpoint string
		localEndpoint  string
		protocol       string
	}{
		{
			name:           "all fields set",
			remoteEndpoint: "localhost:8080",
			localEndpoint:  "localhost:9090",
			protocol:       "grpc",
		},
		{
			name:           "empty remote endpoint",
			remoteEndpoint: "",
			localEndpoint:  "localhost:9090",
			protocol:       "http",
		},
		{
			name:           "empty local endpoint",
			remoteEndpoint: "remote.example.com:80",
			localEndpoint:  "",
			protocol:       "grpc",
		},
		{
			name:           "empty protocol",
			remoteEndpoint: "localhost:8080",
			localEndpoint:  "localhost:9090",
			protocol:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagInfo := &ChanTagInfoBase{
				RemoteEndpoint: tt.remoteEndpoint,
				LocalEndpoint:  tt.localEndpoint,
				Protocol:       tt.protocol,
			}

			assert.Equal(t, tt.remoteEndpoint, tagInfo.GetRemoteEndpoint())
			assert.Equal(t, tt.localEndpoint, tagInfo.GetLocalEndpoint())
			assert.Equal(t, tt.protocol, tagInfo.GetProtocol())
		})
	}
}

// Test ChanTagInfo interface compliance
func TestChanTagInfoInterface(t *testing.T) {
	// Test that ChanTagInfoBase implements ChanTagInfo
	var _ ChanTagInfo = (*ChanTagInfoBase)(nil)

	tagInfo := &ChanTagInfoBase{
		RemoteEndpoint: "localhost:8080",
		LocalEndpoint:  "localhost:9090",
		Protocol:       "grpc",
	}

	// Verify all interface methods work
	assert.Equal(t, "localhost:8080", tagInfo.GetRemoteEndpoint())
	assert.Equal(t, "localhost:9090", tagInfo.GetLocalEndpoint())
	assert.Equal(t, "grpc", tagInfo.GetProtocol())

	// Verify isChanTagInfo method exists (not part of interface, but internal marker)
	assert.NotPanics(t, func() {
		tagInfo.isChanTagInfo()
	})
}

// Test ChanBeginBase functionality
func TestChanBeginBase(t *testing.T) {
	tests := []struct {
		name   string
		client bool
	}{
		{
			name:   "client channel begin",
			client: true,
		},
		{
			name:   "server channel begin",
			client: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			begin := &ChanBeginBase{
				Client: tt.client,
			}

			assert.Equal(t, tt.client, begin.IsClient())

			// Verify interface methods
			assert.NotPanics(t, func() {
				begin.isChanStats()
			})
			assert.NotPanics(t, func() {
				begin.isBegin()
			})
		})
	}
}

// Test ChanBegin interface compliance
func TestChanBeginInterface(t *testing.T) {
	// Test that ChanBeginBase implements ChanBegin
	var _ ChanBegin = (*ChanBeginBase)(nil)

	begin := &ChanBeginBase{
		Client: true,
	}

	// Verify all interface methods work
	assert.True(t, begin.IsClient())
	assert.NotPanics(t, func() {
		begin.isChanStats()
	})
	assert.NotPanics(t, func() {
		begin.isBegin()
	})
}

// Test ChanEndBase functionality
func TestChanEndBase(t *testing.T) {
	tests := []struct {
		name   string
		client bool
	}{
		{
			name:   "client channel end",
			client: true,
		},
		{
			name:   "server channel end",
			client: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			end := &ChanEndBase{
				Client: tt.client,
			}

			assert.Equal(t, tt.client, end.IsClient())

			// Verify interface methods
			assert.NotPanics(t, func() {
				end.isChanStats()
			})
			assert.NotPanics(t, func() {
				end.isEnd()
			})
		})
	}
}

// Test ChanEnd interface compliance
func TestChanEndInterface(t *testing.T) {
	// Test that ChanEndBase implements ChanEnd
	var _ ChanEnd = (*ChanEndBase)(nil)

	end := &ChanEndBase{
		Client: false,
	}

	// Verify all interface methods work
	assert.False(t, end.IsClient())
	assert.NotPanics(t, func() {
		end.isChanStats()
	})
	assert.NotPanics(t, func() {
		end.isEnd()
	})
}

// Test ChanStats interface compliance
func TestChanStatsInterface(t *testing.T) {
	// Test that both ChanBeginBase and ChanEndBase implement ChanStats
	var _ ChanStats = (*ChanBeginBase)(nil)
	var _ ChanStats = (*ChanEndBase)(nil)

	begin := &ChanBeginBase{Client: true}
	end := &ChanEndBase{Client: false}

	// Verify IsClient method works for both
	assert.True(t, begin.IsClient())
	assert.False(t, end.IsClient())

	// Verify marker methods
	assert.NotPanics(t, func() {
		begin.isChanStats()
		begin.isBegin()
	})
	assert.NotPanics(t, func() {
		end.isChanStats()
		end.isEnd()
	})
}

// Test custom implementations
func TestCustomChanTagInfo(t *testing.T) {
	// Test a custom implementation of ChanTagInfo
	customTagInfo := struct {
		ChanTagInfoBase
		customField string
	}{
		ChanTagInfoBase: ChanTagInfoBase{
			RemoteEndpoint: "custom.remote:8080",
			LocalEndpoint:  "custom.local:9090",
			Protocol:       "custom-protocol",
		},
		customField: "custom-value",
	}

	// Should still implement ChanTagInfo interface
	var tagInfo ChanTagInfo = &customTagInfo

	assert.Equal(t, "custom.remote:8080", tagInfo.GetRemoteEndpoint())
	assert.Equal(t, "custom.local:9090", tagInfo.GetLocalEndpoint())
	assert.Equal(t, "custom-protocol", tagInfo.GetProtocol())
}

// Test custom implementations for ChanBegin
func TestCustomChanBegin(t *testing.T) {
	// Test a custom implementation of ChanBegin
	customBegin := struct {
		ChanBeginBase
		customField string
	}{
		ChanBeginBase: ChanBeginBase{Client: true},
		customField:   "custom-value",
	}

	// Should still implement ChanBegin interface
	var begin ChanBegin = &customBegin

	assert.True(t, begin.IsClient())
}

// Test custom implementations for ChanEnd
func TestCustomChanEnd(t *testing.T) {
	// Test a custom implementation of ChanEnd
	customEnd := struct {
		ChanEndBase
		customField string
	}{
		ChanEndBase: ChanEndBase{Client: false},
		customField: "custom-value",
	}

	// Should still implement ChanEnd interface
	var end ChanEnd = &customEnd

	assert.False(t, end.IsClient())
}

// Test nil values
func TestNilValues(t *testing.T) {
	// Test that implementations can handle being instantiated with zero values
	begin := &ChanBeginBase{}
	assert.False(t, begin.IsClient())

	end := &ChanEndBase{}
	assert.False(t, end.IsClient())

	tagInfo := &ChanTagInfoBase{}
	assert.Empty(t, tagInfo.GetRemoteEndpoint())
	assert.Empty(t, tagInfo.GetLocalEndpoint())
	assert.Empty(t, tagInfo.GetProtocol())
}

// Test interface nil compliance
func TestInterfaceNilCompliance(t *testing.T) {
	// These should not panic when called on nil pointers
	var tagInfo ChanTagInfo = nil
	var begin ChanBegin = nil
	var end ChanEnd = nil
	var stats ChanStats = nil

	// We can't call methods on nil interfaces without panicking,
	// but the interface type itself should be assignable
	var _ ChanTagInfo = tagInfo
	var _ ChanBegin = begin
	var _ ChanEnd = end
	var _ ChanStats = stats
}
