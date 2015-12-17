// Package push sends notifications over HTTP/2 to
// Apple's Push Notification Service.
package push

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

// Headers sent with a push to control the notification.
//
// TODO: need documentation on format of available headers.
type Headers struct {
	Expiration time.Time // apns-expiration
	// apns-id
	// other headers such as priority
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
func (s *Service) Push(deviceToken string, headers Headers, payload json.Marshaler) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.PushBytes(deviceToken, headers, b)
}

// PushBytes notification to APN service.
func (s *Service) PushBytes(deviceToken string, headers Headers, payload []byte) error {
	urlStr := fmt.Sprintf("%v/3/device/%v", s.Host, deviceToken)

	req, err := http.NewRequest("POST", urlStr, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	// TODO: set the apns-* headers

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
