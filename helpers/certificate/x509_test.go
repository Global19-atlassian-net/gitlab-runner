package certificate

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCertificate(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	gen := X509Generator{}
	cert, pem, err := gen.Generate("127.0.0.1")

	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	tlsListener := tls.NewListener(listener, &tlsConfig)

	srv := http.Server{
		Addr: tlsListener.Addr().String(),
	}
	go func() {
		err := srv.Serve(tlsListener)
		require.EqualError(t, err, "http: Server closed")
	}()
	defer srv.Close()

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(pem)

	tlsClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	req, err := http.NewRequest(http.MethodPost, "https://"+srv.Addr, nil)
	require.NoError(t, err)

	_, err = tlsClient.Do(req)
	assert.NoError(t, err)

	// Client with no Root CA
	client := &http.Client{}
	req, err = http.NewRequest(http.MethodPost, "https://"+srv.Addr, nil)
	require.NoError(t, err)

	_, err = client.Do(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "certificate signed by unknown authority")
}
