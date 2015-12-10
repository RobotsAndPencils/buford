package buford

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestPush(t *testing.T) {
	deviceToken := "c2732227a1d8021cfaf781d71fb2f908c61f5861079a00954a5453f1d0281433"
	payload := []byte(`{ "aps" : { "alert" : "Hello HTTP/2" } }`)

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
	})

	service := Service{
		Client: http.DefaultClient,
		Host:   server.URL,
	}

	err := service.Push(deviceToken, payload)
	if err != nil {
		t.Error(err)
	}
}
