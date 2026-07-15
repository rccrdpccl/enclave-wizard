package deploy

import (
	"sort"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/config"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/plugins"
)

// AddonPluginsFromConfig reads the configuration and returns the names of
// enabled addon plugins sorted by their declared order.
func AddonPluginsFromConfig(reader *config.Reader, registry *plugins.Registry) []string {
	if reader == nil {
		return nil
	}
	cfg, err := reader.ReadAll()
	if err != nil {
		return nil
	}
	type addonInfo struct {
		name  string
		order int
	}
	var addons []addonInfo
	for _, name := range cfg.Global.EnabledPlugins {
		p, ok := registry.Get(name)
		if !ok {
			continue
		}
		if p.Type == models.PluginTypeAddon {
			addons = append(addons, addonInfo{name: p.Name, order: p.Order})
		}
	}
	sort.Slice(addons, func(i, j int) bool { return addons[i].order < addons[j].order })
	result := make([]string, len(addons))
	for i, a := range addons {
		result[i] = a.name
	}
	return result
}
