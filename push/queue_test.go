package push_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RobotsAndPencils/buford/push"
)

func TestQueuePush(t *testing.T) {
	const (
		workers = 10
		number  = 100
	)
	payload := []byte(`{ "aps" : { "alert" : "Hello HTTP/2" } }`)

	handler := http.NewServeMux()
	server := httptest.NewServer(handler)

	handler.HandleFunc("/3/device/", func(w http.ResponseWriter, r *http.Request) {
		deviceToken := strings.TrimPrefix(r.URL.String(), "/3/device/")
		// echo back the deviceToken as the id (not the real behavior)
		w.Header().Set("apns-id", deviceToken)
	})

	service := push.NewService(http.DefaultClient, server.URL)
	queue := push.NewQueue(service, workers)

	go func() {
		for i := 0; i < number; i++ {
			id, deviceToken, err := queue.Response()
			if err != nil {
				t.Error(err)
			}
			if id != deviceToken {
				t.Errorf("Expected %q == %q.", id, deviceToken)
			}
		}
	}()

	for i := 0; i < number; i++ {
		queue.Push(fmt.Sprintf("%04d", i), nil, payload)
	}
	queue.Wait()
}
