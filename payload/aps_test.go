package payload_test

import (
	"testing"

	"github.com/RobotsAndPencils/buford/payload"
	"github.com/RobotsAndPencils/buford/payload/badge"
)

func TestSimpleAlert(t *testing.T) {
	p := payload.APS{Alert: payload.Alert{Body: "Message received from Bob"}}
	expected := []byte(`{"aps":{"alert":"Message received from Bob"}}`)
	testPayload(t, p, expected)
}

func TestBadgeAndSound(t *testing.T) {
	p := payload.APS{
		Alert: payload.Alert{Body: "You got your emails."},
		Badge: badge.New(9),
		Sound: "bingbong.aiff",
	}
	expected := []byte(`{"aps":{"alert":"You got your emails.","badge":9,"sound":"bingbong.aiff"}}`)
	testPayload(t, p, expected)
}

func TestContentAvailable(t *testing.T) {
	p := payload.APS{ContentAvailable: true}
	expected := []byte(`{"aps":{"content-available":1}}`)
	testPayload(t, p, expected)
}

func TestCustomArray(t *testing.T) {
	p := payload.APS{Alert: payload.Alert{Body: "Message received from Bob"}}
	pm := p.Map()
	pm["acme2"] = []string{"bang", "whiz"}
	expected := []byte(`{"acme2":["bang","whiz"],"aps":{"alert":"Message received from Bob"}}`)
	testPayload(t, pm, expected)
}

func TestAlertDictionary(t *testing.T) {
	p := payload.APS{Alert: payload.Alert{Title: "Message", Body: "Message received from Bob"}}
	expected := []byte(`{"aps":{"alert":{"title":"Message","body":"Message received from Bob"}}}`)
	testPayload(t, p, expected)
}
