package tls

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/pem"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureCert_GeneratesWhenNeitherExists(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := EnsureCert(certPath, keyPath); err != nil {
		t.Fatalf("EnsureCert failed: %v", err)
	}

	cert := parseCertFile(t, certPath)
	key := parseKeyFile(t, keyPath)

	if key.Curve != elliptic.P256() {
		t.Errorf("expected P-256 curve, got %v", key.Curve.Params().Name)
	}

	if !key.PublicKey.Equal(cert.PublicKey) {
		t.Error("certificate public key does not match private key")
	}
}

func TestEnsureCert_NoOpWhenBothExist(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := EnsureCert(certPath, keyPath); err != nil {
		t.Fatalf("initial EnsureCert failed: %v", err)
	}

	certBefore, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("read cert: %v", err)
	}
	keyBefore, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("read key: %v", err)
	}

	if err := EnsureCert(certPath, keyPath); err != nil {
		t.Fatalf("second EnsureCert failed: %v", err)
	}

	certAfter, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("read cert after: %v", err)
	}
	keyAfter, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("read key after: %v", err)
	}

	if string(certBefore) != string(certAfter) {
		t.Error("certificate file was modified on second call")
	}
	if string(keyBefore) != string(keyAfter) {
		t.Error("key file was modified on second call")
	}
}

func TestEnsureCert_ErrorWhenOnlyCertExists(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := os.WriteFile(certPath, []byte("dummy"), 0640); err != nil {
		t.Fatal(err)
	}

	err := EnsureCert(certPath, keyPath)
	if err == nil {
		t.Fatal("expected error when only cert exists")
	}
}

func TestEnsureCert_ErrorWhenOnlyKeyExists(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := os.WriteFile(keyPath, []byte("dummy"), 0600); err != nil {
		t.Fatal(err)
	}

	err := EnsureCert(certPath, keyPath)
	if err == nil {
		t.Fatal("expected error when only key exists")
	}
}

func TestEnsureCert_CertContainsCorrectSANs(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := EnsureCert(certPath, keyPath); err != nil {
		t.Fatalf("EnsureCert failed: %v", err)
	}

	cert := parseCertFile(t, certPath)

	hasLocalhost := false
	for _, name := range cert.DNSNames {
		if name == "localhost" {
			hasLocalhost = true
			break
		}
	}
	if !hasLocalhost {
		t.Errorf("cert DNSNames %v does not include localhost", cert.DNSNames)
	}

	hostname, _ := os.Hostname()
	if hostname != "" && hostname != "localhost" {
		hasHostname := false
		for _, name := range cert.DNSNames {
			if name == hostname {
				hasHostname = true
				break
			}
		}
		if !hasHostname {
			t.Errorf("cert DNSNames %v does not include hostname %q", cert.DNSNames, hostname)
		}
	}

	hasLoopback := false
	for _, ip := range cert.IPAddresses {
		if ip.Equal(net.IPv4(127, 0, 0, 1)) {
			hasLoopback = true
			break
		}
	}
	if !hasLoopback {
		t.Errorf("cert IPAddresses %v does not include 127.0.0.1", cert.IPAddresses)
	}
}

func TestEnsureCert_CreatesSubdirectories(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "sub", "dir", "server.crt")
	keyPath := filepath.Join(dir, "other", "dir", "server.key")

	if err := EnsureCert(certPath, keyPath); err != nil {
		t.Fatalf("EnsureCert failed: %v", err)
	}

	if _, err := os.Stat(certPath); err != nil {
		t.Errorf("cert file not created: %v", err)
	}
	if _, err := os.Stat(keyPath); err != nil {
		t.Errorf("key file not created: %v", err)
	}
}

func parseCertFile(t *testing.T, path string) *x509.Certificate {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cert file: %v", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		t.Fatal("failed to decode PEM block from cert file")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}
	return cert
}

func parseKeyFile(t *testing.T, path string) *ecdsa.PrivateKey {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read key file: %v", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		t.Fatal("failed to decode PEM block from key file")
	}
	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("parse EC private key: %v", err)
	}
	return key
}
