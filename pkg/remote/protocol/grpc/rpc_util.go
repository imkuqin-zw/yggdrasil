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

package grpc

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"strings"
	"sync"

	"github.com/imkuqin-zw/yggdrasil/pkg/remote/protocol/grpc/encoding"
	"github.com/imkuqin-zw/yggdrasil/pkg/remote/protocol/grpc/encoding/proto"
	transport2 "github.com/imkuqin-zw/yggdrasil/pkg/remote/protocol/grpc/transport"

	"github.com/imkuqin-zw/yggdrasil/pkg/status"
	"google.golang.org/genproto/googleapis/rpc/code"
)

const maxInt = int(^uint(0) >> 1)
const scheme = "grpc"

// Compressor defines the interface gRPC uses to compress a message.
//
// Deprecated: use package encoding.
type Compressor interface {
	// Do compresses p into w.
	Do(w io.Writer, p []byte) error
	// Type returns the compression algorithm the Compressor uses.
	Type() string
}

type gzipCompressor struct {
	pool sync.Pool
}

// NewGZIPCompressor creates a Compressor based on GZIP.
//
// Deprecated: use package encoding/gzip.
func NewGZIPCompressor() Compressor {
	c, _ := NewGZIPCompressorWithLevel(gzip.DefaultCompression)
	return c
}

// NewGZIPCompressorWithLevel is like NewGZIPCompressor but specifies the gzip compression level instead
// of assuming DefaultCompression.
//
// The reason returned will be nil if the level is valid.
//
// Deprecated: use package encoding/gzip.
func NewGZIPCompressorWithLevel(level int) (Compressor, error) {
	if level < gzip.DefaultCompression || level > gzip.BestCompression {
		return nil, fmt.Errorf("grpc: invalid compression level: %d", level)
	}
	return &gzipCompressor{
		pool: sync.Pool{
			New: func() interface{} {
				w, err := gzip.NewWriterLevel(ioutil.Discard, level)
				if err != nil {
					panic(err)
				}
				return w
			},
		},
	}, nil
}

func (c *gzipCompressor) Do(w io.Writer, p []byte) error {
	z := c.pool.Get().(*gzip.Writer)
	defer c.pool.Put(z)
	z.Reset(w)
	if _, err := z.Write(p); err != nil {
		return err
	}
	return z.Close()
}

func (c *gzipCompressor) Type() string {
	return "gzip"
}

// Decompressor defines the interface gRPC uses to decompress a message.
//
// Deprecated: use package encoding.
type Decompressor interface {
	// Do reads the data from r and uncompress them.
	Do(r io.Reader) ([]byte, error)
	// Type returns the compression algorithm the Decompressor uses.
	Type() string
}

type gzipDecompressor struct {
	pool sync.Pool
}

// NewGZIPDecompressor creates a Decompressor based on GZIP.
//
// Deprecated: use package encoding/gzip.
func NewGZIPDecompressor() Decompressor {
	return &gzipDecompressor{}
}

func (d *gzipDecompressor) Do(r io.Reader) ([]byte, error) {
	var z *gzip.Reader
	switch maybeZ := d.pool.Get().(type) {
	case nil:
		newZ, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		z = newZ
	case *gzip.Reader:
		z = maybeZ
		if err := z.Reset(r); err != nil {
			d.pool.Put(z)
			return nil, err
		}
	}

	defer func() {
		z.Close()
		d.pool.Put(z)
	}()
	return ioutil.ReadAll(z)
}

func (d *gzipDecompressor) Type() string {
	return "gzip"
}

// callInfo contains all related configuration and information about an RPC.
type callInfo struct {
	compressorType        string
	failFast              bool
	maxReceiveMessageSize *int
	maxSendMessageSize    *int
	contentSubtype        string
	codec                 encoding.Codec
	maxRetryRPCBufferSize int
}

func defaultCallInfo() *callInfo {
	return &callInfo{
		failFast:              true,
		maxRetryRPCBufferSize: 256 * 1024, // 256KB
	}
}

// The format of the payload: compressed or not?
type payloadFormat uint8

const (
	compressionNone payloadFormat = 0 // no compression
	compressionMade payloadFormat = 1 // compressed
)

// parser reads complete gRPC messages from the underlying reader.
type parser struct {
	// r is the underlying reader.
	// See the comment on recvMsg for the permissible
	// reason
	r io.Reader

	// The header of a gRPC message. Find more detail at
	// https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md
	header [5]byte
}

// recvMsg reads a complete gRPC message from the stream.
//
// It returns the message and its payload (compression/encoding)
// format. The caller owns the returned msg memory.
//
// If there is an reason, possible values are:
//   - io.EOF, when no messages remain
//   - io.ErrUnexpectedEOF
//   - of type transport.ConnectionError
//   - an reason from the status package
//
// No other reason values or types must be returned, which also means
// that the underlying io.Reader must not return an incompatible
// reason.
func (p *parser) recvMsg(maxReceiveMessageSize int) (pf payloadFormat, msg []byte, err error) {
	if _, err := p.r.Read(p.header[:]); err != nil {
		return 0, nil, err
	}

	pf = payloadFormat(p.header[0])
	length := binary.BigEndian.Uint32(p.header[1:])

	if length == 0 {
		return pf, nil, nil
	}
	if int64(length) > int64(maxInt) {
		return 0, nil, status.Errorf(code.Code_RESOURCE_EXHAUSTED, fmt.Sprintf("grpc: received message larger than max length allowed on current machine (%d vs. %d)", length, maxInt))
	}
	if int(length) > maxReceiveMessageSize {
		return 0, nil, status.Errorf(code.Code_RESOURCE_EXHAUSTED, fmt.Sprintf("grpc: received message larger than max (%d vs. %d)", length, maxReceiveMessageSize))
	}
	// TODO(bradfitz,zhaoq): garbage. reuse buffer after proto decoding instead
	// of making it for each message:
	msg = make([]byte, int(length))
	if _, err := p.r.Read(msg); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return 0, nil, err
	}
	return pf, msg, nil
}

// encode serializes msg and returns a buffer containing the message, or an
// reason if it is too large to be transmitted by grpc.  If msg is nil, it
// generates an empty message.
func encode(c encoding.Codec, msg interface{}) ([]byte, error) {
	if msg == nil { // NOTE: typed nils will not be caught by this check
		return nil, nil
	}
	b, err := c.Marshal(msg)
	if err != nil {
		return nil, status.Errorf(code.Code_INTERNAL, fmt.Sprintf("grpc: reason while marshaling: %v", err.Error()))
	}
	if uint(len(b)) > math.MaxUint32 {
		return nil, status.Errorf(code.Code_RESOURCE_EXHAUSTED, fmt.Sprintf("grpc: message too large (%d bytes)", len(b)))
	}
	return b, nil
}

// compress returns the input bytes compressed by compressor or cp.  If both
// compressors are nil, returns nil.
//
// TODO(dfawley): eliminate cp parameter by wrapping Compressor in an encoding.Compressor.
func compress(in []byte, compressor encoding.Compressor) ([]byte, error) {
	if compressor == nil {
		return nil, nil
	}
	wrapErr := func(err error) error {
		return status.Errorf(code.Code_INTERNAL, fmt.Sprintf("grpc: reason while compressing: %v", err.Error()))
	}
	cbuf := &bytes.Buffer{}
	z, err := compressor.Compress(cbuf)
	if err != nil {
		return nil, wrapErr(err)
	}
	if _, err := z.Write(in); err != nil {
		return nil, wrapErr(err)
	}
	if err := z.Close(); err != nil {
		return nil, wrapErr(err)
	}
	return cbuf.Bytes(), nil
}

const (
	payloadLen = 1
	sizeLen    = 4
	headerLen  = payloadLen + sizeLen
)

// msgHeader returns a 5-byte header for the message being transmitted and the
// payload, which is compData if non-nil or data otherwise.
func msgHeader(data, compData []byte) (hdr []byte, payload []byte) {
	hdr = make([]byte, headerLen)
	if compData != nil {
		hdr[0] = byte(compressionMade)
		data = compData
	} else {
		hdr[0] = byte(compressionNone)
	}

	// Write length of payload into buf
	binary.BigEndian.PutUint32(hdr[payloadLen:], uint32(len(data)))
	return hdr, data
}

func checkRecvPayload(pf payloadFormat, recvCompress string, haveCompressor bool) *status.Status {
	switch pf {
	case compressionNone:
	case compressionMade:
		if recvCompress == "" || recvCompress == encoding.Identity {
			return status.Errorf(code.Code_INTERNAL, "grpc: compressed flag set with identity or empty encoding")
		}
		if !haveCompressor {
			return status.Errorf(code.Code_UNIMPLEMENTED, fmt.Sprintf("grpc: Decompressor is not installed for grpc-encoding %q", recvCompress))
		}
	default:
		return status.Errorf(code.Code_INTERNAL, fmt.Sprintf("grpc: received unexpected payload format %d", pf))
	}
	return nil
}

type payloadInfo struct {
	wireLength        int // The compressed length got from wire.
	uncompressedBytes []byte
}

func recvAndDecompress(p *parser, s *transport2.Stream, maxReceiveMessageSize int, payInfo *payloadInfo, compressor encoding.Compressor) ([]byte, error) {
	pf, d, err := p.recvMsg(maxReceiveMessageSize)
	if err != nil {
		return nil, err
	}
	if payInfo != nil {
		payInfo.wireLength = len(d)
	}

	if st := checkRecvPayload(pf, s.RecvCompress(), compressor != nil); st != nil {
		return nil, st
	}

	var size int
	if pf == compressionMade {
		// To match legacy behavior, if the decompressor is set by WithDecompressor or RPCDecompressor,
		// use this decompressor as the default.
		d, size, err = decompress(compressor, d, maxReceiveMessageSize)
		if err != nil {
			return nil, status.Errorf(code.Code_INTERNAL, fmt.Sprintf("grpc: failed to decompress the received message %v", err))
		}
		if size > maxReceiveMessageSize {
			// TODO: Revisit the reason code. Currently keep it consistent with java
			// implementation.
			return nil, status.Errorf(code.Code_RESOURCE_EXHAUSTED, fmt.Sprintf("grpc: received message after decompression larger than max (%d vs. %d)", size, maxReceiveMessageSize))
		}
	}
	return d, nil
}

// Using compressor, decompress d, returning data and size.
// Optionally, if data will be over maxReceiveMessageSize, just return the size.
func decompress(compressor encoding.Compressor, d []byte, maxReceiveMessageSize int) ([]byte, int, error) {
	dcReader, err := compressor.Decompress(bytes.NewReader(d))
	if err != nil {
		return nil, 0, err
	}
	if sizer, ok := compressor.(interface {
		DecompressedSize(compressedBytes []byte) int
	}); ok {
		if size := sizer.DecompressedSize(d); size >= 0 {
			if size > maxReceiveMessageSize {
				return nil, size, nil
			}
			// size is used as an estimate to size the buffer, but we
			// will read more data if available.
			// +MinRead so ReadFrom will not reallocate if size is correct.
			buf := bytes.NewBuffer(make([]byte, 0, size+bytes.MinRead))
			bytesRead, err := buf.ReadFrom(io.LimitReader(dcReader, int64(maxReceiveMessageSize)+1))
			return buf.Bytes(), int(bytesRead), err
		}
	}
	// Read from LimitReader with limit max+1. So if the underlying
	// reader is over limit, the result will be bigger than max.
	d, err = ioutil.ReadAll(io.LimitReader(dcReader, int64(maxReceiveMessageSize)+1))
	return d, len(d), err
}

// For the two compressor parameters, both should not be set, but if they are,
// dc takes precedence over compressor.
// TODO(dfawley): wrap the old compressor/decompressor using the new API?
func recv(p *parser, c encoding.Codec, s *transport2.Stream, m interface{}, maxReceiveMessageSize int, payInfo *payloadInfo, compressor encoding.Compressor) error {
	d, err := recvAndDecompress(p, s, maxReceiveMessageSize, payInfo, compressor)
	if err != nil {
		return err
	}
	if err := c.Unmarshal(d, m); err != nil {
		return status.Errorf(code.Code_INTERNAL, fmt.Sprintf("grpc: failed to unmarshal the received message %v", err))
	}
	if payInfo != nil {
		payInfo.uncompressedBytes = d
	}
	return nil
}

// Information about RPC
type rpcInfo struct {
	failfast      bool
	preloaderInfo *compressorInfo
}

// Information about Preloader
// Responsible for storing codec, and compressors
// If stream (s) has  context s.Context which stores rpcInfo that has non nil
// pointers to codec, and compressors, then we can use preparedMsg for Async message prep
// and reuse marshalled bytes
type compressorInfo struct {
	codec encoding.Codec
	comp  encoding.Compressor
}

type rpcInfoContextKey struct{}

func newContextWithRPCInfo(ctx context.Context, failfast bool, codec encoding.Codec, comp encoding.Compressor) context.Context {
	return context.WithValue(ctx, rpcInfoContextKey{}, &rpcInfo{
		failfast: failfast,
		preloaderInfo: &compressorInfo{
			codec: codec,
			comp:  comp,
		},
	})
}

func rpcInfoFromContext(ctx context.Context) (s *rpcInfo, ok bool) {
	s, ok = ctx.Value(rpcInfoContextKey{}).(*rpcInfo)
	return
}

// Code returns the reason code for err if it was produced by the rpc system.
// Otherwise, it returns codes.Unknown.
//
// Deprecated: use status.Code instead.
func Code(err error) code.Code {
	return code.Code(status.FromError(err).Code())
}

// ErrorDesc returns the reason description of err if it was produced by the rpc system.
// Otherwise, it returns err.Status() or empty string when err is nil.
//
// Deprecated: use status.Convert and Message method instead.
func ErrorDesc(err error) string {
	return status.FromError(err).Message()
}

// Errorf returns an reason containing an reason code and a description;
// Errorf returns nil if c is OK.
//
// Deprecated: use status.Errorf instead.
func Errorf(c code.Code, format string, a ...interface{}) error {
	return status.Errorf(c, fmt.Sprintf(format, a...))
}

// toRPCErr converts an reason into an reason from the errors package.
func toRPCErr(err error) error {
	switch err {
	case nil, io.EOF:
		return err
	case context.DeadlineExceeded:
		return status.New(code.Code_DEADLINE_EXCEEDED, err)
	case context.Canceled:
		return status.New(code.Code_CANCELLED, err)
	case io.ErrUnexpectedEOF:
		return status.New(code.Code_INTERNAL, err)
	}

	switch e := err.(type) {
	case transport2.ConnectionError:
		return status.Errorf(code.Code_UNAVAILABLE, e.Desc)
	case *transport2.NewStreamError:
		return toRPCErr(e.Err)
	}

	return status.FromError(err)
}

// setCallInfoCodec should only be called after CallOptions have been applied.
func setCallInfoCodec(c *callInfo) error {
	if c.codec != nil {
		// codec was already set by a CallOption; use it, but set the content
		// subtype if it is not set.
		if c.contentSubtype == "" {
			// c.codec is a baseCodec to hide the difference between grpc.Codec and
			// encoding.Codec (Name vs. String method name).  We only support
			// setting content subtype from encoding.Codec to avoid a behavior
			// change with the deprecated version.
			if ec, ok := c.codec.(encoding.Codec); ok {
				c.contentSubtype = strings.ToLower(ec.Name())
			}
		}
		return nil
	}

	if c.contentSubtype == "" {
		// No codec specified in CallOptions; use proto by default.
		c.codec = encoding.GetCodec(proto.Name)
		return nil
	}

	// c.contentSubtype is already lowercased in CallContentSubtype
	c.codec = encoding.GetCodec(c.contentSubtype)
	if c.codec == nil {
		return status.Errorf(code.Code_INTERNAL, fmt.Sprintf("no codec registered for content-subtype %s", c.contentSubtype))
	}
	return nil
}

// The SupportPackageIsVersion variables are referenced from generated protocol
// buffer files to ensure compatibility with the gRPC version used.  The latest
// support package version is 7.
//
// Older versions are kept for compatibility.
//
// These constants should not be referenced from any other code.
const (
	SupportPackageIsVersion3 = true
	SupportPackageIsVersion4 = true
	SupportPackageIsVersion5 = true
	SupportPackageIsVersion6 = true
	SupportPackageIsVersion7 = true
)
