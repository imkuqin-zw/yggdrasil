package otlpgrpc

type Config struct {
	// TlsMode is the TLS configuration mode
	// ServerTLS: Server-side TLS configuration
	// MutualTLS: Mutual TLS configuration
	// Insecure: No TLS configuration
	TlsMode string
	// CaCrt is the CA certificate
	CaCrt string
	// ServerName is the server name
	ServerName string
	// ClientCrt is the client certificate
	ClientCrt string
	// ClientKey is the client key
	ClientKey string
	// Endpoint is the endpoint
	Endpoint string
}
