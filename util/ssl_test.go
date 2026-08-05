package util

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckAndCreateSSLForHostsCreatesServerCertificate(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "tls", "server.crt")
	keyPath := filepath.Join(dir, "tls", "server.key")
	if err := CheckAndCreateSSLForHosts(certPath, keyPath, "localhost", "127.0.0.1"); err != nil {
		t.Fatalf("creating certificate: %v", err)
	}

	pemBytes, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("reading certificate: %v", err)
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		t.Fatal("decoding certificate PEM")
	}
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parsing certificate: %v", err)
	}
	if err := certificate.VerifyHostname("localhost"); err != nil {
		t.Fatalf("certificate is not valid for localhost: %v", err)
	}
	if err := certificate.VerifyHostname("127.0.0.1"); err != nil {
		t.Fatalf("certificate is not valid for 127.0.0.1: %v", err)
	}
	keyInfo, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("checking private key: %v", err)
	}
	if keyInfo.Mode().Perm() != 0600 {
		t.Fatalf("expected private-key mode 0600, got %04o", keyInfo.Mode().Perm())
	}
}
