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

	// Topic for certificates with multiple topics.
	Topic string
}

// Service error responses.
var (
	// These could be checked prior to sending the request to Apple.

	ErrPayloadEmpty    = errors.New("the message payload was empty")
	ErrPayloadTooLarge = errors.New("the message payload was too large")

	// Device token errors.

	ErrMissingDeviceToken = errors.New("device token was not specified")
	ErrBadDeviceToken     = errors.New("bad device token")
	ErrTooManyRequests    = errors.New("too many requests were made consecutively to the same device token")

	// Header errors.

	ErrBadMessageID      = errors.New("the ID header value is bad")
	ErrBadExpirationDate = errors.New("the Expiration header value is bad")
	ErrBadPriority       = errors.New("the apns-priority value is bad")
	ErrBadTopic          = errors.New("the Topic header was invalid")

	// Certificate and topic errors.

	ErrBadCertificate            = errors.New("the certificate was bad")
	ErrBadCertificateEnvironment = errors.New("certificate was for the wrong environment")
	ErrForbidden                 = errors.New("there was an error with the certificate")

	ErrMissingTopic           = errors.New("the Topic header of the request was not specified and was required")
	ErrTopicDisallowed        = errors.New("pushing to this topic is not allowed")
	ErrUnregistered           = errors.New("device token is inactive for the specified topic")
	ErrDeviceTokenNotForTopic = errors.New("device token does not match the specified topic")

	// These errros should never happen when using Push.

	ErrDuplicateHeaders = errors.New("one or more headers were repeated")
	ErrBadPath          = errors.New("the request contained a bad :path")
	ErrMethodNotAllowed = errors.New("the specified :method was not POST")

	// Fatal server errors.

	ErrIdleTimeout         = errors.New("idle time out")
	ErrShutdown            = errors.New("the server is shutting down")
	ErrInternalServerError = errors.New("an internal server error occurred")
	ErrServiceUnavailable  = errors.New("the service is unavailable")

	// HTTP Status errors.

	ErrBadRequest = errors.New("bad request")
	ErrGone       = errors.New("the device token is no longer active for the topic")
	ErrUnknown    = errors.New("unknown error")
)

var errorReason = map[string]error{
	"PayloadEmpty":              ErrPayloadEmpty,
	"PayloadTooLarge":           ErrPayloadTooLarge,
	"BadTopic":                  ErrBadTopic,
	"TopicDisallowed":           ErrTopicDisallowed,
	"BadMessageId":              ErrBadMessageID,
	"BadExpirationDate":         ErrBadExpirationDate,
	"BadPriority":               ErrBadPriority,
	"MissingDeviceToken":        ErrMissingDeviceToken,
	"BadDeviceToken":            ErrBadDeviceToken,
	"DeviceTokenNotForTopic":    ErrDeviceTokenNotForTopic,
	"Unregistered":              ErrUnregistered,
	"DuplicateHeaders":          ErrDuplicateHeaders,
	"BadCertificateEnvironment": ErrBadCertificateEnvironment,
	"BadCertificate":            ErrBadCertificate,
	"Forbidden":                 ErrForbidden,
	"BadPath":                   ErrBadPath,
	"MethodNotAllowed":          ErrMethodNotAllowed,
	"TooManyRequests":           ErrTooManyRequests,
	"IdleTimeout":               ErrIdleTimeout,
	"Shutdown":                  ErrShutdown,
	"InternalServerError":       ErrInternalServerError,
	"ServiceUnavailable":        ErrServiceUnavailable,
	"MissingTopic":              ErrMissingTopic,
}

type response struct {
	// Reason for failure
	Reason string `json:"reason"`
	// Timestamp for 410 errors (maybe this is an int)
	Timestamp string `json:"timestamp"`
}

const statusTooManyRequests = 429

// Push notification to APN service after performing serialization.
func (s *Service) Push(deviceToken string, headers *Headers, payload json.Marshaler) (string, error) {
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
	headers.set(req)

	resp, err := s.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return resp.Header.Get("apns-id"), nil
	}

	// read entire response body
	// TODO: could decode while reading instead
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response response
	json.Unmarshal(body, &response)

	if e, ok := errorReason[response.Reason]; ok {
		return "", e
	}

	// fallback to HTTP status codes if reason not found in JSON

	switch resp.StatusCode {
	case http.StatusBadRequest:
		return "", ErrBadRequest
	case http.StatusForbidden:
		return "", ErrForbidden
	case http.StatusMethodNotAllowed:
		return "", ErrMethodNotAllowed
	case http.StatusGone:
		// TODO: this should return an error structure with timestamp
		// but I don't know the format of timestamp (Unix time?)
		// and there may be a JSON response handled above (ErrUnregistered?)
		return "", ErrGone
	case http.StatusRequestEntityTooLarge:
		return "", ErrPayloadTooLarge
	case statusTooManyRequests:
		return "", ErrTooManyRequests
	case http.StatusInternalServerError:
		return "", ErrInternalServerError
	case http.StatusServiceUnavailable:
		return "", ErrServiceUnavailable
	}

	return "", ErrUnknown
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
