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

// Package peer defines various peer information associated with RPCs and
// corresponding utils.
package peer

import (
	"context"
	"net"

	"github.com/imkuqin-zw/yggdrasil/pkg/remote/credentials"
)

// Peer contains the information of the peer for an RPC, such as the address
// and authentication information.
type Peer struct {
	// Addr is the peer address.
	Addr net.Addr
	// LocalAddr is the local address.
	LocalAddr net.Addr
	// AuthInfo is the authentication information of the transport.
	// It is nil if there is no transport security being used.
	AuthInfo credentials.AuthInfo
	// Protocol is the protocol used for the RPC.
	Protocol string
}

type peerKey struct{}

// PeerWithContext creates a new context with peer information attached.
func PeerWithContext(ctx context.Context, p *Peer) context.Context {
	return context.WithValue(ctx, peerKey{}, p)
}

// PeerFromContext returns the peer information in ctx if it exists.
func PeerFromContext(ctx context.Context) (p *Peer, ok bool) {
	p, ok = ctx.Value(peerKey{}).(*Peer)
	return
}
