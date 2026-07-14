package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"gopkg.in/yaml.v3"
)

// LoadResult holds plugins and their schemas loaded from the enclave directory.
type LoadResult struct {
	Plugins []models.Plugin
	Schemas map[string][]byte // plugin name -> JSON schema bytes
}

func LoadFromDir(enclaveDir string) ([]models.Plugin, error) {
	result, err := LoadFromDirWithSchemas(enclaveDir)
	if err != nil {
		return nil, err
	}
	return result.Plugins, nil
}

// LoadFromDirWithSchemas loads plugins and their config schemas from the enclave
// directory. Schemas are looked for at plugins/{name}/schemas/config.yaml and
// plugins/{name}/schemas/config.json. YAML schemas are converted to JSON.
func LoadFromDirWithSchemas(enclaveDir string) (*LoadResult, error) {
	pluginsDir := filepath.Join(enclaveDir, "plugins")
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil, fmt.Errorf("reading plugins directory: %w", err)
	}

	result := &LoadResult{
		Schemas: make(map[string][]byte),
	}

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

		result.Plugins = append(result.Plugins, p)

		// Try loading the schema
		schema := loadSchema(filepath.Join(pluginsDir, entry.Name(), "schemas"))
		if schema != nil {
			result.Schemas[p.Name] = schema
		}
	}

	sort.Slice(result.Plugins, func(i, j int) bool {
		if result.Plugins[i].Order != result.Plugins[j].Order {
			return result.Plugins[i].Order < result.Plugins[j].Order
		}
		return result.Plugins[i].Name < result.Plugins[j].Name
	})

	return result, nil
}

// loadSchema attempts to read a config schema from the schemas directory.
// It tries config.json first, then config.yaml (converting to JSON).
// Returns nil if no schema is found.
func loadSchema(schemasDir string) []byte {
	// Try JSON first
	jsonPath := filepath.Join(schemasDir, "config.json")
	if data, err := os.ReadFile(jsonPath); err == nil {
		// Validate it's actually JSON
		if json.Valid(data) {
			return data
		}
	}

	// Try YAML, convert to JSON
	yamlPath := filepath.Join(schemasDir, "config.yaml")
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil
	}

	var parsed any
	if yaml.Unmarshal(data, &parsed) != nil {
		return nil
	}

	jsonData, err := json.Marshal(parsed)
	if err != nil {
		return nil
	}
	return jsonData
}
