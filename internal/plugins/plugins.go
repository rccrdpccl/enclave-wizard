package plugins

import "github.com/rh-ecosystem-edge/enclave-wizard/internal/models"

var All = []models.Plugin{
	{Name: "lvms", Type: models.PluginTypeFoundation, Description: "Logical Volume Manager Storage"},
	{Name: "odf", Type: models.PluginTypeFoundation, Description: "OpenShift Data Foundation"},
	{Name: "nvidia-gpu", Type: models.PluginTypeAddon, Description: "NVIDIA GPU Operator"},
	{Name: "openshift-ai", Type: models.PluginTypeAddon, Description: "Red Hat OpenShift AI"},
}

var byName map[string]models.Plugin

func init() {
	byName = make(map[string]models.Plugin, len(All))
	for _, p := range All {
		byName[p.Name] = p
	}
}

func Get(name string) (models.Plugin, bool) {
	p, ok := byName[name]
	return p, ok
}

func ValidateCombination(names []string) []models.ValidationError {
	var errs []models.ValidationError
	for _, name := range names {
		if _, ok := byName[name]; !ok {
			errs = append(errs, models.ValidationError{
				Field:   "plugins",
				Message: "unknown plugin: " + name,
			})
		}
	}
	return errs
}
