package governor

import (
	"fmt"

	"github.com/imkuqin-zw/yggdrasil/pkg/log"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xnet"
)

type Config struct {
	Host string
	Port uint64
}

// Address
func (config *Config) Address() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}

// Build
func (config *Config) Build() *Server {
	var err error
	config.Host, err = xnet.Extract(config.Host)
	if err != nil {
		log.Fatalf("create governor server, error: %s", err.Error())
	}

	return newServer(config)
}
