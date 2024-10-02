package certificates

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"os"
	"time"

	"github.com/pkg/errors"
)

func GenerateCert(ca bool, parent *x509.Certificate, parentKey *ecdsa.PrivateKey, serialNumber int64, subject string, notBefore time.Time, notAfter time.Time) (*pem.Block, *pem.Block, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate private key")
	}

	template := x509.Certificate{
		SerialNumber:          new(big.Int).SetInt64(serialNumber),
		Subject:               pkix.Name{CommonName: subject},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	template.DNSNames = []string{subject}

	if ca {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
		template.KeyUsage |= x509.KeyUsageCRLSign
	}

	signingKey := parentKey

	if parent == nil {
		parent = &template
		signingKey = key
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, parent, &key.PublicKey, signingKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to create certificate")
	}

	b, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to marshal ecdsa")
	}
	return &pem.Block{Type: "CERTIFICATE", Bytes: cert},
		&pem.Block{Type: "ECDSA PRIVATE KEY", Bytes: b},
		nil
}

func ReadCert(path string, withKey bool) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	data, err := io.ReadAll(fd)
	if err != nil {
		return nil, nil, err
	}

	crt, remainingData := pem.Decode(data)
	if crt == nil {
		return nil, nil, errors.New("could not decode crt")
	}

	xcert, err := x509.ParseCertificate(crt.Bytes)
	if err != nil {
		return nil, nil, err
	}

	if !withKey {
		return xcert, nil, nil
	}

	key, _ := pem.Decode(remainingData)
	if key == nil {
		return nil, nil, errors.New("could not decode key")
	}

	xkey, err := x509.ParseECPrivateKey(key.Bytes)
	if err != nil {
		return nil, nil, err
	}

	return xcert, xkey, nil
}

func WriteCertToFile(path string, crt *pem.Block, key *pem.Block) error {
	fd, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		return err
	}

	defer fd.Close()

	if crt != nil {
		err = pem.Encode(fd, crt)
		if err != nil {
			return err
		}
	}

	if key != nil {
		err = pem.Encode(fd, key)
	}

	return err
}

func ReadCRL(path string) (*pkix.CertificateList, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer fd.Close()

	data, err := io.ReadAll(fd)
	if err != nil {
		return nil, err
	}

	return x509.ParseCRL(data)
}

func CreateCRL(path string, crt *x509.Certificate, key *ecdsa.PrivateKey, number *big.Int, revokedCerts []pkix.RevokedCertificate) (*pkix.CertificateList, error) {
	crl, err := x509.CreateRevocationList(
		rand.Reader,
		&x509.RevocationList{
			SignatureAlgorithm:  x509.ECDSAWithSHA512,
			RevokedCertificates: revokedCerts,
			Number:              number,
			ThisUpdate:          time.Now(),
			NextUpdate:          time.Now().Add(30 * 24 * 12 * time.Hour), // 1 year
			ExtraExtensions:     []pkix.Extension{},
		},
		crt,
		key,
	)

	if err != nil {
		return nil, err
	}

	fd, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		return nil, err
	}

	defer fd.Close()

	pemBlock := pem.Block{
		Type:  "X509 CRL",
		Bytes: crl,
	}

	err = pem.Encode(fd, &pemBlock)
	if err != nil {
		return nil, err
	}

	return x509.ParseCRL(crl)
}

func GetCRLNumber(crl *pkix.CertificateList) (*big.Int, error) {
	var oidExtensionCRLNumber = []int{2, 5, 29, 20}
	var number big.Int

	for _, ext := range crl.TBSCertList.Extensions {
		if ext.Id.Equal(oidExtensionCRLNumber) {
			_, err := asn1.Unmarshal(ext.Value, &number)
			if err != nil {
				return nil, err
			}

			return &number, nil
		}
	}

	return nil, fmt.Errorf("CRL does not contain a valid number")
}
