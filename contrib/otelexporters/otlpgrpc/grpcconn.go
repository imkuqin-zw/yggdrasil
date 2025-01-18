package otlpgrpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"time"

	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var grpcConn *grpc.ClientConn

func initGrpcConn() {
	if grpcConn != nil {
		return
	}
	cfg := &Config{}
	if err := config.Get("otlp.grpc").Scan(cfg); err != nil {
		logger.FatalField("failed to scan otlp.grpc config", logger.Err(err))
	}
	var creds credentials.TransportCredentials
	switch cfg.TlsMode {
	case "ServerTLS":
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM([]byte(cfg.CaCrt)) {
			logger.Fatal("credentials: failed to append certificates")
		}
		creds = credentials.NewClientTLSFromCert(cp, cfg.ServerName)
	case "MutualTLS":
		certificate, err := tls.X509KeyPair([]byte(cfg.ClientCrt), []byte(cfg.ClientKey))
		if err != nil {
			logger.FatalField("fault to load client crt and key", logger.Err(err))
		}
		// 构建CertPool以校验服务端证书有效性
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM([]byte(cfg.CaCrt)) {
			logger.Fatal("credentials: failed to append certificates")
		}
		creds = credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{certificate},
			ServerName:   cfg.ServerName,
			RootCAs:      cp,
		})
	default:
		creds = insecure.NewCredentials()
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, cfg.Endpoint,
		grpc.WithTransportCredentials(creds),
		//grpc.WithBlock(),
	)
	if err != nil {
		logger.FatalField("failed to create gRPC connection to collector", logger.Err(err))
	}
	grpcConn = conn
}
