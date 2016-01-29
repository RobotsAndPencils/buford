// Package pushpackage creates push packages for website push notifications.
package pushpackage

import (
	"archive/zip"
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"io"
	"log"
	"path"

	"github.com/st3fan/gocrypto/pkcs7"
)

// New push package.
func New(w io.Writer, website *Website, iconset IconSet, cert *x509.Certificate, key *rsa.PrivateKey) error {
	z := zip.NewWriter(w)

	// manifest is a map of relative file paths to their SHA checksums
	manifest := make(map[string]string, len(iconset)+1)

	b, err := json.Marshal(website)
	if err != nil {
		return err
	}
	r := bytes.NewReader(b)

	zf, err := z.Create("website.json")
	if err != nil {
		return err
	}
	checksum, err := copyAndChecksum(zf, r)
	manifest["website.json"] = checksum

	for _, icon := range iconset {
		// NOTE: only forward slashes are allowed in zip files
		// (therefore using path rather than filepath)
		name := path.Join(iconDirectory, icon.Name)

		zf, err := z.Create(name)
		if err != nil {
			return err
		}
		checksum, err := copyAndChecksum(zf, icon.Reader)
		manifest[name] = checksum
	}

	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return err
	}

	log.Println(string(manifestBytes))

	zf, err = z.Create("manifest.json")
	if err != nil {
		return err
	}
	zf.Write(manifestBytes)

	// sign manifest.json with PKCS #7
	// and add signature to the zip file
	zf, err = z.Create("signature")
	if err != nil {
		return err
	}
	sig, err := pkcs7.Sign(bytes.NewReader(manifestBytes), cert, key)
	if err != nil {
		return err
	}
	zf.Write(sig)

	return z.Close()
}
