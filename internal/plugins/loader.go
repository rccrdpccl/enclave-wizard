package plugins

import (
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

type Loader struct {
	enclaveDir string
}

func NewLoader(enclaveDir string) *Loader {
	return &Loader{enclaveDir: enclaveDir}
}

// LoadAll discovers and loads plugin descriptors from plugins/*/plugin.yaml.
// TODO: walk plugins/ directory, parse each plugin.yaml into PluginDescriptor
func (l *Loader) LoadAll() ([]models.PluginDescriptor, error) {
	return []models.PluginDescriptor{
		{Name: "lvms", Type: "foundation", Order: intPtr(10), Mirror: strPtr("core")},
		{Name: "odf", Type: "foundation", Order: intPtr(10), Mirror: strPtr("core")},
		{Name: "nvidia-gpu", Type: "addon", Order: intPtr(110), Mirror: strPtr("plugin"), Catalog: strPtr("certified")},
		{Name: "openshift-ai", Type: "addon", Order: intPtr(100), Mirror: strPtr("plugin")},
	}, nil
}

// ValidateCombination checks whether the given plugin names form a valid combination.
// TODO: encode dependency rules (e.g. openshift-ai requires nvidia-gpu)
func (l *Loader) ValidateCombination(pluginNames []string) []models.ValidationError {
	return nil
}

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }
