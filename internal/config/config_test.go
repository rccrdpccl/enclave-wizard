package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"gopkg.in/yaml.v3"
)

// newEnclaveDir creates the expected directory layout under a temp dir and
// returns the enclave root. Optionally seeds config files from the provided map
// (filename → YAML content).
func newEnclaveDir(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(configDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return root
}

//--- Reader.ConfigExists ---

func TestConfigExists_ReturnsTrueWhenFilePresent(t *testing.T) {
	root := newEnclaveDir(t, map[string]string{
		"global.yaml": "baseDomain: test.local\n",
	})
	r := NewReader(root)
	if !r.ConfigExists() {
		t.Error("expected ConfigExists=true when global.yaml is present")
	}
}

func TestConfigExists_ReturnsFalseWhenFileMissing(t *testing.T) {
	root := newEnclaveDir(t, nil)
	r := NewReader(root)
	if r.ConfigExists() {
		t.Error("expected ConfigExists=false when global.yaml is absent")
	}
}

func TestConfigExists_ReturnsFalseForNonexistentDir(t *testing.T) {
	r := NewReader("/nonexistent/path/xyz")
	if r.ConfigExists() {
		t.Error("expected ConfigExists=false for nonexistent enclave dir")
	}
}

// --- Reader.ReadAll happy path ---

func TestReadAll_AllFilesMissing_ReturnsZeroValues(t *testing.T) {
	root := newEnclaveDir(t, nil)
	r := NewReader(root)

	cfg, err := r.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Global.BaseDomain != "" {
		t.Errorf("expected empty BaseDomain, got %q", cfg.Global.BaseDomain)
	}
	if len(cfg.CloudInfra.DiscoveryHosts) != 0 {
		t.Errorf("expected no discovery hosts, got %v", cfg.CloudInfra.DiscoveryHosts)
	}
}

func TestReadAll_ReadsGlobalFields(t *testing.T) {
	root := newEnclaveDir(t, map[string]string{
		"global.yaml": "baseDomain: example.com\nclusterName: mgmt\n",
	})
	cfg, err := NewReader(root).ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if cfg.Global.BaseDomain != "example.com" {
		t.Errorf("BaseDomain: got %q", cfg.Global.BaseDomain)
	}
	if cfg.Global.ClusterName != "mgmt" {
		t.Errorf("ClusterName: got %q", cfg.Global.ClusterName)
	}
}

func TestReadAll_ReadsCertificates(t *testing.T) {
	root := newEnclaveDir(t, map[string]string{
		"certificates.yaml": "sslCACertificate: |\n  -----BEGIN CERTIFICATE-----\n  TEST\n  -----END CERTIFICATE-----\n",
	})
	cfg, err := NewReader(root).ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if cfg.Certificates.SSLCACertificate == nil {
		t.Fatal("expected SSLCACertificate to be set")
	}
}

func TestReadAll_ReadsCloudInfraDiscoveryHosts(t *testing.T) {
	root := newEnclaveDir(t, map[string]string{
		"cloud_infra.yaml": "discovery_hosts:\n  - bmcAddress: 192.168.1.10\n",
	})
	cfg, err := NewReader(root).ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(cfg.CloudInfra.DiscoveryHosts) != 1 {
		t.Fatalf("expected 1 discovery host, got %d", len(cfg.CloudInfra.DiscoveryHosts))
	}
}

// --- Reader.ReadAll: discovery_hosts fallback ---

func TestReadAll_FallsBackToGlobalDiscoveryHosts(t *testing.T) {
	// discovery_hosts in global.yaml should be merged when cloud_infra.yaml is empty.
	root := newEnclaveDir(t, map[string]string{
		"global.yaml": "baseDomain: test.local\ndiscovery_hosts:\n  - redfish: 10.0.0.1\n",
	})
	cfg, err := NewReader(root).ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(cfg.CloudInfra.DiscoveryHosts) != 1 {
		t.Fatalf("expected 1 fallback discovery host, got %d", len(cfg.CloudInfra.DiscoveryHosts))
	}
	if cfg.CloudInfra.DiscoveryHosts[0].Redfish != "10.0.0.1" {
		t.Errorf("unexpected host redfish: %v", cfg.CloudInfra.DiscoveryHosts[0])
	}
}

func TestReadAll_CloudInfraTakesPrecedenceOverGlobal(t *testing.T) {
	// When cloud_infra.yaml already has hosts, global.yaml hosts must be ignored.
	root := newEnclaveDir(t, map[string]string{
		"global.yaml":      "discovery_hosts:\n  - redfish: 10.0.0.1\n",
		"cloud_infra.yaml": "discovery_hosts:\n  - redfish: 10.0.0.2\n",
	})
	cfg, err := NewReader(root).ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(cfg.CloudInfra.DiscoveryHosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(cfg.CloudInfra.DiscoveryHosts))
	}
	if cfg.CloudInfra.DiscoveryHosts[0].Redfish != "10.0.0.2" {
		t.Errorf("expected cloud_infra host to win, got %v", cfg.CloudInfra.DiscoveryHosts[0])
	}
}

// --- Reader.ReadAll: malformed YAML ---

func TestReadAll_MalformedGlobalYAML_ReturnsError(t *testing.T) {
	root := newEnclaveDir(t, map[string]string{
		"global.yaml": ":\tinvalid: yaml: [\n",
	})
	_, err := NewReader(root).ReadAll()
	if err == nil {
		t.Fatal("expected error for malformed global.yaml, got nil")
	}
}

func TestReadAll_MalformedCertificatesYAML_ReturnsError(t *testing.T) {
	root := newEnclaveDir(t, map[string]string{
		"certificates.yaml": ":\tinvalid: yaml: [\n",
	})
	_, err := NewReader(root).ReadAll()
	if err == nil {
		t.Fatal("expected error for malformed certificates.yaml, got nil")
	}
}

func TestReadAll_MalformedCloudInfraYAML_ReturnsError(t *testing.T) {
	root := newEnclaveDir(t, map[string]string{
		"cloud_infra.yaml": ":\tinvalid: yaml: [\n",
	})
	_, err := NewReader(root).ReadAll()
	if err == nil {
		t.Fatal("expected error for malformed cloud_infra.yaml, got nil")
	}
}

// --- Writer.WriteAll ---

func TestWriteAll_CreatesConfigDir(t *testing.T) {
	root := t.TempDir() // no config/ subdir yet
	w := NewWriter(root)

	if err := w.WriteAll(&models.EnclaveConfig{}); err != nil {
		t.Fatalf("WriteAll: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "config")); err != nil {
		t.Errorf("config dir not created: %v", err)
	}
}

func TestWriteAll_WritesAllThreeFiles(t *testing.T) {
	root := t.TempDir()
	if err := NewWriter(root).WriteAll(&models.EnclaveConfig{}); err != nil {
		t.Fatalf("WriteAll: %v", err)
	}
	for _, name := range []string{"global.yaml", "certificates.yaml", "cloud_infra.yaml"} {
		if _, err := os.Stat(filepath.Join(root, "config", name)); err != nil {
			t.Errorf("expected %s to exist: %v", name, err)
		}
	}
}

func TestWriteAll_OutputIsValidYAML(t *testing.T) {
	root := t.TempDir()
	cfg := &models.EnclaveConfig{}
	cfg.Global.BaseDomain = "write.test"
	cfg.Global.ClusterName = "cluster1"

	if err := NewWriter(root).WriteAll(cfg); err != nil {
		t.Fatalf("WriteAll: %v", err)
	}

	for _, name := range []string{"global.yaml", "certificates.yaml", "cloud_infra.yaml"} {
		data, err := os.ReadFile(filepath.Join(root, "config", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		var m map[string]any
		if err := yaml.Unmarshal(data, &m); err != nil {
			t.Errorf("%s is not valid YAML: %v", name, err)
		}
	}
}

// --- Round-trip: WriteAll → ReadAll ---

func TestWriteAllThenReadAll_GlobalRoundTrips(t *testing.T) {
	root := t.TempDir()

	want := &models.EnclaveConfig{}
	want.Global.BaseDomain = "roundtrip.test"
	want.Global.ClusterName = "mgmt"
	want.Global.APIVIP = "192.168.1.100"

	if err := NewWriter(root).WriteAll(want); err != nil {
		t.Fatalf("WriteAll: %v", err)
	}

	got, err := NewReader(root).ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	if got.Global.BaseDomain != want.Global.BaseDomain {
		t.Errorf("BaseDomain: want %q, got %q", want.Global.BaseDomain, got.Global.BaseDomain)
	}
	if got.Global.ClusterName != want.Global.ClusterName {
		t.Errorf("ClusterName: want %q, got %q", want.Global.ClusterName, got.Global.ClusterName)
	}
	if got.Global.APIVIP != want.Global.APIVIP {
		t.Errorf("APIVIP: want %q, got %q", want.Global.APIVIP, got.Global.APIVIP)
	}
}

func TestWriteAllThenReadAll_CertificatesRoundTrip(t *testing.T) {
	root := t.TempDir()
	cert := "-----BEGIN CERTIFICATE-----\nABCD\n-----END CERTIFICATE-----\n"

	want := &models.EnclaveConfig{}
	want.Certificates.SSLCACertificate = &cert

	if err := NewWriter(root).WriteAll(want); err != nil {
		t.Fatalf("WriteAll: %v", err)
	}

	got, err := NewReader(root).ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	if got.Certificates.SSLCACertificate == nil {
		t.Fatal("SSLCACertificate is nil after round-trip")
	}
}

func TestWriteAllThenReadAll_DiscoveryHostsRoundTrip(t *testing.T) {
	root := t.TempDir()

	want := &models.EnclaveConfig{}
	want.CloudInfra.DiscoveryHosts = []models.HostEntry{
		{Redfish: "192.168.2.10"},
		{Redfish: "192.168.2.11"},
	}

	if err := NewWriter(root).WriteAll(want); err != nil {
		t.Fatalf("WriteAll: %v", err)
	}

	got, err := NewReader(root).ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	if len(got.CloudInfra.DiscoveryHosts) != 2 {
		t.Fatalf("expected 2 discovery hosts, got %d", len(got.CloudInfra.DiscoveryHosts))
	}
	if got.CloudInfra.DiscoveryHosts[0].Redfish != "192.168.2.10" {
		t.Errorf("host[0]: got %v", got.CloudInfra.DiscoveryHosts[0])
	}
}

func TestWriteAll_FilePermsAre0640(t *testing.T) {
	root := t.TempDir()
	if err := NewWriter(root).WriteAll(&models.EnclaveConfig{}); err != nil {
		t.Fatalf("WriteAll: %v", err)
	}
	for _, name := range []string{"global.yaml", "certificates.yaml", "cloud_infra.yaml"} {
		info, err := os.Stat(filepath.Join(root, "config", name))
		if err != nil {
			t.Fatalf("stat %s: %v", name, err)
		}
		if perm := info.Mode().Perm(); perm != 0640 {
			t.Errorf("%s: expected 0640, got %o", name, perm)
		}
	}
}

