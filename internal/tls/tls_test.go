package tls

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEnsureCert_GeneratesWhenMissing(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := EnsureCert(certPath, keyPath); err != nil {
		t.Fatalf("EnsureCert: %v", err)
	}

	pair, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		t.Fatalf("LoadX509KeyPair: %v", err)
	}

	cert, err := x509.ParseCertificate(pair.Certificate[0])
	if err != nil {
		t.Fatalf("ParseCertificate: %v", err)
	}

	if cert.Subject.CommonName != "enclave-wizard" {
		t.Errorf("CN = %q, want %q", cert.Subject.CommonName, "enclave-wizard")
	}

	if !containsString(cert.DNSNames, "localhost") {
		t.Error("SAN missing DNS:localhost")
	}

	if !containsIP(cert.IPAddresses, net.IPv4(127, 0, 0, 1)) {
		t.Error("SAN missing IP:127.0.0.1")
	}

	expectedExpiry := time.Now().Add(365 * 24 * time.Hour)
	if cert.NotAfter.Before(expectedExpiry.Add(-1*time.Minute)) ||
		cert.NotAfter.After(expectedExpiry.Add(1*time.Minute)) {
		t.Errorf("NotAfter = %v, want ~%v", cert.NotAfter, expectedExpiry)
	}
}

func TestEnsureCert_SkipsWhenExists(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := EnsureCert(certPath, keyPath); err != nil {
		t.Fatalf("initial EnsureCert: %v", err)
	}

	certBefore, _ := os.ReadFile(certPath)
	keyBefore, _ := os.ReadFile(keyPath)

	if err := EnsureCert(certPath, keyPath); err != nil {
		t.Fatalf("second EnsureCert: %v", err)
	}

	certAfter, _ := os.ReadFile(certPath)
	keyAfter, _ := os.ReadFile(keyPath)

	if string(certBefore) != string(certAfter) {
		t.Error("cert file was modified on second call")
	}
	if string(keyBefore) != string(keyAfter) {
		t.Error("key file was modified on second call")
	}
}

func TestEnsureCert_CreatesParentDirectories(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "a", "b", "server.crt")
	keyPath := filepath.Join(dir, "a", "b", "server.key")

	if err := EnsureCert(certPath, keyPath); err != nil {
		t.Fatalf("EnsureCert: %v", err)
	}

	if !fileExists(certPath) {
		t.Error("cert file not created")
	}
	if !fileExists(keyPath) {
		t.Error("key file not created")
	}
}

func TestEnsureCert_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := EnsureCert(certPath, keyPath); err != nil {
		t.Fatalf("EnsureCert: %v", err)
	}

	keyInfo, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("stat key: %v", err)
	}
	if mode := keyInfo.Mode().Perm(); mode != 0600 {
		t.Errorf("key permission = %o, want 0600", mode)
	}

	certInfo, err := os.Stat(certPath)
	if err != nil {
		t.Fatalf("stat cert: %v", err)
	}
	if mode := certInfo.Mode().Perm(); mode != 0640 {
		t.Errorf("cert permission = %o, want 0640", mode)
	}
}

func containsString(ss []string, target string) bool {
	for _, s := range ss {
		if s == target {
			return true
		}
	}
	return false
}

func containsIP(ips []net.IP, target net.IP) bool {
	for _, ip := range ips {
		if ip.Equal(target) {
			return true
		}
	}
	return false
}
