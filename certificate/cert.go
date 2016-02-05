// Package certificate loads Push Services certificates exported from your
// Keychain in Personal Information Exchange format (*.p12).
package certificate

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"

	"golang.org/x/crypto/pkcs12"
)

// Certificate errors
var (
	ErrExpired = errors.New("certificate has expired or is not yet valid")
)

// Load a .p12 certificate from disk.
func Load(filename, password string) (*x509.Certificate, *rsa.PrivateKey, error) {
	p12, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to load %s: %v", filename, err)
	}
	return Decode(p12, password)
}

// Decode and verify an in memory .p12 certificate (DER binary format).
func Decode(p12 []byte, password string) (*x509.Certificate, *rsa.PrivateKey, error) {
	// decode an x509.Certificate to verify
	privateKey, cert, err := pkcs12.Decode(p12, password)
	if err != nil {
		return nil, nil, err
	}
	if err := verify(cert); err != nil {
		return nil, nil, err
	}

	// assert that private key is RSA
	priv, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, errors.New("expected RSA private key type")
	}
	return cert, priv, nil
}

// TLS wraps an x509 certificate as a tls.Certificate.
func TLS(cert *x509.Certificate, privateKey *rsa.PrivateKey) tls.Certificate {
	return tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  privateKey,
		Leaf:        cert,
	}
}

// verify checks if a certificate has expired
func verify(cert *x509.Certificate) error {
	_, err := cert.Verify(x509.VerifyOptions{})
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case x509.CertificateInvalidError:
		switch e.Reason {
		case x509.Expired:
			return ErrExpired
		default:
			return err
		}
	case x509.UnknownAuthorityError:
		// Apple cert isn't in the cert pool
		// ignoring this error
		return nil
	default:
		return err
	}
}
