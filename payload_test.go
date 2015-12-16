package buford_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/RobotsAndPencils/buford"
)

func TestSimpleAlert(t *testing.T) {
	p := buford.Payload{Alert: buford.Alert{Body: "Message received from Bob"}}
	expected := []byte(`{"aps":{"alert":"Message received from Bob"}}`)
	testPayload(t, p, expected)
}

func TestBadgeAndSound(t *testing.T) {
	p := buford.Payload{
		Alert: buford.Alert{Body: "You got your emails."},
		Badge: buford.NewBadge(9),
		Sound: "bingbong.aiff",
	}
	expected := []byte(`{"aps":{"alert":"You got your emails.","badge":9,"sound":"bingbong.aiff"}}`)
	testPayload(t, p, expected)
}

func TestContentAvailable(t *testing.T) {
	p := buford.Payload{
		ContentAvailable: true,
	}
	expected := []byte(`{"aps":{"content-available":1}}`)
	testPayload(t, p, expected)
}

func TestAlertDictionary(t *testing.T) {
	p := buford.Payload{Alert: buford.Alert{Title: "Message", Body: "Message received from Bob"}}
	expected := []byte(`{"aps":{"alert":{"title":"Message","body":"Message received from Bob"}}}`)
	testPayload(t, p, expected)
}

func TestMDM(t *testing.T) {
	p := buford.MDMPayload{"00000000-1111-3333-4444-555555555555"}
	expected := []byte(`{"mdm":"00000000-1111-3333-4444-555555555555"}`)
	testPayload(t, p, expected)
}

func testPayload(t *testing.T, p interface{}, expected []byte) {
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}
	if !reflect.DeepEqual(b, expected) {
		t.Errorf("Expected %s, got %s", expected, b)
	}
}
