package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

func LoadTLSConfig() (*tls.Config, error) {
	certFile := os.Getenv("TRAIT_SERVICE_CERT_FILE")
	keyFile := os.Getenv("TRAIT_SERVICE_KEY_FILE")
	caFile := os.Getenv("TRAIT_SERVICE_CA_FILE")

	// Загружаем сертификат и ключ клиента (Task сервис)
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	// Загружаем CA сертификат
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA cert")
	}

	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   "auth-service", // Имя из CN сертификата Auth
	}

	return &tlsConfig, nil
}
