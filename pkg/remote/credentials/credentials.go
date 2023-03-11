/*
 *
 * Copyright 2014 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package credentials implements various credentials supported by gRPC library,
// which encapsulate all the state needed by a client to authenticate with a
// server and make various assertions, e.g., about the client's identity, role,
// or whether it is authorized to make a particular call.
package credentials // import "github.com/imkuqin-zw/yggdrasil/pkg/remote/credentials"

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
)

// requestInfoKey is a struct to be used as the key to store RequestInfo in a
// context.
type requestInfoKey struct{}

// NewRequestInfoContext creates a context with ri.
func NewRequestInfoContext(ctx context.Context, ri interface{}) context.Context {
	return context.WithValue(ctx, requestInfoKey{}, ri)
}

// RequestInfoFromContext extracts the RequestInfo from ctx.
func RequestInfoFromContext(ctx context.Context) interface{} {
	return ctx.Value(requestInfoKey{})
}

// clientHandshakeInfoKey is a struct used as the key to store
// ClientHandshakeInfo in a context.
type clientHandshakeInfoKey struct{}

// ClientHandshakeInfoFromContext extracts the ClientHandshakeInfo from ctx.
func ClientHandshakeInfoFromContext(ctx context.Context) ClientHandshakeInfo {
	return ctx.Value(clientHandshakeInfoKey{}).(ClientHandshakeInfo)
}

// NewClientHandshakeInfoContext creates a context with chi.
func NewClientHandshakeInfoContext(ctx context.Context, chi interface{}) context.Context {
	return context.WithValue(ctx, clientHandshakeInfoKey{}, chi)
}

// SecurityLevel defines the protection level on an established connection.
//
// This API is experimental.
type SecurityLevel int

const (
	// InvalidSecurityLevel indicates an invalid security level.
	// The zero SecurityLevel value is invalid for backward compatibility.
	InvalidSecurityLevel SecurityLevel = iota
	// NoSecurity indicates a connection is insecure.
	NoSecurity
	// IntegrityOnly indicates a connection only provides integrity protection.
	IntegrityOnly
	// PrivacyAndIntegrity indicates a connection provides both privacy and integrity protection.
	PrivacyAndIntegrity
)

// String returns SecurityLevel in a string format.
func (s SecurityLevel) String() string {
	switch s {
	case NoSecurity:
		return "NoSecurity"
	case IntegrityOnly:
		return "IntegrityOnly"
	case PrivacyAndIntegrity:
		return "PrivacyAndIntegrity"
	}
	return fmt.Sprintf("invalid SecurityLevel: %v", int(s))
}

// CommonAuthInfo contains authenticated information common to AuthInfo implementations.
// It should be embedded in a struct implementing AuthInfo to provide additional information
// about the credentials.
//
// This API is experimental.
type CommonAuthInfo struct {
	SecurityLevel SecurityLevel
}

// GetCommonAuthInfo returns the pointer to CommonAuthInfo struct.
func (c CommonAuthInfo) GetCommonAuthInfo() CommonAuthInfo {
	return c
}

// ProtocolInfo provides information regarding the gRPC wire protocol version,
// security protocol, security protocol version in use, server name, etc.
type ProtocolInfo struct {
	// ProtocolVersion is the gRPC wire protocol version.
	ProtocolVersion string
	// SecurityProtocol is the security protocol in use.
	SecurityProtocol string
	// SecurityVersion is the security protocol version.  It is a static version string from the
	// credentials, not a value that reflects per-connection protocol negotiation.  To retrieve
	// details about the credentials used for a connection, use the Peer's AuthInfo field instead.
	//
	// Deprecated: please use Peer.AuthInfo.
	SecurityVersion string
	// ServerName is the user-configured server name.
	ServerName string
}

// AuthInfo defines the common interface for the auth information the users are interested in.
// A struct that implements AuthInfo should embed CommonAuthInfo by including additional
// information about the credentials in it.
type AuthInfo interface {
	AuthType() string
}

// ErrConnDispatched indicates that rawConn has been dispatched out of gRPC
// and the caller should not close rawConn.
var ErrConnDispatched = errors.New("credentials: rawConn is dispatched out of gRPC")

// TransportCredentials defines the common interface for all the live gRPC wire
// protocols and supported transport security protocols (e.g., TLS, SSL).
type TransportCredentials interface {
	// ClientHandshake does the authentication handshake specified by the
	// corresponding authentication protocol on rawConn for clients. It returns
	// the authenticated connection and the corresponding auth information
	// about the connection.  The auth information should embed CommonAuthInfo
	// to return additional information about the credentials. Implementations
	// must use the provided context to implement timely cancellation.  gRPC
	// will try to reconnect if the reason returned is a temporary reason
	// (io.EOF, context.DeadlineExceeded or err.Temporary() == true).  If the
	// returned reason is a wrapper reason, implementations should make sure that
	// the reason implements Temporary() to have the correct retry behaviors.
	// Additionally, ClientHandshakeInfo data will be available via the context
	// passed to this call.
	//
	// The second argument to this method is the `:authority` header value used
	// while creating new streams on this connection after authentication
	// succeeds. Implementations must use this as the server name during the
	// authentication handshake.
	//
	// If the returned net.Conn is closed, it MUST close the net.Conn provided.
	ClientHandshake(context.Context, string, net.Conn) (net.Conn, AuthInfo, error)
	// ServerHandshake does the authentication handshake for servers. It returns
	// the authenticated connection and the corresponding auth information about
	// the connection. The auth information should embed CommonAuthInfo to return additional information
	// about the credentials.
	//
	// If the returned net.Conn is closed, it MUST close the net.Conn provided.
	ServerHandshake(net.Conn) (net.Conn, AuthInfo, error)
	// Info provides the ProtocolInfo of this TransportCredentials.
	Info() ProtocolInfo
	// Clone makes a copy of this TransportCredentials.
	Clone() TransportCredentials
	// OverrideServerName specifies the value used for the following:
	// - verifying the hostname on the returned certificates
	// - as SNI in the client's handshake to support virtual hosting
	// - as the value for `:authority` header at stream creation time
	//
	// Deprecated: use grpc.WithAuthority instead. Will be supported
	// throughout 1.x.
	OverrideServerName(string) error
	Name() string
}

// RequestInfo contains request data attached to the context passed to GetRequestMetadata calls.
//
// This API is experimental.
type RequestInfo struct {
	// The method passed to Invoke or NewStream for this RPC. (For proto methods, this has the format "/some.Service/Method")
	Method string
	// AuthInfo contains the information from a security handshake (TransportCredentials.ClientHandshake, TransportCredentials.ServerHandshake)
	AuthInfo AuthInfo
}

// ClientHandshakeInfo holds data to be passed to ClientHandshake. This makes
// it possible to pass arbitrary data to the handshaker from gRPC, resolver,
// balancer etc. Individual credential implementations control the actual
// format of the data that they are willing to receive.
//
// This API is experimental.
type ClientHandshakeInfo struct {
	// Attributes contains the attributes for the address. It could be provided
	// by the gRPC, resolver, balancer etc.
	Attributes map[string]interface{}
}

// CheckSecurityLevel checks if a connection's security level is greater than or equal to the specified one.
// It returns success if 1) the condition is satisified or 2) AuthInfo struct does not implement GetCommonAuthInfo() method
// or 3) CommonAuthInfo.SecurityLevel has an invalid zero value. For 2) and 3), it is for the purpose of backward-compatibility.
//
// This API is experimental.
func CheckSecurityLevel(ai AuthInfo, level SecurityLevel) error {
	type internalInfo interface {
		GetCommonAuthInfo() CommonAuthInfo
	}
	if ai == nil {
		return errors.New("AuthInfo is nil")
	}
	if ci, ok := ai.(internalInfo); ok {
		// CommonAuthInfo.SecurityLevel has an invalid value.
		if ci.GetCommonAuthInfo().SecurityLevel == InvalidSecurityLevel {
			return nil
		}
		if ci.GetCommonAuthInfo().SecurityLevel < level {
			return fmt.Errorf("requires SecurityLevel %v; connection has %v", level, ci.GetCommonAuthInfo().SecurityLevel)
		}
	}
	// The condition is satisfied or AuthInfo struct does not implement GetCommonAuthInfo() method.
	return nil
}

type Builder func(serviceName string, client bool) TransportCredentials

var (
	mu       sync.RWMutex
	builders = map[string]Builder{}
)

func GetBuilder(name string) Builder {
	mu.RLock()
	defer mu.RUnlock()
	f, _ := builders[name]
	return f
}

func RegisterBuilder(name string, f Builder) {
	mu.Lock()
	defer mu.Unlock()
	builders[name] = f
}
