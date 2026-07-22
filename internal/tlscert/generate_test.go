package tlscert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestGenerateSelfSigned_CreatesFiles(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := GenerateSelfSigned(certPath, keyPath); err != nil {
		t.Fatalf("GenerateSelfSigned() error = %v", err)
	}

	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("reading cert: %v", err)
	}
	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		t.Fatal("cert PEM: expected CERTIFICATE block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parsing cert: %v", err)
	}

	hasLocalhost := false
	for _, name := range cert.DNSNames {
		if name == "localhost" {
			hasLocalhost = true
		}
	}
	if !hasLocalhost {
		t.Errorf("SANs missing localhost, got %v", cert.DNSNames)
	}

	hasLoopback := false
	for _, ip := range cert.IPAddresses {
		if ip.Equal(net.ParseIP("127.0.0.1")) {
			hasLoopback = true
		}
	}
	if !hasLoopback {
		t.Errorf("SANs missing 127.0.0.1, got %v", cert.IPAddresses)
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("reading key: %v", err)
	}
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil || keyBlock.Type != "EC PRIVATE KEY" {
		t.Fatalf("key PEM: expected EC PRIVATE KEY block, got %q", keyBlock.Type)
	}
}

func TestGenerateSelfSigned_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := GenerateSelfSigned(certPath, keyPath); err != nil {
		t.Fatalf("first call: %v", err)
	}
	firstCert, _ := os.ReadFile(certPath)

	if err := GenerateSelfSigned(certPath, keyPath); err != nil {
		t.Fatalf("second call: %v", err)
	}
	secondCert, _ := os.ReadFile(certPath)

	if string(firstCert) == string(secondCert) {
		t.Error("expected overwrite to produce different cert")
	}
}

func TestGenerateSelfSigned_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "deep", "nested", "tls")
	certPath := filepath.Join(nested, "server.crt")
	keyPath := filepath.Join(nested, "server.key")

	if err := GenerateSelfSigned(certPath, keyPath); err != nil {
		t.Fatalf("GenerateSelfSigned() error = %v", err)
	}

	if _, err := os.Stat(certPath); err != nil {
		t.Errorf("cert file not created: %v", err)
	}
}

func TestGenerateSelfSigned_KeyPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permissions not meaningful on Windows")
	}

	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := GenerateSelfSigned(certPath, keyPath); err != nil {
		t.Fatalf("GenerateSelfSigned() error = %v", err)
	}

	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("stat key: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("key permissions = %o, want 0600", perm)
	}
}

func generateTestCert(t *testing.T, dir string, notBefore, notAfter time.Time) (certPath, keyPath string) {
	t.Helper()
	certPath = filepath.Join(dir, "server.crt")
	keyPath = filepath.Join(dir, "server.key")

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generating test key: %v", err)
	}

	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature,
		DNSNames:     []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("creating test cert: %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		t.Fatalf("writing test cert: %v", err)
	}

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshaling test key: %v", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		t.Fatalf("writing test key: %v", err)
	}

	return certPath, keyPath
}

func TestEnsureCert_GeneratesWhenMissing(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := EnsureCert(certPath, keyPath); err != nil {
		t.Fatalf("EnsureCert() error = %v", err)
	}

	if _, err := os.Stat(certPath); err != nil {
		t.Error("cert not generated")
	}
	if _, err := os.Stat(keyPath); err != nil {
		t.Error("key not generated")
	}
}

func TestEnsureCert_ValidatesExisting(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := generateTestCert(t, dir,
		time.Now().Add(-time.Hour),
		time.Now().Add(365*24*time.Hour),
	)
	beforeCert, _ := os.ReadFile(certPath)

	if err := EnsureCert(certPath, keyPath); err != nil {
		t.Fatalf("EnsureCert() error = %v", err)
	}

	afterCert, _ := os.ReadFile(certPath)
	if string(beforeCert) != string(afterCert) {
		t.Error("valid cert was overwritten")
	}
}

func TestEnsureCert_RejectsExpired(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := generateTestCert(t, dir,
		time.Now().Add(-48*time.Hour),
		time.Now().Add(-1*time.Hour),
	)

	err := EnsureCert(certPath, keyPath)
	if err == nil {
		t.Fatal("expected error for expired cert")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("error should mention 'expired', got: %v", err)
	}
}

func TestEnsureCert_RejectsMismatchedKeyPair(t *testing.T) {
	dir := t.TempDir()
	certPath, _ := generateTestCert(t, dir,
		time.Now().Add(-time.Hour),
		time.Now().Add(365*24*time.Hour),
	)

	otherKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	otherKeyDER, _ := x509.MarshalECPrivateKey(otherKey)
	otherKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: otherKeyDER})
	keyPath := filepath.Join(dir, "server.key")
	os.WriteFile(keyPath, otherKeyPEM, 0600)

	err := EnsureCert(certPath, keyPath)
	if err == nil {
		t.Fatal("expected error for mismatched key pair")
	}
}

func TestEnsureCert_RejectsPartialCertOnly(t *testing.T) {
	dir := t.TempDir()
	certPath, _ := generateTestCert(t, dir,
		time.Now().Add(-time.Hour),
		time.Now().Add(365*24*time.Hour),
	)
	os.Remove(filepath.Join(dir, "server.key"))

	err := EnsureCert(certPath, filepath.Join(dir, "server.key"))
	if err == nil {
		t.Fatal("expected error when cert exists but key missing")
	}
	if !strings.Contains(err.Error(), "key") {
		t.Errorf("error should mention key, got: %v", err)
	}
}

func TestEnsureCert_RejectsPartialKeyOnly(t *testing.T) {
	dir := t.TempDir()
	_, keyPath := generateTestCert(t, dir,
		time.Now().Add(-time.Hour),
		time.Now().Add(365*24*time.Hour),
	)
	os.Remove(filepath.Join(dir, "server.crt"))

	err := EnsureCert(filepath.Join(dir, "server.crt"), keyPath)
	if err == nil {
		t.Fatal("expected error when key exists but cert missing")
	}
	if !strings.Contains(err.Error(), "certificate") {
		t.Errorf("error should mention certificate, got: %v", err)
	}
}

func TestEnsureCert_AcceptsRSACert(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating RSA key: %v", err)
	}
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "test-rsa"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		DNSNames:     []string{"localhost"},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, rsaKey.Public(), rsaKey)
	if err != nil {
		t.Fatalf("creating RSA cert: %v", err)
	}
	os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}), 0644)
	os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rsaKey)}), 0600)

	if err := EnsureCert(certPath, keyPath); err != nil {
		t.Fatalf("EnsureCert() should accept RSA cert, got: %v", err)
	}
}
