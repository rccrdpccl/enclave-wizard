package defaults

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Defaults struct {
	Disconnected     bool              `json:"disconnected"`
	MasterMaxPods    int               `json:"masterMaxPods"`
	DiskEncryption   bool              `json:"diskEncryption"`
	OCMirrorLogLevel string            `json:"ocMirrorLogLevel"`
	StoragePlugin    string            `json:"storagePlugin"`
	PluginDefaults   map[string]any    `json:"pluginDefaults,omitempty"`
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
	Name     string         `yaml:"name"`
	Defaults map[string]any `yaml:"defaults"`
}

func (r *Reader) ReadAll() (*Defaults, error) {
	d := &Defaults{}

	if err := r.readDeploymentDefaults(d); err != nil {
		return nil, fmt.Errorf("reading deployment defaults: %w", err)
	}

	r.readPluginDefaults(d)

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

func (r *Reader) readPluginDefaults(d *Defaults) {
	pluginsDir := filepath.Join(r.enclaveDir, "plugins")
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return
	}

	merged := make(map[string]any)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		path := filepath.Join(pluginsDir, entry.Name(), "plugin.yaml")
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var pf pluginFile
		if yaml.Unmarshal(data, &pf) != nil {
			continue
		}

		for k, v := range pf.Defaults {
			merged[k] = v
		}
	}

	if len(merged) > 0 {
		d.PluginDefaults = merged
	}
}
