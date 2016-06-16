package push

import (
	"net/http"
	"testing"
	"time"
)

func TestHeaders(t *testing.T) {
	headers := Headers{
		ID:          "uuid",
		CollapseID:  "game1.score.identifier",
		Expiration:  time.Unix(12622780800, 0),
		LowPriority: true,
		Topic:       "bundle-id",
	}

	reqHeader := http.Header{}
	headers.set(reqHeader)

	testHeader(t, reqHeader, "apns-id", "uuid")
	testHeader(t, reqHeader, "apns-collapse-id", "game1.score.identifier")
	testHeader(t, reqHeader, "apns-expiration", "12622780800")
	testHeader(t, reqHeader, "apns-priority", "5")
	testHeader(t, reqHeader, "apns-topic", "bundle-id")
}

func TestNilHeader(t *testing.T) {
	var headers *Headers
	reqHeader := http.Header{}
	headers.set(reqHeader)

	testHeader(t, reqHeader, "apns-id", "")
	testHeader(t, reqHeader, "apns-collapse-id", "")
	testHeader(t, reqHeader, "apns-expiration", "")
	testHeader(t, reqHeader, "apns-priority", "")
	testHeader(t, reqHeader, "apns-topic", "")
}

func TestEmptyHeaders(t *testing.T) {
	headers := Headers{}
	reqHeader := http.Header{}
	headers.set(reqHeader)

	testHeader(t, reqHeader, "apns-id", "")
	testHeader(t, reqHeader, "apns-collapse-id", "")
	testHeader(t, reqHeader, "apns-expiration", "")
	testHeader(t, reqHeader, "apns-priority", "")
	testHeader(t, reqHeader, "apns-topic", "")
}

func testHeader(t *testing.T, reqHeader http.Header, key, expected string) {
	actual := reqHeader.Get(key)
	if actual != expected {
		t.Errorf("Expected %s %q, got %q.", key, expected, actual)
	}
}
