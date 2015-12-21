package payload_test

import (
	"testing"

	"github.com/RobotsAndPencils/buford/payload"
)

func TestBrowser(t *testing.T) {
	p := payload.Browser{
		Alert: payload.BrowserAlert{
			Title:  "Flight A998 Now Boarding",
			Body:   "Boarding has begun for Flight A998.",
			Action: "View",
		},
		URLArgs: []string{"boarding", "A998"},
	}
	expected := []byte(`{"aps":{"alert":{"title":"Flight A998 Now Boarding","body":"Boarding has begun for Flight A998.","action":"View"},"url-args":["boarding","A998"]}}`)
	testPayload(t, p, expected)
}
