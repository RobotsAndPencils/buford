package push_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/RobotsAndPencils/buford/push"
)

func TestPush(t *testing.T) {
	deviceToken := "c2732227a1d8021cfaf781d71fb2f908c61f5861079a00954a5453f1d0281433"
	payload := []byte(`{ "aps" : { "alert" : "Hello HTTP/2" } }`)
	apnsID := "922D9F1F-B82E-B337-EDC9-DB4FC8527676"

	handler := http.NewServeMux()
	server := httptest.NewServer(handler)

	handler.HandleFunc("/3/device/", func(w http.ResponseWriter, r *http.Request) {
		expectURL := fmt.Sprintf("/3/device/%s", deviceToken)
		if r.URL.String() != expectURL {
			t.Errorf("Expected url %v, got %v", expectURL, r.URL)
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(body, payload) {
			t.Errorf("Expected body %v, got %v", payload, body)
		}

		w.Header().Set("apns-id", apnsID)
	})

	service := push.Service{
		Client: http.DefaultClient,
		Host:   server.URL,
	}

	id, err := service.PushBytes(deviceToken, &push.Headers{}, payload)
	if err != nil {
		t.Error(err)
	}
	if id != apnsID {
		t.Errorf("Expected apns-id %q, but got %q.", apnsID, id)
	}
}

func TestBadPriorityPush(t *testing.T) {
	deviceToken := "c2732227a1d8021cfaf781d71fb2f908c61f5861079a00954a5453f1d0281433"
	payload := []byte(`{ "aps" : { "alert" : "Hello HTTP/2" } }`)

	handler := http.NewServeMux()
	server := httptest.NewServer(handler)

	handler.HandleFunc("/3/device/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"reason": "BadPriority"}`))
	})

	service := push.Service{
		Client: http.DefaultClient,
		Host:   server.URL,
	}

	_, err := service.PushBytes(deviceToken, nil, payload)
	if err != push.ErrBadPriority {
		t.Errorf("Expected error %v, got %v.", push.ErrBadPriority, err)
	}
}

func TestTimestampError(t *testing.T) {
	deviceToken := "c2732227a1d8021cfaf781d71fb2f908c61f5861079a00954a5453f1d0281433"
	payload := []byte(`{ "aps" : { "alert" : "Hello HTTP/2" } }`)

	handler := http.NewServeMux()
	server := httptest.NewServer(handler)

	handler.HandleFunc("/3/device/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGone)
		w.Write([]byte(`{"reason":"Unregistered","timestamp":1262278080000}`))
	})

	service := push.Service{
		Client: http.DefaultClient,
		Host:   server.URL,
	}

	_, err := service.PushBytes(deviceToken, nil, payload)

	if err == push.ErrUnregistered {
		t.Error("Expected error structure, got constant.")
	}

	e, ok := err.(*push.Error)
	if !ok {
		t.Fatalf("Expected push error, got %v.", err)
	}

	if e.Err != push.ErrUnregistered {
		t.Errorf("Expected error %v, got %v.", push.ErrUnregistered, err)
	}
	// fix the unixtime
	expected := time.Unix(1262278080, 0)
	if e.Timestamp != expected {
		t.Errorf("Expected timestamp %v, got %v.", expected, e.Timestamp)
	}

	if e.DeviceToken != deviceToken {
		t.Errorf("Expected device token %v, got %v.", deviceToken, e.DeviceToken)
	}
}
