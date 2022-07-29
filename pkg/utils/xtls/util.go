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

package xtls

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

//TLSCipherSuiteMap is a map with key of type string and value of type unsigned integer
var tlsCipherSuiteMap = map[string]uint16{
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256": tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384": tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
}

//TLSVersionMap is a map with key of type string and value of type unsigned integer
var tlsVersionMap = map[string]uint16{
	"TLSv1.0": tls.VersionTLS10,
	"TLSv1.1": tls.VersionTLS11,
	"TLSv1.2": tls.VersionTLS12,
}

func getX509CACertPool(caCert []byte) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCert)
	return pool, nil
}

func parseSSLCipherSuites(ciphers string) ([]uint16, error) {
	cipherSuiteList := make([]uint16, 0)
	cipherSuiteNameList := strings.Split(ciphers, ",")
	for _, cipherSuiteName := range cipherSuiteNameList {
		cipherSuiteName = strings.TrimSpace(cipherSuiteName)
		if len(cipherSuiteName) == 0 {
			continue
		}

		if cipherSuite, ok := tlsCipherSuiteMap[cipherSuiteName]; ok {
			cipherSuiteList = append(cipherSuiteList, cipherSuite)
		} else {
			// 配置算法不存在
			return nil, fmt.Errorf("cipher %s not exist", cipherSuiteName)
		}
	}

	return cipherSuiteList, nil
}

func parseSSLProtocol(sprotocol string) (uint16, error) {
	var result uint16 = tls.VersionTLS12
	if protocol, ok := tlsVersionMap[sprotocol]; ok {
		result = protocol
	} else {
		return result, fmt.Errorf("invalid ssl minimal version invalid(%s)", sprotocol)
	}

	return result, nil
}

func getTLSConfig(sslConfig *SSLConfig, role string) (tlsConfig *tls.Config, err error) {
	clientAuthMode := tls.NoClientCert
	var pool *x509.CertPool
	// ca file is needed when veryPeer is true
	if sslConfig.VerifyPeer {
		pool, err = getX509CACertPool(sslConfig.CA)
		if err != nil {
			return nil, err
		}

		clientAuthMode = tls.RequireAndVerifyClientCert
	}

	// certificate is necessary for server, optional for client
	var certs []tls.Certificate
	if !(role == "client" && sslConfig.Key == nil && sslConfig.Cert == nil) {
		cipher, ok := cipheres[sslConfig.CipherPlugin]
		if !ok {
			return nil, fmt.Errorf("cipher [%s] nof found", sslConfig.CipherPlugin)
		}
		certs, err = loadTLSCertificate(sslConfig.Cert, sslConfig.Key, strings.TrimSpace(sslConfig.CertPWD), cipher)
		if err != nil {
			return nil, err
		}
	}
	cipherSuites, err := parseSSLCipherSuites(sslConfig.CipherSuites)
	if err != nil {
		return nil, err
	}
	maxVersion, err := parseSSLProtocol(sslConfig.MaxVersion)
	if err != nil {
		return nil, err
	}
	minVersion, err := parseSSLProtocol(sslConfig.MinVersion)
	if err != nil {
		return nil, err
	}
	switch role {
	case "server":
		tlsConfig = &tls.Config{
			ClientCAs:                pool,
			Certificates:             certs,
			CipherSuites:             cipherSuites,
			PreferServerCipherSuites: true,
			ClientAuth:               clientAuthMode,
			MinVersion:               minVersion,
			MaxVersion:               maxVersion,
		}
	case "client":
		tlsConfig = &tls.Config{
			RootCAs:            pool,
			Certificates:       certs,
			CipherSuites:       cipherSuites,
			InsecureSkipVerify: !sslConfig.VerifyPeer,
			MinVersion:         minVersion,
			MaxVersion:         maxVersion,
			ServerName:         sslConfig.ServerName,
		}
	}

	return tlsConfig, nil
}

//LoadTLSCertificate is a function used to load a certificate
func loadTLSCertificate(certContent, keyContent []byte, passphase string, cipher Cipher) ([]tls.Certificate, error) {
	keyBlock, _ := pem.Decode(keyContent)
	if keyBlock == nil {
		errorMsg := "decode key failed"
		return nil, errors.New(errorMsg)
	}

	plainpass, err := cipher.Decrypt(passphase)
	if err != nil {
		return nil, err
	}

	if x509.IsEncryptedPEMBlock(keyBlock) {
		keyData, err := x509.DecryptPEMBlock(keyBlock, []byte(plainpass))
		if err != nil {
			return nil, errors.New("decrypt key failed")
		}

		// 解密成功，重新编码为无加密的PEM格式文件
		plainKeyBlock := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: keyData,
		}

		keyContent = pem.EncodeToMemory(plainKeyBlock)
	}

	cert, err := tls.X509KeyPair(certContent, keyContent)
	if err != nil {
		errorMsg := "load X509 key pair from cert with key failed"
		return nil, errors.New(errorMsg)
	}

	var certs []tls.Certificate
	certs = append(certs, cert)

	return certs, nil
}
