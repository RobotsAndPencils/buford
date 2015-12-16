package buford

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"

	"golang.org/x/crypto/pkcs12"
)

// Certificate errors
var (
	ErrExpiredCert = errors.New("certificate has expired or is not yet valid")
)

// LoadCert loads a .p12 certificate from disk.
func LoadCert(name, password string) (tls.Certificate, error) {
	p12, err := ioutil.ReadFile(name)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("Unable to load %s: %v", name, err)
	}
	return DecodeCert(p12, password)
}

// DecodeCert decodes an in memory .p12 certificate.
func DecodeCert(p12 []byte, password string) (tls.Certificate, error) {
	// decode an x509.Certificate to verify
	privateKey, cert, err := pkcs12.Decode(p12, password)
	if err != nil {
		return tls.Certificate{}, err
	}
	if err := verify(cert); err != nil {
		return tls.Certificate{}, err
	}

	// wrap in a tls certificate
	return tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  privateKey,
		Leaf:        cert,
	}, nil
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
			return ErrExpiredCert
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
