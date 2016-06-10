package push_test

import (
	"encoding/json"
	"flag"
	"testing"

	"github.com/RobotsAndPencils/buford/certificate"
	"github.com/RobotsAndPencils/buford/payload"
	"github.com/RobotsAndPencils/buford/payload/badge"
	"github.com/RobotsAndPencils/buford/push"
)

var (
	deviceToken        string
	filename, password string
	workers            uint
)

func init() {
	flag.StringVar(&deviceToken, "token", "", "Device token")
	flag.StringVar(&filename, "cert", "", "Path to p12 certificate file")
	flag.StringVar(&password, "pwd", "", "Password for p12 file.")
	flag.UintVar(&workers, "w", 20, "Workers to send notifications")
	flag.Parse()
}

// GODEBUG=http2debug=1 go test ./push -cert ../cert.p12 -token device-token -v -bench . -benchtime 10s
func BenchmarkPush(b *testing.B) {
	if filename == "" || deviceToken == "" {
		b.Skip("Skipping benchmark without cert file and device token.")
	}

	cert, err := certificate.Load(filename, password)
	if err != nil {
		b.Fatal(err)
	}

	client, err := push.NewClient(cert)
	if err != nil {
		b.Fatal(err)
	}

	service := push.NewService(client, push.Development)
	queue := push.NewQueue(service, workers)

	p := payload.APS{
		Alert: payload.Alert{Body: "Hello HTTP/2"},
		Badge: badge.New(42),
	}
	bytes, err := json.Marshal(p)
	if err != nil {
		b.Fatal(err)
	}

	// warm up the connection
	_, err = service.Push(deviceToken, nil, bytes)
	if err != nil {
		b.Fatal(err)
	}

	// handle responses
	go func() {
		for {
			_, _, err := queue.Response()
			if err != nil {
				b.Fatal(err)
			}
		}
	}()

	b.ResetTimer()
	// this benchmark is the time to send the notifications without waiting
	// for the responses
	for i := 0; i < b.N; i++ {
		queue.Push(deviceToken, nil, bytes)
	}
	queue.Wait()
}
