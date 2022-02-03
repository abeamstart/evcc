package cloud

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"errors"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var Host = "cloud.evcc.io:8080"

var conn *grpc.ClientConn

//go:embed ca-cert.pem
var caCert []byte

func caPEM() []byte {
	copy := bytes.NewBuffer(caCert)
	return copy.Bytes()
}

func loadTLSCredentials() (*tls.Config, error) {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	if !certPool.AppendCertsFromPEM(caPEM()) {
		return nil, errors.New("failed to add CA certificate")
	}

	// create the credentials and return it
	config := &tls.Config{
		RootCAs: certPool,
	}

	return config, nil
}

func Connection(uri string) (*grpc.ClientConn, error) {
	var err error
	if conn == nil {
		creds := insecure.NewCredentials()
		if !strings.HasPrefix(uri, "localhost") {
			var tlsConfig *tls.Config
			if tlsConfig, err = loadTLSCredentials(); err != nil {
				return nil, err
			}

			creds = credentials.NewTLS(tlsConfig)
		}
		conn, err = grpc.Dial(uri, grpc.WithTransportCredentials(creds))
	}

	return conn, err
}
