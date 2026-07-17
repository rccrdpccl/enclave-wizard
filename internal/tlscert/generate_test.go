package tlscert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/pem"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestEnsureCert_GeneratesFiles(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := EnsureCert(certPath, keyPath, true); err != nil {
		t.Fatalf("EnsureCert() error = %v", err)
	}

	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("reading cert: %v", err)
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("reading key: %v", err)
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		t.Fatal("failed to decode cert PEM")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		t.Fatalf("parsing certificate: %v", err)
	}

	if !containsString(cert.DNSNames, "localhost") {
		t.Errorf("cert DNSNames %v missing 'localhost'", cert.DNSNames)
	}

	hasLoopback := false
	for _, ip := range cert.IPAddresses {
		if ip.Equal(net.IPv4(127, 0, 0, 1)) {
			hasLoopback = true
			break
		}
	}
	if !hasLoopback {
		t.Errorf("cert IPAddresses %v missing 127.0.0.1", cert.IPAddresses)
	}

	if cert.PublicKeyAlgorithm != x509.ECDSA {
		t.Errorf("expected ECDSA key, got %v", cert.PublicKeyAlgorithm)
	}
	pub, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("public key is not *ecdsa.PublicKey")
	}
	if pub.Curve != elliptic.P256() {
		t.Errorf("expected P-256 curve, got %v", pub.Curve.Params().Name)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		t.Fatal("failed to decode key PEM")
	}
	if _, err := x509.ParseECPrivateKey(keyBlock.Bytes); err != nil {
		t.Fatalf("parsing EC private key: %v", err)
	}

	// Verify self-signed
	roots := x509.NewCertPool()
	roots.AddCert(cert)
	if _, err := cert.Verify(x509.VerifyOptions{Roots: roots}); err != nil {
		t.Errorf("certificate self-verification failed: %v", err)
	}
}

func TestEnsureCert_SkipsIfExists(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := EnsureCert(certPath, keyPath, true); err != nil {
		t.Fatalf("first EnsureCert() error = %v", err)
	}

	certBefore, _ := os.ReadFile(certPath)
	keyBefore, _ := os.ReadFile(keyPath)

	if err := EnsureCert(certPath, keyPath, true); err != nil {
		t.Fatalf("second EnsureCert() error = %v", err)
	}

	certAfter, _ := os.ReadFile(certPath)
	keyAfter, _ := os.ReadFile(keyPath)

	if string(certBefore) != string(certAfter) {
		t.Error("cert file was regenerated when it should have been skipped")
	}
	if string(keyBefore) != string(keyAfter) {
		t.Error("key file was regenerated when it should have been skipped")
	}
}

func TestEnsureCert_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "sub", "tls")
	certPath := filepath.Join(nested, "server.crt")
	keyPath := filepath.Join(nested, "server.key")

	if err := EnsureCert(certPath, keyPath, true); err != nil {
		t.Fatalf("EnsureCert() error = %v", err)
	}

	if _, err := os.Stat(certPath); err != nil {
		t.Errorf("cert file not created: %v", err)
	}
	if _, err := os.Stat(keyPath); err != nil {
		t.Errorf("key file not created: %v", err)
	}
}

func TestEnsureCert_KeyPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permissions not applicable on Windows")
	}

	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := EnsureCert(certPath, keyPath, true); err != nil {
		t.Fatalf("EnsureCert() error = %v", err)
	}

	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("stat key file: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("key file permissions = %o, want 0600", perm)
	}
}

func TestEnsureCert_UserProvidedCertsUsed(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := EnsureCert(certPath, keyPath, true); err != nil {
		t.Fatalf("generating initial certs: %v", err)
	}

	certBefore, _ := os.ReadFile(certPath)

	if err := EnsureCert(certPath, keyPath, false); err != nil {
		t.Fatalf("EnsureCert(selfSigned=false) with existing files: %v", err)
	}

	certAfter, _ := os.ReadFile(certPath)
	if string(certBefore) != string(certAfter) {
		t.Error("cert file was modified when user-provided certs should be left alone")
	}
}

func TestEnsureCert_UserProvidedCertMissing(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	err := EnsureCert(certPath, keyPath, false)
	if err == nil {
		t.Fatal("expected error when cert file does not exist with selfSigned=false")
	}
	if !strings.Contains(err.Error(), "TLS certificate not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestEnsureCert_UserProvidedKeyMissing(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	os.WriteFile(certPath, []byte("cert"), 0644)

	err := EnsureCert(certPath, keyPath, false)
	if err == nil {
		t.Fatal("expected error when key file does not exist with selfSigned=false")
	}
	if !strings.Contains(err.Error(), "TLS key not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func containsString(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
