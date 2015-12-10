// Package buford push notifications to Apple.
package buford

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Apple host locations
const (
	Sandbox = "https://api.sandbox.push.apple.com"
	Live    = "https://api.push.apple.com"
)

// Service is the Apple Push Notification Service.
type Service struct {
	Client *http.Client
	Host   string
}

// Service error responses
var (
	ErrBadDeviceToken = errors.New("bad device token")
	ErrForbidden      = errors.New("forbidden, check your certificate")
)

// NewClient sets up an HTTPS client.
func NewClient(cert tls.Certificate) *http.Client {
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	config.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: config}

	return &http.Client{Transport: transport}
}

type response struct {
	Reason string `json:"reason"`
	// timestamp, other fields?
}

// Push a notification.
func (s *Service) Push(deviceToken string, payload []byte) error {
	urlStr := fmt.Sprintf("%v/3/device/%v", s.Host, deviceToken)

	req, err := http.NewRequest("POST", urlStr, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apns-expiration", "0")
	// TODO: apns-id, other headers such as priority

	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// read entire response body
	// TODO: could decode while reading instead if not logging body too
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// http.StatusBadRequest, http.StatusForbidden
	log.Println(resp.StatusCode)
	// logging full responses while learning the API
	log.Println(string(body))

	var response response
	json.Unmarshal(body, &response)

	switch response.Reason {
	case "BadDeviceToken":
		return ErrBadDeviceToken
	case "Forbidden":
		return ErrForbidden
	}
	return fmt.Errorf("Error response: %v", response.Reason)
}
