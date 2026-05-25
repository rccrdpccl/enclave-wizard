package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"gopkg.in/yaml.v3"
)

func LoadFromDir(enclaveDir string) ([]models.Plugin, error) {
	pluginsDir := filepath.Join(enclaveDir, "plugins")
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil, fmt.Errorf("reading plugins directory: %w", err)
	}

	var plugins []models.Plugin
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if entry.Name() == "example" {
			continue
		}

		path := filepath.Join(pluginsDir, entry.Name(), "plugin.yaml")
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var p models.Plugin
		if yaml.Unmarshal(data, &p) != nil {
			continue
		}
		if p.Name == "" {
			continue
		}

		plugins = append(plugins, p)
	}

	sort.Slice(plugins, func(i, j int) bool {
		if plugins[i].Order != plugins[j].Order {
			return plugins[i].Order < plugins[j].Order
		}
		return plugins[i].Name < plugins[j].Name
	})

	return plugins, nil
}
