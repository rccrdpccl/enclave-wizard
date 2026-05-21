package plugins

import "github.com/rh-ecosystem-edge/enclave-wizard/internal/models"

type Registry struct {
	all    []models.Plugin
	byName map[string]models.Plugin
}

func NewRegistry(plugins []models.Plugin) *Registry {
	byName := make(map[string]models.Plugin, len(plugins))
	for _, p := range plugins {
		byName[p.Name] = p
	}
	return &Registry{all: plugins, byName: byName}
}

func (r *Registry) All() []models.Plugin {
	return r.all
}

func (r *Registry) Get(name string) (models.Plugin, bool) {
	p, ok := r.byName[name]
	return p, ok
}

func (r *Registry) ValidateCombination(names []string) []models.ValidationError {
	var errs []models.ValidationError
	for _, name := range names {
		if _, ok := r.byName[name]; !ok {
			errs = append(errs, models.ValidationError{
				Field:   "plugins",
				Message: "unknown plugin: " + name,
			})
		}
	}
	return errs
}
