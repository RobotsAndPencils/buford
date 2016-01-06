package payload_test

import (
	"testing"

	"github.com/RobotsAndPencils/buford/payload"
	"github.com/RobotsAndPencils/buford/payload/badge"
)

func TestPayload(t *testing.T) {
	var tests = []struct {
		input    payload.APS
		expected []byte
	}{
		{
			payload.APS{
				Alert: payload.Alert{Body: "Message received from Bob"},
			},
			[]byte(`{"aps":{"alert":"Message received from Bob"}}`),
		},
		{
			payload.APS{
				Alert: payload.Alert{Body: "You got your emails."},
				Badge: badge.New(9),
				Sound: "bingbong.aiff",
			},
			[]byte(`{"aps":{"alert":"You got your emails.","badge":9,"sound":"bingbong.aiff"}}`),
		},
		{
			payload.APS{ContentAvailable: true},
			[]byte(`{"aps":{"content-available":1}}`),
		},
		{
			payload.APS{
				Alert: payload.Alert{
					Title: "Message",
					Body:  "Message received from Bob",
				},
			},
			[]byte(`{"aps":{"alert":{"title":"Message","body":"Message received from Bob"}}}`),
		},
	}

	for _, tt := range tests {
		testPayload(t, tt.input, tt.expected)
	}
}

func TestCustomArray(t *testing.T) {
	p := payload.APS{Alert: payload.Alert{Body: "Message received from Bob"}}
	pm := p.Map()
	pm["acme2"] = []string{"bang", "whiz"}
	expected := []byte(`{"acme2":["bang","whiz"],"aps":{"alert":"Message received from Bob"}}`)
	testPayload(t, pm, expected)
}
