package xtls

import (
	"crypto/tls"
)

// SSLConfig struct stores the necessary info for SSL configuration
type SSLConfig struct {
	CipherPlugin string
	VerifyPeer   bool
	CipherSuites string
	MinVersion   string
	MaxVersion   string
	CA           []byte
	Cert         []byte
	Key          []byte
	CertPWD      string
	ServerName   string
}

// ClientTLSConfig function gets client side TLS config
func (c *SSLConfig) ClientTLSConfig() (*tls.Config, error) {
	return getTLSConfig(c, "client")
}

// ServerTLSConfig function gets server side TLD config
func (c *SSLConfig) ServerTLSConfig() (*tls.Config, error) {
	return getTLSConfig(c, "server")
}
