package payload_test

import (
	"testing"

	"github.com/RobotsAndPencils/buford/payload"
)

func TestMDM(t *testing.T) {
	p := payload.MDM{"00000000-1111-3333-4444-555555555555"}
	expected := []byte(`{"mdm":"00000000-1111-3333-4444-555555555555"}`)
	testPayload(t, p, expected)
}
