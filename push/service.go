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
	Host          string
	Client        *http.Client
	notifications chan notification
	responses     chan response
}

// notification to send.
type notification struct {
	DeviceToken string
	Headers     *Headers
	Payload     []byte
}

// response from sending notification.
type response struct {
	ApnsID       string
	Err          error
	Notification *notification
}

// NewService creates a new service.
func NewService(client *http.Client, host string, workers uint) *Service {
	service := &Service{
		Client: client,
		Host:   host,
		// unbuffered channels
		notifications: make(chan notification),
		responses:     make(chan response),
	}

	// startup workers to send notifications
	for i := uint(0); i < workers; i++ {
		go worker(service)
	}
	return service
}

// Shutdown the workers
func (s *Service) Shutdown() {
	close(s.notifications)
}

func worker(s *Service) {
	for {
		n, more := <-s.notifications
		if !more {
			return
		}
		id, err := s.pushBytes(n.DeviceToken, n.Headers, n.Payload)
		s.responses <- response{ApnsID: id, Err: err, Notification: &n}
	}
}

// NewClient sets up an HTTP/2 client for a certificate.
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

// Push notification queues a notification APN service after performing serialization.
func (s *Service) Push(deviceToken string, headers *Headers, payload interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	s.PushBytes(deviceToken, headers, b)
	return nil
}

// PushBytes queues a notification to APN service.
func (s *Service) PushBytes(deviceToken string, headers *Headers, payload []byte) {
	n := notification{
		DeviceToken: deviceToken,
		Headers:     headers,
		Payload:     payload,
	}
	s.notifications <- n
}

// Response blocks waiting for a response. Order of responses isn't guaranteed.
func (s *Service) Response() (id string, deviceToken string, err error) {
	resp := <-s.responses
	return resp.ApnsID, resp.Notification.DeviceToken, resp.Err
}

// pushBytes sends a notification and waits for a response.
func (s *Service) pushBytes(deviceToken string, headers *Headers, payload []byte) (string, error) {
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
