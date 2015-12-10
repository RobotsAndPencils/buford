package buford

import "testing"

func TestValidCert(t *testing.T) {
	const name = "/Users/nathany/src/github.com/RobotsAndPencils/nx-client/ruby/Pushy.p12"

	_, err := LoadCert(name, "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestExpiredCert(t *testing.T) {
	const name = "/Users/nathany/src/github.com/RobotsAndPencils/nx-client/ruby/Pushy-expired.p12"

	_, err := LoadCert(name, "")
	if err != ErrExpiredCert {
		t.Fatal("Expected expired cert error, got", err)
	}
}

func TestMissingFile(t *testing.T) {
	_, err := LoadCert("hide-and-seek.p12", "")
	if err == nil {
		t.Fatal("Expected file not found, got", err)
	}
}
