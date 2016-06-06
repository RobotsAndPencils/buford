package push_test

import (
	"testing"

	"github.com/RobotsAndPencils/buford/certificate"
	"github.com/RobotsAndPencils/buford/push"
)

func TestNewService(t *testing.T) {
	const name = "../testdata/cert.p12"

	cert, err := certificate.Load(name, "")
	if err != nil {
		t.Fatal(err)
	}

	service, err := push.NewService(push.Development, cert)
	if err != nil {
		t.Fatal(err)
	}

	const expectedTopic = ""
	if service.Topic != expectedTopic {
		t.Errorf("Expected topic %q, got %q.", expectedTopic, service.Topic)
	}
}
