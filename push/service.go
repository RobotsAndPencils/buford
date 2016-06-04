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

// Apple host locations for configuring Service.
const (
	Development = "https://api.development.push.apple.com"
	Production  = "https://api.push.apple.com"
)

// Service is the Apple Push Notification Service that you send notifications to.
type Service struct {
	Host   string
	Client *http.Client
	Topic  string // extracted from certificate
}

// Headers sent with a push to control the notification (optional)
type Headers struct {
	// ID for the notification. Apple generates one if omitted.
	// This should be a UUID with 32 lowercase hexadecimal digits.
	// TODO: use a UUID type.
	ID string

	// Apple will retry delivery until this time. The default behavior only tries once.
	Expiration time.Time

	// Allow Apple to group messages together to reduce power consumption.
	// By default messages are sent immediately.
	LowPriority bool
}

// Error responses from Apple
type Error struct {
	Reason    error
	Status    int // http StatusCode
	Timestamp time.Time
}

// Service error responses.
var (
	// These could be checked prior to sending the request to Apple.

	ErrPayloadEmpty    = errors.New("PayloadEmpty")
	ErrPayloadTooLarge = errors.New("PayloadTooLarge")

	// Device token errors.

	ErrMissingDeviceToken = errors.New("MissingDeviceToken")
	ErrBadDeviceToken     = errors.New("BadDeviceToken")
	ErrTooManyRequests    = errors.New("TooManyRequests")

	// Header errors.

	ErrBadMessageID      = errors.New("BadMessageID")
	ErrBadExpirationDate = errors.New("BadExpirationDate")
	ErrBadPriority       = errors.New("BadPriority")
	ErrBadTopic          = errors.New("BadTopic")

	// Certificate and topic errors.

	ErrBadCertificate            = errors.New("BadCertificate")
	ErrBadCertificateEnvironment = errors.New("BadCertificateEnvironment")
	ErrForbidden                 = errors.New("Forbidden")

	ErrMissingTopic           = errors.New("MissingTopic")
	ErrTopicDisallowed        = errors.New("TopicDisallowed")
	ErrUnregistered           = errors.New("Unregistered")
	ErrDeviceTokenNotForTopic = errors.New("DeviceTokenNotForTopic")

	// These errors should never happen when using Push.

	ErrDuplicateHeaders = errors.New("DuplicateHeaders")
	ErrBadPath          = errors.New("BadPath")
	ErrMethodNotAllowed = errors.New("MethodNotAllowed")

	// Fatal server errors.

	ErrIdleTimeout         = errors.New("IdleTimeout")
	ErrShutdown            = errors.New("Shutdown")
	ErrInternalServerError = errors.New("InternalServerError")
	ErrServiceUnavailable  = errors.New("ServiceUnavailable")
)

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
	if s.Topic != "" {
		req.Header.Set("apns-topic", s.Topic)
	}

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
}

// mapErrorReason converts Apple error responses into exported Err variables
// for comparisons.
func mapErrorReason(reason string) error {
	var e error
	switch reason {
	case "PayloadEmpty":
		e = ErrPayloadEmpty
	case "PayloadTooLarge":
		e = ErrPayloadTooLarge
	case "BadTopic":
		e = ErrBadTopic
	case "TopicDisallowed":
		e = ErrTopicDisallowed
	case "BadMessageId":
		e = ErrBadMessageID
	case "BadExpirationDate":
		e = ErrBadExpirationDate
	case "BadPriority":
		e = ErrBadPriority
	case "MissingDeviceToken":
		e = ErrMissingDeviceToken
	case "BadDeviceToken":
		e = ErrBadDeviceToken
	case "DeviceTokenNotForTopic":
		e = ErrDeviceTokenNotForTopic
	case "Unregistered":
		e = ErrUnregistered
	case "DuplicateHeaders":
		e = ErrDuplicateHeaders
	case "BadCertificateEnvironment":
		e = ErrBadCertificateEnvironment
	case "BadCertificate":
		e = ErrBadCertificate
	case "Forbidden":
		e = ErrForbidden
	case "BadPath":
		e = ErrBadPath
	case "MethodNotAllowed":
		e = ErrMethodNotAllowed
	case "TooManyRequests":
		e = ErrTooManyRequests
	case "IdleTimeout":
		e = ErrIdleTimeout
	case "Shutdown":
		e = ErrShutdown
	case "InternalServerError":
		e = ErrInternalServerError
	case "ServiceUnavailable":
		e = ErrServiceUnavailable
	case "MissingTopic":
		e = ErrMissingTopic
	default:
		e = errors.New(reason)
	}
	return e
}

func (e *Error) Error() string {
	switch e.Reason {
	case ErrPayloadEmpty:
		return "the message payload was empty"
	case ErrPayloadTooLarge:
		return "the message payload was too large"
	case ErrMissingDeviceToken:
		return "device token was not specified"
	case ErrBadDeviceToken:
		return "bad device token"
	case ErrTooManyRequests:
		return "too many requests were made consecutively to the same device token"
	case ErrBadMessageID:
		return "the ID header value is bad"
	case ErrBadExpirationDate:
		return "the Expiration header value is bad"
	case ErrBadPriority:
		return "the apns-priority value is bad"
	case ErrBadTopic:
		return "the Topic header was invalid"
	case ErrBadCertificate:
		return "the certificate was bad"
	case ErrBadCertificateEnvironment:
		return "certificate was for the wrong environment"
	case ErrForbidden:
		return "there was an error with the certificate"
	case ErrMissingTopic:
		return "the Topic header of the request was not specified and was required"
	case ErrTopicDisallowed:
		return "pushing to this topic is not allowed"
	case ErrUnregistered:
		return fmt.Sprintf("device token is inactive for the specified topic (last invalid at %v)", e.Timestamp)
	case ErrDeviceTokenNotForTopic:
		return "device token does not match the specified topic"
	case ErrDuplicateHeaders:
		return "one or more headers were repeated"
	case ErrBadPath:
		return "the request contained a bad :path"
	case ErrMethodNotAllowed:
		return "the specified :method was not POST"
	case ErrIdleTimeout:
		return "idle time out"
	case ErrShutdown:
		return "the server is shutting down"
	case ErrInternalServerError:
		return "an internal server error occurred"
	case ErrServiceUnavailable:
		return "the service is unavailable"
	default:
		return fmt.Sprintf("unknown error: %v", e.Reason.Error())
	}
}
