// Package push sends notifications over HTTP/2 to
// Apple's Push Notification Service.
package push

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Apple host locations.
const (
	Development = "https://api.development.push.apple.com"
	Production  = "https://api.push.apple.com"
)

// Service is the Apple Push Notification Service that you send notifications to.
type Service struct {
	Client *http.Client
	Host   string
}

// Headers sent with a push to control the notification (optional)
type Headers struct {
	// ID for the notification. Apple generates one if omitted.
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

	// These errors should never happen when using Push.

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
	ErrUnknown    = errors.New("unknown error")
)

// Error with a timestamp
type Error struct {
	Err         error
	Timestamp   time.Time
	DeviceToken string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%v (device token %v last invalid at %v)", e.Err.Error(), e.DeviceToken, e.Timestamp)
}

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

var errorStatus = map[int]error{
	http.StatusBadRequest:            ErrBadRequest,
	http.StatusForbidden:             ErrForbidden,
	http.StatusMethodNotAllowed:      ErrMethodNotAllowed,
	http.StatusGone:                  ErrUnregistered,
	http.StatusRequestEntityTooLarge: ErrPayloadTooLarge,
	http.StatusTooManyRequests:       ErrTooManyRequests,
	http.StatusInternalServerError:   ErrInternalServerError,
	http.StatusServiceUnavailable:    ErrServiceUnavailable,
}

type response struct {
	// Reason for failure
	Reason string `json:"reason"`
	// Timestamp for 410 StatusGone (ErrUnregistered)
	Timestamp int64 `json:"timestamp"`
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

	var response response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	e, ok := errorReason[response.Reason]
	if !ok {
		// fallback to HTTP status codes if reason not found in JSON
		e, ok = errorStatus[resp.StatusCode]
		if !ok {
			e = ErrUnknown
		}
	}

	// error constant if there was no timestamp
	if response.Timestamp == 0 {
		return "", e
	}

	// error struct with a timestamp
	return "", &Error{
		Err: e,
		// the response.Timestamp is Milliseconds, but time.Unix() requires seconds
		Timestamp:   time.Unix(response.Timestamp/1000, 0),
		DeviceToken: deviceToken,
	}
}

// set headers for an HTTP request
func (h *Headers) set(reqHeader http.Header) {
	// headers are optional
	if h == nil {
		return
	}

	if h.ID != "" {
		reqHeader.Set("apns-id", h.ID)
	} // when omitted, Apple will generate a UUID for you

	if !h.Expiration.IsZero() {
		reqHeader.Set("apns-expiration", strconv.FormatInt(h.Expiration.Unix(), 10))
	}

	if h.LowPriority {
		reqHeader.Set("apns-priority", "5")
	} // when omitted, the default priority is 10

	if h.Topic != "" {
		reqHeader.Set("apns-topic", h.Topic)
	}
}
