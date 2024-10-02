package main

import (
	"crypto/rand"
	"crypto/x509/pkix"
	"flag"
	"fmt"
	"log"
	"math"
	"math/big"
	"time"

	"github.com/ca0s/despiste/certificates"
)

const (
	ModeInitCA = "init-ca"
	ModeCert   = "cert"
	ModeRevoke = "revoke"
)

func main() {
	var (
		action string

		caFile       string
		caPublicFile string
		certFile     string

		serialNumber int64
		subject      string
		notBefore    string
		notAfter     string

		crlFile string

		tNotBefore time.Time
		tNotAfter  time.Time

		ValidModes = []string{ModeInitCA, ModeCert, ModeRevoke}
	)

	flag.StringVar(&action, "action", "", "Action to perform. Options are init-ca and cert")

	flag.StringVar(&caFile, "ca", "data/certs/cafull.pem", "File to write the CA PEM cert+key to")
	flag.StringVar(&caPublicFile, "capub", "data/certs/ca.pem", "File to write the CA PEM public cert to")

	flag.StringVar(&certFile, "cert", "", "File to write the cert PEM cert+key to")

	flag.StringVar(&crlFile, "crl", "data/certs/crl.pem", "File to store the CRL")

	flag.Int64Var(&serialNumber, "serial", 0, "Serial number for the new certificate. A random value is chosen if no value given")
	flag.StringVar(&subject, "subject", "", "Subject for the new certificate. Must match whatever name you will assign to your upstreams")
	flag.StringVar(&notBefore, "not-before", "", "Certificate validity start")
	flag.StringVar(&notAfter, "not-after", "", "Certificate validity end")

	flag.Parse()

	validMode := false
	for _, m := range ValidModes {
		if action == m {
			validMode = true
			break
		}
	}

	if !validMode {
		flag.Usage()
		return
	}

	if notBefore != "" {
		tNotBefore, err := time.Parse(time.RFC822, notBefore)
		if err != nil {
			log.Printf("invalid time in not-before: %s\n", err)
			return
		}

		if tNotBefore.Before(time.Now().Add(365 * 24 * time.Hour)) {
			log.Printf("WARN: not-before value is soon, may be a good idea to provide a later value\n")
		}
	} else {
		tNotBefore = time.Now()
	}

	if notAfter != "" {
		tNotAfter, err := time.Parse(time.RFC822, notAfter)
		if err != nil {
			log.Printf("invalid time in not-after: %s\n", err)
			return
		}

		if tNotAfter.Before(time.Now().Add(365 * 24 * time.Hour)) {
			log.Printf("WARN: not-after value is soon, may be a good idea to provide a later value\n")
		}
	} else {
		tNotAfter = time.Now().Add(2 * 365 * 24 * time.Hour)
	}

	if subject == "" && action != ModeRevoke {
		log.Printf("subject cannot be empty\n")
		return
	}

	if serialNumber == 0 {
		bn, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			log.Printf("could not create random serial: %s\n", err)
			return
		}

		serialNumber = bn.Int64()
	}

	if certFile == "" {
		certFile = fmt.Sprintf("data/certs/%s.pem", subject)
	}

	switch action {
	case "init-ca":
		newCA, newCAkey, err := certificates.GenerateCert(
			true, nil, nil,
			serialNumber, subject,
			tNotBefore, tNotAfter,
		)
		if err != nil {
			log.Printf("could not create CA certificate: %s\n", err.Error())
			return
		}

		err = certificates.WriteCertToFile(caFile, newCA, newCAkey)
		if err != nil {
			log.Printf("error writing CA certificate+key to %s: %s\n", certFile, err)
			return
		}

		err = certificates.WriteCertToFile(caPublicFile, newCA, nil)
		if err != nil {
			log.Printf("error writing CA public certificate to %s: %s\n", certFile, err)
			return
		}

	case "cert":
		caCert, caKey, err := certificates.ReadCert(caFile, true)
		if err != nil {
			log.Printf("could not read CA certificate: %s\n", err.Error())
			return
		}

		newCert, newKey, err := certificates.GenerateCert(
			false, caCert, caKey,
			serialNumber, subject,
			tNotBefore, tNotAfter,
		)
		if err != nil {
			log.Printf("error creating certificate: %s\n", err.Error())
			return
		}

		err = certificates.WriteCertToFile(certFile, newCert, newKey)
		if err != nil {
			log.Printf("error writing certificate to %s: %s\n", certFile, err)
			return
		}

	case "revoke":
		caCert, caKey, err := certificates.ReadCert(caFile, true)
		if err != nil {
			log.Printf("could not read CA certificate: %s\n", err.Error())
			return
		}

		currentCRL, err := certificates.ReadCRL(crlFile)
		if err != nil {
			log.Printf("creating a new CRL\n")

			currentCRL, err = certificates.CreateCRL(crlFile, caCert, caKey, big.NewInt(0), []pkix.RevokedCertificate{})
			if err != nil {
				log.Printf("could not create an empty CRL: %s\n", err)
				return
			}
		}

		revokedCert, _, err := certificates.ReadCert(certFile, false)
		if err != nil {
			log.Printf("could not read certificate at %s: %s\n", certFile, err)
			return
		}

		revokedSerial := revokedCert.SerialNumber

		for _, alreadyRevokedSerial := range currentCRL.TBSCertList.RevokedCertificates {
			if alreadyRevokedSerial.SerialNumber.Cmp(revokedCert.SerialNumber) == 0 {
				log.Printf("certificate is already revoked")
				return
			}
		}

		newRevokedList := append(
			currentCRL.TBSCertList.RevokedCertificates,
			pkix.RevokedCertificate{
				SerialNumber:   revokedSerial,
				RevocationTime: time.Now(),
				Extensions:     []pkix.Extension{},
			},
		)

		currentCRLNumber, err := certificates.GetCRLNumber(currentCRL)
		if err != nil {
			log.Printf("could not get current CRL number: %s\n", err)
			currentCRLNumber = big.NewInt(0)
		}

		_, err = certificates.CreateCRL(
			crlFile,
			caCert,
			caKey,
			currentCRLNumber.Add(currentCRLNumber, big.NewInt(1)),
			newRevokedList,
		)

		if err != nil {
			log.Printf("could not create CRL: %s\n", err)
			return
		}

		log.Printf("certificate revoked\n")

	default:
		flag.Usage()
		return
	}
}
