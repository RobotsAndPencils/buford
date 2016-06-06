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

var deviceToken, filename, password, environment string

func init() {
	flag.StringVar(&deviceToken, "token", "", "Device token")
	flag.StringVar(&filename, "cert", "", "Path to p12 certificate file")
	flag.StringVar(&password, "pwd", "", "Password for p12 file.")
	flag.StringVar(&environment, "env", "development", "Environment")
	flag.Parse()
}

// GODEBUG=http2debug=1 go test ./push -cert ../cert.p12 -token device-token -v -bench . -benchtime 30s
func BenchmarkPush(b *testing.B) {
	if filename == "" || deviceToken == "" {
		b.Skipf("Skipping benchmark without cert file and device token.")
	}

	cert, err := certificate.Load(filename, password)
	if err != nil {
		b.Fatal(err)
	}

	service, err := push.NewService(push.Development, cert)
	if err != nil {
		b.Fatal(err)
	}
	if environment == "production" {
		service.Host = push.Production
	}

	p := payload.APS{
		Alert: payload.Alert{Body: "Hello HTTP/2"},
		Badge: badge.New(42),
	}

	payload, err := json.Marshal(p)
	if err != nil {
		b.Fatal(err)
	}

	// warm up the connection
	_, err = service.PushBytes(deviceToken, nil, payload)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = service.PushBytes(deviceToken, nil, payload)
		if err != nil {
			b.Fatal(err)
		}
	}
}
