package push

import (
	"crypto/tls"
	"net/http"

	"golang.org/x/net/http2"
)

// NewClient sets up an HTTP/2 client for a certificate.
// If you need to do something custom, you can always specify your own
// http.Client for Service.
func NewClient(cert tls.Certificate) (*http.Client, error) {
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	config.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: config}

	if err := http2.ConfigureTransport(transport); err != nil {
		return nil, err
	}

	return &http.Client{Transport: transport}, nil
}
