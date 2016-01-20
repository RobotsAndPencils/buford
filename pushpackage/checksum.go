package pushpackage

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
)

// copyAndChecksum calculates a checksum while writing to another output
func copyAndChecksum(w io.Writer, r io.Reader) (checksum string, err error) {
	h := sha1.New()
	mw := io.MultiWriter(w, h)
	_, err = io.Copy(mw, r)
	if err != nil {
		return "", err
	}
	checksum = hex.EncodeToString(h.Sum(nil))
	return checksum, nil
}
