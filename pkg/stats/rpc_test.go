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

	"github.com/imkuqin-zw/yggdrasil/pkg/metadata"
	"github.com/stretchr/testify/assert"
)

// Test RPCTagInfoBase functionality
func TestRPCTagInfoBase(t *testing.T) {
	tests := []struct {
		name       string
		fullMethod string
	}{
		{
			name:       "valid full method",
			fullMethod: "/package.service/method",
		},
		{
			name:       "empty full method",
			fullMethod: "",
		},
		{
			name:       "method with multiple slashes",
			fullMethod: "/v1/package.service/method",
		},
		{
			name:       "method without leading slash",
			fullMethod: "package.service/method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagInfo := &RPCTagInfoBase{
				FullMethod: tt.fullMethod,
			}

			assert.Equal(t, tt.fullMethod, tagInfo.GetFullMethod())

			// Verify marker method
			assert.NotPanics(t, func() {
				tagInfo.isRPCTagInfo()
			})
		})
	}
}

// Test RPCTagInfo interface compliance
func TestRPCTagInfoInterface(t *testing.T) {
	// Test that RPCTagInfoBase implements RPCTagInfo
	var _ RPCTagInfo = (*RPCTagInfoBase)(nil)

	tagInfo := &RPCTagInfoBase{
		FullMethod: "/test.service/method",
	}

	assert.Equal(t, "/test.service/method", tagInfo.GetFullMethod())
	assert.NotPanics(t, func() {
		tagInfo.isRPCTagInfo()
	})
}

// Test RPCBeginBase functionality
func TestRPCBeginBase(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name         string
		client       bool
		beginTime    time.Time
		clientStream bool
		serverStream bool
		protocol     string
	}{
		{
			name:         "client unary RPC",
			client:       true,
			beginTime:    now,
			clientStream: false,
			serverStream: false,
			protocol:     "grpc",
		},
		{
			name:         "server streaming RPC",
			client:       false,
			beginTime:    now,
			clientStream: false,
			serverStream: true,
			protocol:     "grpc",
		},
		{
			name:         "client streaming RPC",
			client:       true,
			beginTime:    now,
			clientStream: true,
			serverStream: false,
			protocol:     "http",
		},
		{
			name:         "bidirectional streaming RPC",
			client:       true,
			beginTime:    now,
			clientStream: true,
			serverStream: true,
			protocol:     "grpc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			begin := &RPCBeginBase{
				Client:       tt.client,
				BeginTime:    tt.beginTime,
				ClientStream: tt.clientStream,
				ServerStream: tt.serverStream,
				Protocol:     tt.protocol,
			}

			assert.Equal(t, tt.client, begin.IsClient())
			assert.Equal(t, tt.beginTime, begin.GetBeginTime())
			assert.Equal(t, tt.clientStream, begin.IsClientStream())
			assert.Equal(t, tt.serverStream, begin.IsServerStream())
			assert.Equal(t, tt.protocol, begin.GetProtocol())

			// Verify interface methods
			assert.NotPanics(t, func() {
				begin.isRPCStats()
			})
		})
	}
}

// Test RPCBegin interface compliance
func TestRPCBeginInterface(t *testing.T) {
	// Test that RPCBeginBase implements RPCBegin
	var _ RPCBegin = (*RPCBeginBase)(nil)

	begin := &RPCBeginBase{
		Client:       true,
		BeginTime:    time.Now(),
		ClientStream: false,
		ServerStream: false,
		Protocol:     "grpc",
	}

	// Verify all interface methods work
	assert.True(t, begin.IsClient())
	assert.NotPanics(t, func() {
		begin.isRPCStats()
	})
}

// Test RPCInPayloadBase functionality
func TestRPCInPayloadBase(t *testing.T) {
	now := time.Now()
	payload := struct{ Field string }{"test"}
	data := []byte(`{"field":"test"}`)
	tests := []struct {
		name          string
		client        bool
		payload       interface{}
		data          []byte
		transportSize int
		recvTime      time.Time
		protocol      string
	}{
		{
			name:          "client inbound payload",
			client:        true,
			payload:       payload,
			data:          data,
			transportSize: 128,
			recvTime:      now,
			protocol:      "grpc",
		},
		{
			name:          "server inbound payload",
			client:        false,
			payload:       payload,
			data:          data,
			transportSize: 256,
			recvTime:      now,
			protocol:      "grpc",
		},
		{
			name:          "empty payload",
			client:        true,
			payload:       nil,
			data:          []byte{},
			transportSize: 0,
			recvTime:      now,
			protocol:      "http",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inPayload := &RPCInPayloadBase{
				Client:        tt.client,
				Payload:       tt.payload,
				Data:          tt.data,
				TransportSize: tt.transportSize,
				RecvTime:      tt.recvTime,
				Protocol:      tt.protocol,
			}

			assert.Equal(t, tt.client, inPayload.IsClient())
			assert.Equal(t, tt.payload, inPayload.GetPayload())
			assert.Equal(t, tt.data, inPayload.GetData())
			assert.Equal(t, tt.transportSize, inPayload.GetTransportSize())
			assert.Equal(t, tt.recvTime, inPayload.GetRecvTime())
			assert.Equal(t, tt.protocol, inPayload.GetProtocol())

			// Verify interface methods
			assert.NotPanics(t, func() {
				inPayload.isRPCStats()
			})
		})
	}
}

// Test RPCInPayload interface compliance
func TestRPCInPayloadInterface(t *testing.T) {
	// Test that RPCInPayloadBase implements RPCInPayload
	var _ RPCInPayload = (*RPCInPayloadBase)(nil)

	now := time.Now()
	payload := struct{ Field string }{"test"}
	data := []byte(`{"field":"test"}`)

	inPayload := &RPCInPayloadBase{
		Client:        true,
		Payload:       payload,
		Data:          data,
		TransportSize: 128,
		RecvTime:      now,
		Protocol:      "grpc",
	}

	// Verify all interface methods work
	assert.True(t, inPayload.IsClient())
	assert.Equal(t, payload, inPayload.GetPayload())
	assert.Equal(t, data, inPayload.GetData())
	assert.Equal(t, 128, inPayload.GetTransportSize())
	assert.Equal(t, now, inPayload.GetRecvTime())
	assert.Equal(t, "grpc", inPayload.GetProtocol())
	assert.NotPanics(t, func() {
		inPayload.isRPCStats()
	})
}

// Test RPCInHeaderBase functionality
func TestRPCInHeaderBase(t *testing.T) {
	md := metadata.MD{"key1": []string{"value1"}, "key2": []string{"value2"}}
	tests := []struct {
		name          string
		header        metadata.MD
		protocol      string
		transportSize int
	}{
		{
			name:          "header with metadata",
			header:        md,
			protocol:      "grpc",
			transportSize: 256,
		},
		{
			name:          "empty header",
			header:        metadata.MD{},
			protocol:      "http",
			transportSize: 0,
		},
		{
			name:          "nil header",
			header:        nil,
			protocol:      "grpc",
			transportSize: 128,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inHeader := &RPCInHeaderBase{
				Header:        tt.header,
				Protocol:      tt.protocol,
				TransportSize: tt.transportSize,
			}

			assert.Equal(t, tt.header, inHeader.GetHeader())
			assert.Equal(t, tt.protocol, inHeader.GetProtocol())
			assert.Equal(t, tt.transportSize, inHeader.GetTransportSize())

			// Verify interface methods
			assert.NotPanics(t, func() {
				inHeader.isRPCStats()
			})
		})
	}
}

// Test RPCInHeader interface compliance
func TestRPCInHeaderInterface(t *testing.T) {
	// Test that RPCInHeaderBase implements RPCInHeader
	var _ RPCInHeader = (*RPCInHeaderBase)(nil)

	md := metadata.MD{"key": []string{"value"}}
	inHeader := &RPCInHeaderBase{
		Header:        md,
		Protocol:      "grpc",
		TransportSize: 128,
	}

	assert.Equal(t, md, inHeader.GetHeader())
	assert.Equal(t, "grpc", inHeader.GetProtocol())
	assert.Equal(t, 128, inHeader.GetTransportSize())
	assert.NotPanics(t, func() {
		inHeader.isRPCStats()
	})
}

// Test RPCClientInHeaderBase functionality
func TestRPCClientInHeaderBase(t *testing.T) {
	baseHeader := &RPCInHeaderBase{
		Header:        metadata.MD{"client": []string{"header"}},
		Protocol:      "grpc",
		TransportSize: 64,
	}

	clientHeader := &RPCClientInHeaderBase{RPCInHeaderBase: *baseHeader}

	// Should implement all RPCInHeader methods
	assert.Equal(t, metadata.MD{"client": []string{"header"}}, clientHeader.GetHeader())
	assert.Equal(t, "grpc", clientHeader.GetProtocol())
	assert.Equal(t, 64, clientHeader.GetTransportSize())

	// Verify marker method
	assert.NotPanics(t, func() {
		clientHeader.clientInHeader()
	})
}

// Test RPCClientInHeader interface compliance
func TestRPCClientInHeaderInterface(t *testing.T) {
	// Test that RPCClientInHeaderBase implements RPCClientInHeader
	var _ RPCClientInHeader = (*RPCClientInHeaderBase)(nil)

	clientHeader := &RPCClientInHeaderBase{
		RPCInHeaderBase: RPCInHeaderBase{
			Header:        metadata.MD{},
			Protocol:      "grpc",
			TransportSize: 128,
		},
	}

	// Should implement RPCInHeader interface
	var _ RPCInHeader = clientHeader
	assert.NotPanics(t, func() {
		clientHeader.GetHeader()
	})
	assert.NotPanics(t, func() {
		clientHeader.clientInHeader()
	})
}

// Test RPCServerInHeaderBase functionality
func TestRPCServerInHeaderBase(t *testing.T) {
	baseHeader := &RPCInHeaderBase{
		Header:        metadata.MD{"server": []string{"header"}},
		Protocol:      "grpc",
		TransportSize: 64,
	}

	serverHeader := &RPCServerInHeaderBase{
		RPCInHeaderBase: *baseHeader,
		FullMethod:      "/test.service/method",
		RemoteEndpoint:  "localhost:8080",
		LocalEndpoint:   "localhost:9090",
	}

	// Should implement all RPCInHeader and RPCServerInHeader methods
	assert.Equal(t, metadata.MD{"server": []string{"header"}}, serverHeader.GetHeader())
	assert.Equal(t, "grpc", serverHeader.GetProtocol())
	assert.Equal(t, 64, serverHeader.GetTransportSize())
	assert.Equal(t, "/test.service/method", serverHeader.GetFullMethod())
	assert.Equal(t, "localhost:8080", serverHeader.GetRemoteEndpoint())
	assert.Equal(t, "localhost:9090", serverHeader.GetLocalEndpoint())

	// Verify marker methods
	assert.NotPanics(t, func() {
		serverHeader.serverInHeader()
	})
}

// Test RPCServerInHeader interface compliance
func TestRPCServerInHeaderInterface(t *testing.T) {
	// Test that RPCServerInHeaderBase implements RPCServerInHeader
	var _ RPCServerInHeader = (*RPCServerInHeaderBase)(nil)

	serverHeader := &RPCServerInHeaderBase{
		RPCInHeaderBase: RPCInHeaderBase{
			Header:        metadata.MD{},
			Protocol:      "grpc",
			TransportSize: 128,
		},
		FullMethod:     "/test.service/method",
		RemoteEndpoint: "localhost:8080",
		LocalEndpoint:  "localhost:9090",
	}

	// Should implement RPCInHeader interface
	var _ RPCInHeader = serverHeader
	assert.NotPanics(t, func() {
		serverHeader.GetHeader()
	})
	assert.NotPanics(t, func() {
		serverHeader.serverInHeader()
	})
}

// Test RPCOutPayloadBase functionality
func TestRPCOutPayloadBase(t *testing.T) {
	now := time.Now()
	payload := struct{ Field string }{"test"}
	data := []byte(`{"field":"test"}`)
	tests := []struct {
		name          string
		client        bool
		payload       interface{}
		data          []byte
		transportSize int
		sendTime      time.Time
		protocol      string
	}{
		{
			name:          "client outbound payload",
			client:        true,
			payload:       payload,
			data:          data,
			transportSize: 128,
			sendTime:      now,
			protocol:      "grpc",
		},
		{
			name:          "server outbound payload",
			client:        false,
			payload:       payload,
			data:          data,
			transportSize: 256,
			sendTime:      now,
			protocol:      "grpc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outPayload := &RPCOutPayloadBase{
				Client:        tt.client,
				Payload:       tt.payload,
				Data:          tt.data,
				TransportSize: tt.transportSize,
				SendTime:      tt.sendTime,
				Protocol:      tt.protocol,
			}

			assert.Equal(t, tt.client, outPayload.IsClient())
			assert.Equal(t, tt.payload, outPayload.GetPayload())
			assert.Equal(t, tt.data, outPayload.GetData())
			assert.Equal(t, tt.transportSize, outPayload.GetTransportSize())
			assert.Equal(t, tt.sendTime, outPayload.GetSendTime())
			assert.Equal(t, tt.protocol, outPayload.GetProtocol())

			// Verify interface methods
			assert.NotPanics(t, func() {
				outPayload.isRPCStats()
			})
		})
	}
}

// Test RPCOutPayload interface compliance
func TestRPCOutPayloadInterface(t *testing.T) {
	// Test that RPCOutPayloadBase implements RPCOutPayload
	var _ RPCOutPayload = (*RPCOutPayloadBase)(nil)

	now := time.Now()
	payload := struct{ Field string }{"test"}
	data := []byte(`{"field":"test"}`)

	outPayload := &RPCOutPayloadBase{
		Client:        true,
		Payload:       payload,
		Data:          data,
		TransportSize: 128,
		SendTime:      now,
		Protocol:      "grpc",
	}

	// Verify all interface methods work
	assert.True(t, outPayload.IsClient())
	assert.Equal(t, payload, outPayload.GetPayload())
	assert.Equal(t, data, outPayload.GetData())
	assert.Equal(t, 128, outPayload.GetTransportSize())
	assert.Equal(t, now, outPayload.GetSendTime())
	assert.Equal(t, "grpc", outPayload.GetProtocol())
	assert.NotPanics(t, func() {
		outPayload.isRPCStats()
	})
}

// Test OutHeaderBase functionality
func TestOutHeaderBase(t *testing.T) {
	md := metadata.MD{"key1": []string{"value1"}, "key2": []string{"value2"}}
	tests := []struct {
		name           string
		client         bool
		header         metadata.MD
		fullMethod     string
		remoteEndpoint string
		localEndpoint  string
		protocol       string
		transportSize  int
	}{
		{
			name:           "client header",
			client:         true,
			header:         md,
			fullMethod:     "/test.service/method",
			remoteEndpoint: "localhost:8080",
			localEndpoint:  "localhost:9090",
			protocol:       "grpc",
			transportSize:  256,
		},
		{
			name:           "server header",
			client:         false,
			header:         md,
			fullMethod:     "/test.service/method",
			remoteEndpoint: "localhost:8080",
			localEndpoint:  "localhost:9090",
			protocol:       "grpc",
			transportSize:  256,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outHeader := &OutHeaderBase{
				Client:         tt.client,
				Header:         tt.header,
				FullMethod:     tt.fullMethod,
				RemoteEndpoint: tt.remoteEndpoint,
				LocalEndpoint:  tt.localEndpoint,
				Protocol:       tt.protocol,
				TransportSize:  tt.transportSize,
			}

			assert.Equal(t, tt.client, outHeader.IsClient())
			assert.Equal(t, tt.header, outHeader.GetHeader())
			assert.Equal(t, tt.fullMethod, outHeader.GetFullMethod())
			assert.Equal(t, tt.remoteEndpoint, outHeader.GetRemoteEndpoint())
			assert.Equal(t, tt.localEndpoint, outHeader.GetLocalEndpoint())
			assert.Equal(t, tt.protocol, outHeader.GetProtocol())
			assert.Equal(t, tt.transportSize, outHeader.GetTransportSize())

			// Verify interface methods
			assert.NotPanics(t, func() {
				outHeader.isRPCStats()
			})
		})
	}
}

// Test RPCOutHeader interface compliance
func TestOutHeaderInterface(t *testing.T) {
	// Test that OutHeaderBase implements RPCOutHeader
	var _ RPCOutHeader = (*OutHeaderBase)(nil)

	md := metadata.MD{"key": []string{"value"}}
	outHeader := &OutHeaderBase{
		Client:         true,
		Header:         md,
		FullMethod:     "/test.service/method",
		RemoteEndpoint: "localhost:8080",
		LocalEndpoint:  "localhost:9090",
		Protocol:       "grpc",
		TransportSize:  128,
	}

	assert.Equal(t, true, outHeader.IsClient())
	assert.Equal(t, md, outHeader.GetHeader())
	assert.Equal(t, "/test.service/method", outHeader.GetFullMethod())
	assert.Equal(t, "localhost:8080", outHeader.GetRemoteEndpoint())
	assert.Equal(t, "localhost:9090", outHeader.GetLocalEndpoint())
	assert.Equal(t, "grpc", outHeader.GetProtocol())
	assert.Equal(t, 128, outHeader.GetTransportSize())
	assert.NotPanics(t, func() {
		outHeader.isRPCStats()
	})
}

// Test RPCOutTrailerBase functionality
func TestRPCOutTrailerBase(t *testing.T) {
	md := metadata.MD{"key1": []string{"value1"}}
	tests := []struct {
		name          string
		client        bool
		trailer       metadata.MD
		transportSize int
	}{
		{
			name:          "server outbound trailer",
			client:        false,
			trailer:       md,
			transportSize: 128,
		},
		{
			name:          "client outbound trailer",
			client:        true,
			trailer:       metadata.MD{},
			transportSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outTrailer := &OutTrailerBase{
				Client:        tt.client,
				Trailer:       tt.trailer,
				TransportSize: tt.transportSize,
			}

			assert.Equal(t, tt.client, outTrailer.IsClient())
			assert.Equal(t, tt.trailer, outTrailer.GetTrailer())
			assert.Equal(t, tt.transportSize, outTrailer.GetTransportSize())

			// Verify interface methods
			assert.NotPanics(t, func() {
				outTrailer.isRPCStats()
			})
		})
	}
}

// Test RPCOutTrailer interface compliance
func TestRPCOutTrailerInterface(t *testing.T) {
	// Test that OutTrailerBase implements RPCOutTrailer
	var _ RPCOutTrailer = (*OutTrailerBase)(nil)

	md := metadata.MD{"key": []string{"value"}}
	outTrailer := &OutTrailerBase{
		Client:        true,
		Trailer:       md,
		TransportSize: 64,
	}

	assert.Equal(t, true, outTrailer.IsClient())
	assert.Equal(t, md, outTrailer.GetTrailer())
	assert.Equal(t, 64, outTrailer.GetTransportSize())
	assert.NotPanics(t, func() {
		outTrailer.isRPCStats()
	})
}

// Test RPCEndBase functionality
func TestRPCEndBase(t *testing.T) {
	now := time.Now()
	err := assert.AnError
	tests := []struct {
		name      string
		client    bool
		beginTime time.Time
		endTime   time.Time
		error     error
		protocol  string
	}{
		{
			name:      "successful RPC",
			client:    true,
			beginTime: now,
			endTime:   now.Add(time.Second),
			error:     nil,
			protocol:  "grpc",
		},
		{
			name:      "failed RPC",
			client:    false,
			beginTime: now,
			endTime:   now.Add(time.Second),
			error:     err,
			protocol:  "grpc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			end := &RPCEndBase{
				Client:    tt.client,
				BeginTime: tt.beginTime,
				EndTime:   tt.endTime,
				Err:       tt.error,
				Protocol:  tt.protocol,
			}

			assert.Equal(t, tt.client, end.IsClient())
			assert.Equal(t, tt.beginTime, end.GetBeginTime())
			assert.Equal(t, tt.endTime, end.GetEndTime())
			assert.Equal(t, tt.error, end.Error())
			assert.Equal(t, tt.protocol, end.GetProtocol())

			// Verify interface methods
			assert.NotPanics(t, func() {
				end.isRPCStats()
			})
		})
	}
}

// Test RPCEnd interface compliance
func TestRPCEndInterface(t *testing.T) {
	// Test that RPCEndBase implements RPCEnd
	var _ RPCEnd = (*RPCEndBase)(nil)

	now := time.Now()
	err := assert.AnError
	end := &RPCEndBase{
		Client:    true,
		BeginTime: now,
		EndTime:   now.Add(time.Second),
		Err:       err,
		Protocol:  "grpc",
	}

	// Verify all interface methods work
	assert.True(t, end.IsClient())
	assert.Equal(t, now, end.GetBeginTime())
	assert.Equal(t, now.Add(time.Second), end.GetEndTime())
	assert.Equal(t, err, end.Error())
	assert.Equal(t, "grpc", end.GetProtocol())
	assert.NotPanics(t, func() {
		end.isRPCStats()
	})
}

// Test nil values and zero initialization
func TestNilAndZeroValues(t *testing.T) {
	// Test zero value initialization
	begin := &RPCBeginBase{}
	assert.False(t, begin.IsClient())
	assert.False(t, begin.IsClientStream())
	assert.False(t, begin.IsServerStream())
	assert.Empty(t, begin.GetProtocol())

	// Test with nil values
	inPayload := &RPCInPayloadBase{}
	assert.False(t, inPayload.IsClient())
	assert.Nil(t, inPayload.GetPayload())
	assert.Nil(t, inPayload.GetData())
	assert.Equal(t, 0, inPayload.GetTransportSize())

	outPayload := &RPCOutPayloadBase{}
	assert.False(t, outPayload.IsClient())
	assert.Nil(t, outPayload.GetPayload())
	assert.Nil(t, outPayload.GetData())
	assert.Equal(t, 0, outPayload.GetTransportSize())
}

// Test RPCStats interface compliance
func TestRPCStatsInterfaceCompliance(t *testing.T) {
	// Test that all RPC stats types implement RPCStats
	var _ RPCStats = (*RPCBeginBase)(nil)
	var _ RPCStats = (*RPCInPayloadBase)(nil)
	var _ RPCStats = (*RPCInHeaderBase)(nil)
	var _ RPCStats = (*RPCInHeaderBase)(nil) // nil interface assignment test
	var _ RPCStats = (*RPCClientInHeaderBase)(nil)
	var _ RPCStats = (*RPCServerInHeaderBase)(nil)
	var _ RPCStats = (*RPCInTrailerBase)(nil)
	var _ RPCStats = (*RPCOutPayloadBase)(nil)
	var _ RPCStats = (*OutHeaderBase)(nil)
	var _ RPCStats = (*OutTrailerBase)(nil)
	var _ RPCStats = (*RPCEndBase)(nil)
}
