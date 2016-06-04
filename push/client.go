package push

import (
	"crypto/tls"
	"net/http"
	"strings"

	"golang.org/x/net/http2"
)

// NewService sets up an HTTP/2 client for a certificate.
func NewService(host string, cert tls.Certificate) (*Service, error) {
	client, err := newClient(cert)
	if err != nil {
		return nil, err
	}

	// extract common name
	// Apple Push Services: {bundle}
	// Apple Development IOS Push Services: {bundle}
	var bundle string
	commonName := cert.Leaf.Subject.CommonName
	n := strings.Index(commonName, ":")
	if n != -1 {
		bundle = strings.TrimSpace(commonName[n+1:])
	}

	return &Service{
		Client: client,
		Host:   host,
		Topic:  bundle,
	}, nil
}

// NewClient sets up an HTTP/2 client for a certificate.
// If you need to do something custom, you can always specify your own
// http.Client for Service.
func newClient(cert tls.Certificate) (*http.Client, error) {
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
