package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"gopkg.in/yaml.v3"
)

func setupTestServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	enclaveDir := t.TempDir()
	mux := http.NewServeMux()
	SetupAPI(mux, enclaveDir)
	return httptest.NewServer(mux), enclaveDir
}

func testConfig() models.EnclaveConfig {
	disc := true
	return models.EnclaveConfig{
		Global: models.GlobalConfig{
			WorkingDir:          "/home/enclave",
			BaseDomain:          "enclave-test.nodns.in",
			ClusterName:         "mgmt",
			MachineNetwork:      "192.168.2.0/24",
			APIVIP:              "192.168.2.10",
			IngressVIP:          "192.168.2.11",
			RendezvousIP:        "192.168.2.100",
			DefaultDNS:          "192.168.2.1",
			DefaultGateway:      "192.168.2.1",
			DefaultPrefix:       24,
			LZBMCIP:             "192.168.2.50",
			QuayUser:            "admin",
			QuayPassword:        "secret",
			QuayBackend:         "LocalStorage",
			BlockStorageBackend: "lvms",
			PullSecret:          map[string]any{"auths": map[string]any{}},
			SSHPubPath:          "/home/enclave/.ssh/id_rsa.pub",
			Disconnected:        &disc,
			AgentHosts: []models.HostEntry{
				{
					Name:            "cp-0",
					MACAddress:      "aa:bb:cc:dd:ee:01",
					IPAddress:       "192.168.2.100",
					Redfish:         "192.168.2.200",
					RedfishUser:     "admin",
					RedfishPassword: "redfish-pass",
					RootDisk:        "/dev/disk/by-path/pci-0000:00:11.4-ata-1.0",
				},
				{
					Name:            "cp-1",
					MACAddress:      "aa:bb:cc:dd:ee:02",
					IPAddress:       "192.168.2.101",
					Redfish:         "192.168.2.201",
					RedfishUser:     "admin",
					RedfishPassword: "redfish-pass",
					RootDisk:        "/dev/disk/by-path/pci-0000:00:11.4-ata-1.0",
				},
				{
					Name:            "cp-2",
					MACAddress:      "aa:bb:cc:dd:ee:03",
					IPAddress:       "192.168.2.102",
					Redfish:         "192.168.2.202",
					RedfishUser:     "admin",
					RedfishPassword: "redfish-pass",
					RootDisk:        "/dev/disk/by-path/pci-0000:00:11.4-ata-1.0",
				},
			},
		},
		Certificates: models.CertificatesConfig{
			SSLCACertificate: strPtr("-----BEGIN CERTIFICATE-----\nTEST\n-----END CERTIFICATE-----"),
		},
		CloudInfra: models.CloudInfraConfig{
			DiscoveryHosts: []models.HostEntry{
				{
					Name:            "worker-0",
					MACAddress:      "aa:bb:cc:dd:ee:10",
					IPAddress:       "192.168.2.110",
					Redfish:         "192.168.2.210",
					RedfishUser:     "admin",
					RedfishPassword: "redfish-pass",
					RootDisk:        "/dev/disk/by-path/pci-0000:00:11.4-ata-1.0",
				},
			},
		},
	}
}

func TestWriteConfig(t *testing.T) {
	srv, enclaveDir := setupTestServer(t)
	defer srv.Close()

	cfg := testConfig()
	body, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshaling request body: %v", err)
	}

	// PUT the config
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/api/v1/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT /api/v1/config: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}

	// Verify global.yaml
	t.Run("global.yaml", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join(enclaveDir, "config", "global.yaml"))
		if err != nil {
			t.Fatalf("reading global.yaml: %v", err)
		}
		var got models.GlobalConfig
		if err := yaml.Unmarshal(data, &got); err != nil {
			t.Fatalf("parsing global.yaml: %v", err)
		}

		assertEqual(t, "clusterName", cfg.Global.ClusterName, got.ClusterName)
		assertEqual(t, "baseDomain", cfg.Global.BaseDomain, got.BaseDomain)
		assertEqual(t, "machineNetwork", cfg.Global.MachineNetwork, got.MachineNetwork)
		assertEqual(t, "apiVIP", cfg.Global.APIVIP, got.APIVIP)
		assertEqual(t, "ingressVIP", cfg.Global.IngressVIP, got.IngressVIP)
		assertEqual(t, "rendezvousIP", cfg.Global.RendezvousIP, got.RendezvousIP)
		assertEqual(t, "defaultDNS", cfg.Global.DefaultDNS, got.DefaultDNS)
		assertEqual(t, "defaultGateway", cfg.Global.DefaultGateway, got.DefaultGateway)
		assertEqual(t, "defaultPrefix", cfg.Global.DefaultPrefix, got.DefaultPrefix)
		assertEqual(t, "quayBackend", cfg.Global.QuayBackend, got.QuayBackend)
		assertEqual(t, "blockStorageBackend", cfg.Global.BlockStorageBackend, got.BlockStorageBackend)
		assertEqual(t, "sshPubPath", cfg.Global.SSHPubPath, got.SSHPubPath)
		assertEqual(t, "agent_hosts count", len(cfg.Global.AgentHosts), len(got.AgentHosts))

		for i, want := range cfg.Global.AgentHosts {
			assertEqual(t, "agent_hosts[%d].name", want.Name, got.AgentHosts[i].Name)
			assertEqual(t, "agent_hosts[%d].ipAddress", want.IPAddress, got.AgentHosts[i].IPAddress)
			assertEqual(t, "agent_hosts[%d].macAddress", want.MACAddress, got.AgentHosts[i].MACAddress)
		}
	})

	// Verify certificates.yaml
	t.Run("certificates.yaml", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join(enclaveDir, "config", "certificates.yaml"))
		if err != nil {
			t.Fatalf("reading certificates.yaml: %v", err)
		}
		var got models.CertificatesConfig
		if err := yaml.Unmarshal(data, &got); err != nil {
			t.Fatalf("parsing certificates.yaml: %v", err)
		}

		if got.SSLCACertificate == nil {
			t.Fatal("sslCACertificate is nil")
		}
		assertEqual(t, "sslCACertificate", *cfg.Certificates.SSLCACertificate, *got.SSLCACertificate)
	})

	// Verify cloud_infra.yaml
	t.Run("cloud_infra.yaml", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join(enclaveDir, "config", "cloud_infra.yaml"))
		if err != nil {
			t.Fatalf("reading cloud_infra.yaml: %v", err)
		}
		var got models.CloudInfraConfig
		if err := yaml.Unmarshal(data, &got); err != nil {
			t.Fatalf("parsing cloud_infra.yaml: %v", err)
		}

		assertEqual(t, "discovery_hosts count", len(cfg.CloudInfra.DiscoveryHosts), len(got.DiscoveryHosts))
		assertEqual(t, "discovery_hosts[0].name", cfg.CloudInfra.DiscoveryHosts[0].Name, got.DiscoveryHosts[0].Name)
		assertEqual(t, "discovery_hosts[0].ipAddress", cfg.CloudInfra.DiscoveryHosts[0].IPAddress, got.DiscoveryHosts[0].IPAddress)
	})
}

func TestWriteConfigRoundTrip(t *testing.T) {
	srv, _ := setupTestServer(t)
	defer srv.Close()

	cfg := testConfig()
	body, _ := json.Marshal(cfg)

	// Write
	putReq, _ := http.NewRequest(http.MethodPut, srv.URL+"/api/v1/config", bytes.NewReader(body))
	putReq.Header.Set("Content-Type", "application/json")
	putResp, err := http.DefaultClient.Do(putReq)
	if err != nil {
		t.Fatalf("PUT: %v", err)
	}
	putResp.Body.Close()

	// Read back
	getResp, err := http.Get(srv.URL + "/api/v1/config")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer getResp.Body.Close()

	var got models.EnclaveConfig
	if err := json.NewDecoder(getResp.Body).Decode(&got); err != nil {
		t.Fatalf("decoding GET response: %v", err)
	}

	assertEqual(t, "clusterName", cfg.Global.ClusterName, got.Global.ClusterName)
	assertEqual(t, "baseDomain", cfg.Global.BaseDomain, got.Global.BaseDomain)
	assertEqual(t, "apiVIP", cfg.Global.APIVIP, got.Global.APIVIP)
	assertEqual(t, "ingressVIP", cfg.Global.IngressVIP, got.Global.IngressVIP)
	assertEqual(t, "agent_hosts count", len(cfg.Global.AgentHosts), len(got.Global.AgentHosts))
	assertEqual(t, "discovery_hosts count", len(cfg.CloudInfra.DiscoveryHosts), len(got.CloudInfra.DiscoveryHosts))

	if got.Certificates.SSLCACertificate == nil {
		t.Fatal("round-trip: sslCACertificate is nil")
	}
	assertEqual(t, "sslCACertificate", *cfg.Certificates.SSLCACertificate, *got.Certificates.SSLCACertificate)
}

func assertEqual[T comparable](t *testing.T, field string, want, got T) {
	t.Helper()
	if want != got {
		t.Errorf("%s: want %v, got %v", field, want, got)
	}
}

func strPtr(s string) *string { return &s }
