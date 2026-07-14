package plugins

import (
	"errors"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

// ErrSchemaNotFound is returned when a plugin has no config schema.
var ErrSchemaNotFound = errors.New("schema not found for plugin")

type Registry struct {
	all      []models.Plugin
	byName   map[string]models.Plugin
	schemas  map[string][]byte // plugin name -> JSON schema bytes
}

func NewRegistry(plugins []models.Plugin) *Registry {
	byName := make(map[string]models.Plugin, len(plugins))
	for _, p := range plugins {
		byName[p.Name] = p
	}
	return &Registry{all: plugins, byName: byName, schemas: make(map[string][]byte)}
}

func (r *Registry) All() []models.Plugin {
	return r.all
}

func (r *Registry) Get(name string) (models.Plugin, bool) {
	p, ok := r.byName[name]
	return p, ok
}

// SetSchema stores a JSON schema for a plugin.
func (r *Registry) SetSchema(name string, schema []byte) {
	r.schemas[name] = schema
}

// GetSchema returns the JSON schema for the named plugin.
// Returns ErrSchemaNotFound if the plugin has no schema.
func (r *Registry) GetSchema(name string) ([]byte, error) {
	schema, ok := r.schemas[name]
	if !ok {
		return nil, ErrSchemaNotFound
	}
	return schema, nil
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
