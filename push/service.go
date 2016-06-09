// Package push sends notifications over HTTP/2 to
// Apple's Push Notification Service.
package push

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/http2"
)

// Apple host locations for configuring Service.
const (
	Development = "https://api.development.push.apple.com"
	Production  = "https://api.push.apple.com"
)

// Service is the Apple Push Notification Service that you send notifications to.
type Service struct {
	Host   string
	Client *http.Client
}

// NewService sets up an HTTP/2 client for a certificate. If you need to do
// something custom, you can always override the fields in Service,
// e.g. to specify your own http.Client.
func NewService(host string, cert tls.Certificate) (*Service, error) {
	client, err := newClient(cert)
	if err != nil {
		return nil, err
	}

	return &Service{
		Client: client,
		Host:   host,
	}, nil
}

// NewClient sets up an HTTP/2 client for a certificate.
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

// Push notification to APN service after performing serialization.
func (s *Service) Push(deviceToken string, headers *Headers, payload interface{}) (string, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return s.PushBytes(deviceToken, headers, b)
}

// PushBytes notification to APN service.
func (s *Service) PushBytes(deviceToken string, headers *Headers, payload []byte) (string, error) {
	urlStr := fmt.Sprintf("%v/3/device/%v", s.Host, deviceToken)

	req, err := http.NewRequest("POST", urlStr, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	headers.set(req.Header)

	resp, err := s.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return resp.Header.Get("apns-id"), nil
	}

	var response struct {
		// Reason for failure
		Reason string `json:"reason"`
		// Timestamp for 410 StatusGone (ErrUnregistered)
		Timestamp int64 `json:"timestamp"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	es := &Error{
		Reason: mapErrorReason(response.Reason),
		Status: resp.StatusCode,
	}

	if response.Timestamp != 0 {
		// the response.Timestamp is Milliseconds, but time.Unix() requires seconds
		es.Timestamp = time.Unix(response.Timestamp/1000, 0).UTC()
	}

	return "", es
}
