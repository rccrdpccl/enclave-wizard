package tlscert

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

func TestEnsureSelfSigned_GeneratesWhenMissing(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := EnsureSelfSigned(certPath, keyPath); err != nil {
		t.Fatalf("EnsureSelfSigned failed: %v", err)
	}

	if _, err := os.Stat(certPath); err != nil {
		t.Fatalf("cert file not created: %v", err)
	}
	if _, err := os.Stat(keyPath); err != nil {
		t.Fatalf("key file not created: %v", err)
	}

	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("stat key: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("key permissions = %o, want 0600", perm)
	}
}

func TestEnsureSelfSigned_NoOpWhenBothExist(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	certContent := []byte("existing-cert")
	keyContent := []byte("existing-key")
	os.WriteFile(certPath, certContent, 0640)
	os.WriteFile(keyPath, keyContent, 0600)

	if err := EnsureSelfSigned(certPath, keyPath); err != nil {
		t.Fatalf("EnsureSelfSigned failed: %v", err)
	}

	gotCert, _ := os.ReadFile(certPath)
	gotKey, _ := os.ReadFile(keyPath)
	if string(gotCert) != string(certContent) {
		t.Error("cert file was modified when both files already existed")
	}
	if string(gotKey) != string(keyContent) {
		t.Error("key file was modified when both files already existed")
	}
}

func TestEnsureSelfSigned_CertProperties(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := EnsureSelfSigned(certPath, keyPath); err != nil {
		t.Fatalf("EnsureSelfSigned failed: %v", err)
	}

	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("reading cert: %v", err)
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("failed to decode PEM block from cert")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parsing certificate: %v", err)
	}

	if cert.PublicKeyAlgorithm != x509.ECDSA {
		t.Errorf("public key algorithm = %v, want ECDSA", cert.PublicKeyAlgorithm)
	}
	pub, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("public key is not ECDSA")
	}
	if pub.Curve != elliptic.P256() {
		t.Errorf("curve = %v, want P-256", pub.Curve.Params().Name)
	}

	if cert.Issuer.CommonName != cert.Subject.CommonName {
		t.Errorf("not self-signed: issuer CN = %q, subject CN = %q", cert.Issuer.CommonName, cert.Subject.CommonName)
	}

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

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("reading key: %v", err)
	}
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		t.Fatal("failed to decode PEM block from key")
	}
	if _, err := x509.ParseECPrivateKey(keyBlock.Bytes); err != nil {
		t.Fatalf("parsing EC private key: %v", err)
	}
}

func TestEnsureSelfSigned_CreatesParentDirectories(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "sub", "dir", "server.crt")
	keyPath := filepath.Join(dir, "sub", "dir", "server.key")

	if err := EnsureSelfSigned(certPath, keyPath); err != nil {
		t.Fatalf("EnsureSelfSigned failed: %v", err)
	}

	if _, err := os.Stat(certPath); err != nil {
		t.Fatalf("cert file not created at nested path: %v", err)
	}
	if _, err := os.Stat(keyPath); err != nil {
		t.Fatalf("key file not created at nested path: %v", err)
	}
}
