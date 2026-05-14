package defaults

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/plugins"
	"gopkg.in/yaml.v3"
)

type Defaults struct {
	Disconnected     bool   `json:"disconnected"`
	MasterMaxPods    int    `json:"masterMaxPods"`
	DiskEncryption   bool   `json:"diskEncryption"`
	OCMirrorLogLevel string `json:"ocMirrorLogLevel"`
	StoragePlugin    string `json:"storagePlugin"`

	LVMSDefaults *models.LVMSConfig `json:"lvmsDefaults,omitempty"`
	ODFDefaults  *models.ODFConfig  `json:"odfDefaults,omitempty"`
}

type Reader struct {
	enclaveDir string
}

func NewReader(enclaveDir string) *Reader {
	return &Reader{enclaveDir: enclaveDir}
}

type deploymentDefaults struct {
	Disconnected     bool   `yaml:"disconnected"`
	MasterMaxPods    int    `yaml:"masterMaxPods"`
	DiskEncryption   bool   `yaml:"diskEncryption"`
	OCMirrorLogLevel string `yaml:"ocMirrorLogLevel"`
	StoragePlugin    string `yaml:"storage_plugin"`
}

type pluginFile struct {
	Defaults map[string]any `yaml:"defaults"`
}

func (r *Reader) ReadAll() (*Defaults, error) {
	d := &Defaults{}

	if err := r.readDeploymentDefaults(d); err != nil {
		return nil, fmt.Errorf("reading deployment defaults: %w", err)
	}

	if err := r.readPluginDefaults(d); err != nil {
		return nil, fmt.Errorf("reading plugin defaults: %w", err)
	}

	return d, nil
}

func (r *Reader) readDeploymentDefaults(d *Defaults) error {
	path := filepath.Join(r.enclaveDir, "defaults", "deployment.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var dd deploymentDefaults
	if err := yaml.Unmarshal(data, &dd); err != nil {
		return fmt.Errorf("parsing deployment.yaml: %w", err)
	}

	d.Disconnected = dd.Disconnected
	d.MasterMaxPods = dd.MasterMaxPods
	d.DiskEncryption = dd.DiskEncryption
	d.OCMirrorLogLevel = dd.OCMirrorLogLevel
	d.StoragePlugin = dd.StoragePlugin
	return nil
}

func (r *Reader) readPluginDefaults(d *Defaults) error {
	for _, p := range plugins.All {
		path := filepath.Join(r.enclaveDir, "plugins", p.Name, "plugin.yaml")
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("reading %s plugin.yaml: %w", p.Name, err)
		}

		var pf pluginFile
		if err := yaml.Unmarshal(data, &pf); err != nil {
			return fmt.Errorf("parsing %s plugin.yaml: %w", p.Name, err)
		}

		if len(pf.Defaults) == 0 {
			continue
		}

		switch p.Name {
		case "lvms":
			if err := extractDefault(pf.Defaults, "lvmsDefaults", &d.LVMSDefaults); err != nil {
				return fmt.Errorf("lvms defaults: %w", err)
			}
		case "odf":
			if err := extractDefault(pf.Defaults, "odfDefaults", &d.ODFDefaults); err != nil {
				return fmt.Errorf("odf defaults: %w", err)
			}
		}
	}
	return nil
}

func extractDefault[T any](defaults map[string]any, key string, target **T) error {
	raw, ok := defaults[key]
	if !ok {
		return nil
	}
	b, err := yaml.Marshal(raw)
	if err != nil {
		return fmt.Errorf("marshaling %s: %w", key, err)
	}
	var v T
	if err := yaml.Unmarshal(b, &v); err != nil {
		return fmt.Errorf("unmarshaling %s: %w", key, err)
	}
	*target = &v
	return nil
}
