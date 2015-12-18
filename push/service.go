// Package push sends notifications over HTTP/2 to
// Apple's Push Notification Service.
package push

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// Apple host locations.
const (
	Sandbox = "https://api.sandbox.push.apple.com"
	Live    = "https://api.push.apple.com"
)

// Service is the Apple Push Notification Service that you send notifications to.
type Service struct {
	Client *http.Client
	Host   string
}

// Headers sent with a push to control the notification (optional)
type Headers struct {
	// ID for the notification. Apple generates one if ommitted.
	// This should be a UUID with 32 lowercase hexadecimal digits.
	// TODO: use a UUID type.
	ID string

	// Apple will retry delivery until this time. The default behavior only tries once.
	Expiration time.Time

	// Allow Apple to group messages to together to reduce power consumption.
	// By default messages are sent immediately.
	LowPriority bool

	// Topic is the bundle ID for your app.
	Topic string
}

// Service error responses.
var (
	ErrBadDeviceToken = errors.New("bad device token")
	ErrForbidden      = errors.New("forbidden, check your certificate")
)

type response struct {
	Reason string `json:"reason"`
	// timestamp, other fields?
}

// Push notification to APN service after performing serialization.
func (s *Service) Push(deviceToken string, headers *Headers, payload json.Marshaler) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.PushBytes(deviceToken, headers, b)
}

// PushBytes notification to APN service.
func (s *Service) PushBytes(deviceToken string, headers *Headers, payload []byte) error {
	urlStr := fmt.Sprintf("%v/3/device/%v", s.Host, deviceToken)

	req, err := http.NewRequest("POST", urlStr, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	headers.set(req)

	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// read entire response body
	// TODO: could decode while reading instead
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

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

// set headers on an HTTP request
func (h *Headers) set(req *http.Request) {
	// headers are optional
	if h == nil {
		return
	}

	if h.ID != "" {
		req.Header.Set("apns-id", h.ID)
	} // when ommitted, Apple will generate a UUID for you

	if !h.Expiration.IsZero() {
		req.Header.Set("apns-expiration", strconv.FormatInt(h.Expiration.Unix(), 10))
	}

	if h.LowPriority {
		req.Header.Set("apns-priority", "5")
	} // when ommitted, the default priority is 10

	if h.Topic != "" {
		req.Header.Set("apns-topic", h.Topic)
	}
}
