package tlscert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

func GenerateSelfSigned(certPath, keyPath string) error {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generating key: %w", err)
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}

	dnsNames := []string{"localhost"}
	if hostname != "localhost" {
		dnsNames = append(dnsNames, hostname)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("generating serial: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: hostname},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     dnsNames,
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("creating certificate: %w", err)
	}

	certDir := filepath.Dir(certPath)
	if err := os.MkdirAll(certDir, 0700); err != nil {
		return fmt.Errorf("creating cert directory: %w", err)
	}
	keyDir := filepath.Dir(keyPath)
	if keyDir != certDir {
		if err := os.MkdirAll(keyDir, 0700); err != nil {
			return fmt.Errorf("creating key directory: %w", err)
		}
	}

	certTmp := certPath + ".tmp"
	if err := os.WriteFile(certTmp, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}), 0644); err != nil {
		return fmt.Errorf("writing certificate: %w", err)
	}
	if err := os.Rename(certTmp, certPath); err != nil {
		os.Remove(certTmp)
		return fmt.Errorf("installing certificate: %w", err)
	}

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return fmt.Errorf("marshaling key: %w", err)
	}
	keyTmp := keyPath + ".tmp"
	if err := os.WriteFile(keyTmp, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0600); err != nil {
		return fmt.Errorf("writing key: %w", err)
	}
	if err := os.Rename(keyTmp, keyPath); err != nil {
		os.Remove(keyTmp)
		return fmt.Errorf("installing key: %w", err)
	}

	return nil
}

func EnsureCert(certPath, keyPath string) error {
	certExists := fileExists(certPath)
	keyExists := fileExists(keyPath)

	if !certExists && !keyExists {
		log.Printf("No TLS certificate found, generating self-signed cert")
		return GenerateSelfSigned(certPath, keyPath)
	}

	if certExists && !keyExists {
		return fmt.Errorf("TLS key file not found but certificate exists: %s — provide both or remove the certificate to regenerate", certPath)
	}
	if !certExists && keyExists {
		return fmt.Errorf("TLS certificate not found but key exists: %s — provide both or remove the key to regenerate", keyPath)
	}

	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("reading certificate: %w", err)
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("reading key: %w", err)
	}

	if _, err := tls.X509KeyPair(certPEM, keyPEM); err != nil {
		return fmt.Errorf("TLS key does not match certificate: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("certificate file contains no PEM data: %s", certPath)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("parsing certificate: %w", err)
	}

	now := time.Now()
	if now.After(cert.NotAfter) {
		return fmt.Errorf("TLS certificate expired on %s: %s", cert.NotAfter.Format(time.DateOnly), certPath)
	}

	daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)
	if daysLeft < 30 {
		log.Printf("WARNING: TLS certificate expires in %d days: %s", daysLeft, certPath)
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
