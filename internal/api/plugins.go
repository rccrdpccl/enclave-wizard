package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/plugins"
)

type PluginsHandler struct {
	registry *plugins.Registry
}

func NewPluginsHandler(registry *plugins.Registry) *PluginsHandler {
	return &PluginsHandler{registry: registry}
}

type PluginsOutput struct {
	Body struct {
		Plugins []models.Plugin `json:"plugins" doc:"Available plugins"`
	}
}

type PluginValidateInput struct {
	Body struct {
		Plugins []string `json:"plugins" doc:"Plugin names to validate as a combination" minItems:"1"`
	}
}

type PluginValidateOutput struct {
	Body struct {
		Valid  bool                     `json:"valid" doc:"Whether the combination is valid"`
		Errors []models.ValidationError `json:"errors,omitempty" doc:"Validation errors, if any"`
	}
}

type GetPluginSchemaInput struct {
	Name string `path:"name" doc:"Plugin name" minLength:"1"`
}

type GetPluginSchemaOutput struct {
	Body json.RawMessage
}

func (h *PluginsHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "list-plugins",
		Method:      http.MethodGet,
		Path:        "/api/v1/plugins",
		Summary:     "List available plugins",
		Description: "Returns all known plugins and their types.",
		Tags:        []string{"Plugins"},
	}, h.listPlugins)

	huma.Register(api, huma.Operation{
		OperationID: "get-plugin-schema",
		Method:      http.MethodGet,
		Path:        "/api/v1/plugins/{name}/schema",
		Summary:     "Get plugin config schema",
		Description: "Returns the JSON schema for the named plugin's configuration. Used by the frontend for dynamic form rendering.",
		Tags:        []string{"Plugins"},
	}, h.getPluginSchema)

	huma.Register(api, huma.Operation{
		OperationID: "validate-plugin-combination",
		Method:      http.MethodPost,
		Path:        "/api/v1/plugins/validate",
		Summary:     "Validate plugin combination",
		Description: "Checks whether the given set of plugins forms a valid deployment combination.",
		Tags:        []string{"Plugins"},
	}, h.validateCombination)
}

func (h *PluginsHandler) listPlugins(_ context.Context, _ *struct{}) (*PluginsOutput, error) {
	out := &PluginsOutput{}
	out.Body.Plugins = h.registry.All()
	return out, nil
}

func (h *PluginsHandler) getPluginSchema(_ context.Context, input *GetPluginSchemaInput) (*GetPluginSchemaOutput, error) {
	if _, ok := h.registry.Get(input.Name); !ok {
		return nil, huma.Error404NotFound("unknown plugin: " + input.Name)
	}

	schema, err := h.registry.GetSchema(input.Name)
	if err != nil {
		if errors.Is(err, plugins.ErrSchemaNotFound) {
			return nil, huma.Error404NotFound("no schema found for plugin: " + input.Name)
		}
		return nil, huma.Error500InternalServerError("failed to read schema", err)
	}

	return &GetPluginSchemaOutput{Body: json.RawMessage(schema)}, nil
}

func (h *PluginsHandler) validateCombination(_ context.Context, input *PluginValidateInput) (*PluginValidateOutput, error) {
	errs := h.registry.ValidateCombination(input.Body.Plugins)
	out := &PluginValidateOutput{}
	out.Body.Valid = len(errs) == 0
	out.Body.Errors = errs
	if !out.Body.Valid {
		slog.Warn("plugin combination invalid", "plugins", input.Body.Plugins, "error_count", len(errs))
	}
	return out, nil
}
