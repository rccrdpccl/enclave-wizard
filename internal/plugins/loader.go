package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"gopkg.in/yaml.v3"
)

type Loader struct {
	enclaveDir string
}

func NewLoader(enclaveDir string) *Loader {
	return &Loader{enclaveDir: enclaveDir}
}

func (l *Loader) pluginsDir() string {
	return filepath.Join(l.enclaveDir, "plugins")
}

// LoadAll discovers plugin descriptors by walking plugins/*/plugin.yaml.
// Results are sorted by type (foundation before addon), then by order.
func (l *Loader) LoadAll() ([]models.PluginDescriptor, error) {
	entries, err := os.ReadDir(l.pluginsDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading plugins directory: %w", err)
	}

	var plugins []models.PluginDescriptor
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(l.pluginsDir(), entry.Name(), "plugin.yaml")
		desc, err := loadPluginFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("loading %s: %w", path, err)
		}
		plugins = append(plugins, *desc)
	}

	sort.Slice(plugins, func(i, j int) bool {
		ti, tj := typeRank(plugins[i].Type), typeRank(plugins[j].Type)
		if ti != tj {
			return ti < tj
		}
		return orderVal(plugins[i].Order) < orderVal(plugins[j].Order)
	})

	return plugins, nil
}

func loadPluginFile(path string) (*models.PluginDescriptor, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var desc models.PluginDescriptor
	if err := yaml.Unmarshal(data, &desc); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filepath.Base(path), err)
	}
	return &desc, nil
}

func typeRank(t string) int {
	if t == "foundation" {
		return 0
	}
	return 1
}

func orderVal(o *int) int {
	if o == nil {
		return 0
	}
	return *o
}

// ValidateCombination checks whether the given plugin names form a valid combination.
// TODO: encode dependency rules (e.g. openshift-ai requires nvidia-gpu)
func (l *Loader) ValidateCombination(pluginNames []string) []models.ValidationError {
	return nil
}
