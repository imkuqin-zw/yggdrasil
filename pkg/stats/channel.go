package stats

// ChanTagInfo defines the relevant information needed by connection context tagger.
type ChanTagInfo interface {
	GetProtocol() string
	GetRemoteEndpoint() string
	GetLocalEndpoint() string
	isChanTagInfo()
}

type ChanTagInfoBase struct {
	// RemoteEndpoint is the remote address of the corresponding transport channel.
	RemoteEndpoint string
	// LocalAddr is the local address of the corresponding transport channel.
	LocalEndpoint string
	// Protocol is the protocol used for the RPC.
	Protocol string
}

// GetRemoteEndpoint is the remote endpoint of the corresponding transport channel.
func (s *ChanTagInfoBase) GetRemoteEndpoint() string { return s.RemoteEndpoint }

// GetLocalEndpoint is the local endpoint of the corresponding transport channel.
func (s *ChanTagInfoBase) GetLocalEndpoint() string { return s.LocalEndpoint }

// GetProtocol is the protocol used for the RPC.
func (s *ChanTagInfoBase) GetProtocol() string { return s.Protocol }

func (s *ChanTagInfoBase) isChanTagInfo() {}

// ChanStats contains the stats of a transport connection.
type ChanStats interface {
	isChanStats()
	IsClient() bool
}

type ChanBegin interface {
	ChanStats
	isBegin()
}

type ChanBeginBase struct {
	// Client is true if this ConnBegin is from client side.
	Client bool
}

// IsClient indicates if this is from client side.
func (s *ChanBeginBase) IsClient() bool { return s.Client }
func (s *ChanBeginBase) isChanStats()   {}
func (s *ChanBeginBase) isBegin()       {}

type ChanEnd interface {
	ChanStats
	isEnd()
}

type ChanEndBase struct {
	// Client is true if this ConnEnd is from client side.
	Client bool
}

// IsClient indicates if this is from client side.
func (s *ChanEndBase) IsClient() bool { return s.Client }
func (s *ChanEndBase) isChanStats()   {}
func (s *ChanEndBase) isEnd()         {}
