package util

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/csv"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"regexp"
	"time"

	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/jordan-wright/email"
)

var (
	firstNameRegex  = regexp.MustCompile(`(?i)first[\s_-]*name`)
	lastNameRegex   = regexp.MustCompile(`(?i)last[\s_-]*name`)
	emailRegex      = regexp.MustCompile(`(?i)email`)
	phoneRegex      = regexp.MustCompile(`(?i)phone`)
	positionRegex   = regexp.MustCompile(`(?i)position`)
	customRegex     = regexp.MustCompile(`(?i)custom`)
	departmentRegex = regexp.MustCompile(`(?i)depart(?:ment)?`)
	companyRegex    = regexp.MustCompile(`(?i)compan(?:y|ies)`)
	cityRegex       = regexp.MustCompile(`(?i)city`)
	stateRegex      = regexp.MustCompile(`(?i)state`)
	countryRegex    = regexp.MustCompile(`(?i)country`)
	unitRegex       = regexp.MustCompile(`(?i)unit`)
	tagsRegex       = regexp.MustCompile(`(?i)tags?`)
)

// ParseMail takes in an HTTP Request and returns an Email object
// TODO: This function will likely be changed to take in a []byte
func ParseMail(r *http.Request) (email.Email, error) {
	e := email.Email{}
	m, err := mail.ReadMessage(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	body, err := ioutil.ReadAll(m.Body)
	e.HTML = body
	return e, err
}

// ParseCSV contains the logic to parse the user provided csv file containing Target entries
func ParseCSV(r *http.Request) ([]models.Target, error) {
	mr, err := r.MultipartReader()
	ts := []models.Target{}
	if err != nil {
		return ts, err
	}
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		// Skip the "submit" part
		if part.FileName() == "" {
			continue
		}
		defer part.Close()
		reader := csv.NewReader(part)
		reader.TrimLeadingSpace = true
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		fi := -1
		li := -1
		ei := -1
		phi := -1
		pi := -1
		ci := -1
		di := -1
		coi := -1
		cti := -1
		sti := -1
		coui := -1
		ui := -1
		tgi := -1
		fn := ""
		ln := ""
		ea := ""
		ph := ""
		ps := ""
		cm := ""
		dp := ""
		co := ""
		ct := ""
		st := ""
		cou := ""
		un := ""
		tg := ""
		for i, v := range record {
			switch {
			case firstNameRegex.MatchString(v):
				fi = i
			case lastNameRegex.MatchString(v):
				li = i
			case emailRegex.MatchString(v):
				ei = i
			case phoneRegex.MatchString(v):
				phi = i
			case positionRegex.MatchString(v):
				pi = i
			case customRegex.MatchString(v):
				ci = i
			case departmentRegex.MatchString(v):
				di = i
			case companyRegex.MatchString(v):
				coi = i
			case cityRegex.MatchString(v):
				cti = i
			case stateRegex.MatchString(v):
				sti = i
			case countryRegex.MatchString(v):
				coui = i
			case unitRegex.MatchString(v):
				ui = i
			case tagsRegex.MatchString(v):
				tgi = i
			}
		}
		if fi == -1 && li == -1 && ei == -1 && pi == -1 {
			continue
		}
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if fi != -1 && len(record) > fi {
				fn = record[fi]
			}
			if li != -1 && len(record) > li {
				ln = record[li]
			}
			if ei != -1 && len(record) > ei {
				// Try to parse as email
				csvEmail, err := mail.ParseAddress(record[ei])
				if err == nil {
					ea = csvEmail.Address
				}
			}

			if phi != -1 && len(record) > phi {
				ph = record[phi]
			}
			if pi != -1 && len(record) > pi {
				ps = record[pi]
			}
			if ci != -1 && len(record) > ci {
				cm = record[ci]
			}
			if di != -1 && len(record) > di {
				dp = record[di]
			}
			if coi != -1 && len(record) > coi {
				co = record[coi]
			}
			if cti != -1 && len(record) > cti {
				ct = record[cti]
			}
			if sti != -1 && len(record) > sti {
				st = record[sti]
			}
			if coui != -1 && len(record) > coui {
				cou = record[coui]
			}
			if ui != -1 && len(record) > ui {
				un = record[ui]
			}
			if tgi != -1 && len(record) > tgi {
				tg = record[tgi]
			}
			t := models.Target{
				BaseRecipient: models.BaseRecipient{
					FirstName:  fn,
					LastName:   ln,
					Email:      ea,
					Phone:      ph,
					Position:   ps,
					Custom:     cm,
					Department: dp,
					Company:    co,
					City:       ct,
					State:      st,
					Country:    cou,
					Unit:       un,
					Tags:       tg,
				},
			}
			ts = append(ts, t)
		}
	}
	return ts, nil
}

// CheckAndCreateSSL creates a self-signed certificate when both files are absent.
// Prefer CheckAndCreateSSLForHosts when the certificate is served to clients.
func CheckAndCreateSSL(cp string, kp string) error {
	return CheckAndCreateSSLForHosts(cp, kp)
}

// CheckAndCreateSSLForHosts creates a self-signed server certificate with
// subject alternative names for the supplied DNS names and IP addresses. It
// never replaces an existing certificate: certificate rotation remains an
// explicit deployment action.
func CheckAndCreateSSLForHosts(cp string, kp string, hosts ...string) error {
	certExists := fileExists(cp)
	keyExists := fileExists(kp)
	if certExists && keyExists {
		return nil
	}
	if certExists != keyExists {
		return fmt.Errorf("incomplete TLS key pair: both %s and %s must exist or be absent", cp, kp)
	}
	if err := os.MkdirAll(filepath.Dir(cp), 0750); err != nil {
		return fmt.Errorf("creating TLS certificate directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(kp), 0750); err != nil {
		return fmt.Errorf("creating TLS key directory: %w", err)
	}

	log.Info("Creating new self-signed TLS certificate")

	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return fmt.Errorf("error generating tls private key: %v", err)
	}

	notBefore := time.Now()
	// Generate a certificate that lasts for 10 years
	notAfter := notBefore.Add(10 * 365 * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)

	if err != nil {
		return fmt.Errorf("tls certificate generation: failed to generate a random serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"EthPhish development"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	for _, host := range hosts {
		if ip := net.ParseIP(host); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else if host != "" {
			template.DNSNames = append(template.DNSNames, host)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, priv.Public(), priv)
	if err != nil {
		return fmt.Errorf("tls certificate generation: failed to create certificate: %s", err)
	}

	certOut, err := os.Create(cp)
	if err != nil {
		return fmt.Errorf("tls certificate generation: failed to open %s for writing: %s", cp, err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(kp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("tls certificate generation: failed to open %s for writing", kp)
	}

	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return fmt.Errorf("tls certificate generation: unable to marshal ECDSA private key: %v", err)
	}

	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	keyOut.Close()

	log.Info("TLS Certificate Generation complete")
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
