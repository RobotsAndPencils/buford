package push

import (
	"crypto/tls"
	"net/http"
)

// NewClient sets up an HTTPS client for a certificate.
// If you need to do something custom, you can always specify your own
// http.Client for Service.
func NewClient(cert tls.Certificate) *http.Client {
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	config.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: config}

	return &http.Client{Transport: transport}
}
