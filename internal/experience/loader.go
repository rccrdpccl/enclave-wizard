package experience

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// ExperiencePlugin defines a plugin required by an experience.
type ExperiencePlugin struct {
	Name  string `json:"name" yaml:"name"`
	Order int    `json:"order,omitempty" yaml:"order,omitempty"`
}

// Experience represents a deployable experience loaded from the enclave directory.
type Experience struct {
	ID          string             `json:"id" yaml:"id"`
	Name        string             `json:"name" yaml:"name"`
	Description string             `json:"description" yaml:"description"`
	Plugins     []ExperiencePlugin `json:"plugins" yaml:"plugins"`
}

// LoadFromDir scans enclaveDir/experiences/*/experience.yaml and returns all
// valid experience definitions. If the experiences directory does not exist,
// an empty slice is returned (not an error).
func LoadFromDir(enclaveDir string) ([]Experience, error) {
	dir := filepath.Join(enclaveDir, "experiences")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading experiences directory: %w", err)
	}

	var experiences []Experience
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		yamlPath := filepath.Join(dir, entry.Name(), "experience.yaml")
		data, err := os.ReadFile(yamlPath)
		if err != nil {
			slog.Debug("skipping experience directory", "dir", entry.Name(), "error", err)
			continue
		}

		var exp Experience
		if err := yaml.Unmarshal(data, &exp); err != nil {
			slog.Warn("skipping malformed experience", "file", yamlPath, "error", err)
			continue
		}

		// Use the directory name as the ID if not set in the YAML.
		if exp.ID == "" {
			exp.ID = entry.Name()
		}

		experiences = append(experiences, exp)
	}

	sort.Slice(experiences, func(i, j int) bool {
		return experiences[i].ID < experiences[j].ID
	})

	return experiences, nil
}
