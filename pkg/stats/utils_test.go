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
	"time"

	"github.com/stretchr/testify/assert"
)

// Test interface nil assignments
func TestInterfaceNilAssignments(t *testing.T) {
	// These should not panic when assigned nil values
	var tagInfo RPCTagInfo = nil
	var rpcStats RPCStats = nil
	var chanTagInfo ChanTagInfo = nil
	var chanStats ChanStats = nil

	// The interface type itself should be assignable
	var _ RPCTagInfo = tagInfo
	var _ RPCStats = rpcStats
	var _ ChanTagInfo = chanTagInfo
	var _ ChanStats = chanStats

	// Should not panic when attempting to check interface compliance with nil
	assert.True(t, true) // Dummy assertion to indicate test completion
}

// Test concrete type instantiations
func TestConcreteTypeInstantiations(t *testing.T) {
	// All these should instantiate without panic
	assert.NotPanics(t, func() {
		_ = &RPCTagInfoBase{}
	})
	assert.NotPanics(t, func() {
		_ = &RPCBeginBase{}
	})
	assert.NotPanics(t, func() {
		_ = &RPCInPayloadBase{}
	})
	assert.NotPanics(t, func() {
		_ = &RPCInHeaderBase{}
	})
	assert.NotPanics(t, func() {
		_ = &RPCClientInHeaderBase{}
	})
	assert.NotPanics(t, func() {
		_ = &RPCServerInHeaderBase{}
	})
	assert.NotPanics(t, func() {
		_ = &RPCInTrailerBase{}
	})
	assert.NotPanics(t, func() {
		_ = &RPCOutPayloadBase{}
	})
	assert.NotPanics(t, func() {
		_ = &OutHeaderBase{}
	})
	assert.NotPanics(t, func() {
		_ = &OutTrailerBase{}
	})
	assert.NotPanics(t, func() {
		_ = &RPCEndBase{}
	})
	assert.NotPanics(t, func() {
		_ = &ChanTagInfoBase{}
	})
	assert.NotPanics(t, func() {
		_ = &ChanBeginBase{}
	})
	assert.NotPanics(t, func() {
		_ = &ChanEndBase{}
	})
	assert.NotPanics(t, func() {
		_ = &handlerChain{}
	})
}

// Test embedded struct functionality
func TestEmbeddedStructs(t *testing.T) {
	// Test that embedded structs work correctly
	clientInHeader := &RPCClientInHeaderBase{
		RPCInHeaderBase: RPCInHeaderBase{
			Header:        nil,
			Protocol:      "grpc",
			TransportSize: 128,
		},
	}

	serverInHeader := &RPCServerInHeaderBase{
		RPCInHeaderBase: RPCInHeaderBase{
			Header:        nil,
			Protocol:      "grpc",
			TransportSize: 256,
		},
		FullMethod:     "/test.service/method",
		RemoteEndpoint: "localhost:8080",
		LocalEndpoint:  "localhost:9090",
	}

	// Client header should implement RPCInHeader
	var _ RPCInHeader = clientInHeader
	assert.Equal(t, "grpc", clientInHeader.GetProtocol())
	assert.Equal(t, 128, clientInHeader.GetTransportSize())

	// Server header should implement RPCServerInHeader
	var _ RPCServerInHeader = serverInHeader
	assert.Equal(t, "grpc", serverInHeader.GetProtocol())
	assert.Equal(t, 256, serverInHeader.GetTransportSize())
	assert.Equal(t, "/test.service/method", serverInHeader.GetFullMethod())
	assert.Equal(t, "localhost:8080", serverInHeader.GetRemoteEndpoint())
	assert.Equal(t, "localhost:9090", serverInHeader.GetLocalEndpoint())
}

// Test interface implementation consistency
func TestInterfaceImplementationConsistency(t *testing.T) {
	// Create instances of each type and verify interface compliance
	now := time.Now()
	begin := &RPCBeginBase{
		Client:       true,
		BeginTime:    now,
		ClientStream: false,
		ServerStream: false,
		Protocol:     "grpc",
	}

	// Should implement both RPCBegin and RPCStats
	var _ RPCBegin = begin
	var _ RPCStats = begin
	assert.True(t, begin.IsClient())
	assert.NotPanics(t, func() { begin.isRPCStats() })

	// Test other interfaces similarly
	end := &RPCEndBase{
		Client:    true,
		BeginTime: now,
		EndTime:   now.Add(1),
		Err:       nil,
		Protocol:  "grpc",
	}

	var _ RPCEnd = end
	var _ RPCStats = end
	assert.True(t, end.IsClient())
	assert.NotPanics(t, func() { end.isRPCStats() })
}

// Helper function for consistent time generation in tests
func testTime() interface{} {
	// This would normally return time.Now() in actual tests
	// For consistency, we'll use a fixed time for predictable testing
	return testFixedTime
}

func testFixedTime() interface{} {
	// Return a consistent time for testing
	return "fixed-time-value"
}
