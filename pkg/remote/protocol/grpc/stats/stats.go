package stats

import "github.com/imkuqin-zw/yggdrasil/pkg/stats"

type ClientInHeader struct {
	stats.RPCClientInHeaderBase
	Compression string
}

func (s *ClientInHeader) GetCompression() string {
	return s.Compression

}

type ServerInHeader struct {
	stats.RPCServerInHeaderBase
	Compression string
}

func (s *ServerInHeader) GetCompression() string {
	return s.Compression

}

type OutHeader struct {
	stats.OutHeaderBase
	Compression string
}

func (s *OutHeader) GetCompression() string {
	return s.Compression
}

type InPayload struct {
	stats.RPCInPayloadBase
	Compression      string
	CompressedLength int
}

func (s *InPayload) GetCompression() string {
	return s.Compression
}

func (s *InPayload) GetCompressedLength() int {
	return s.CompressedLength
}

type OutPayload struct {
	stats.RPCOutPayloadBase
	Compression      string
	CompressedLength int
}

func (s *OutPayload) GetCompression() string {
	return s.Compression
}

func (s *OutPayload) GetCompressedLength() int {
	return s.CompressedLength
}
