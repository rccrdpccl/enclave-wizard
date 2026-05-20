package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/plugins"
)

type PluginsHandler struct{}

func NewPluginsHandler() *PluginsHandler {
	return &PluginsHandler{}
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
	out.Body.Plugins = plugins.All
	return out, nil
}

func (h *PluginsHandler) validateCombination(_ context.Context, input *PluginValidateInput) (*PluginValidateOutput, error) {
	errs := plugins.ValidateCombination(input.Body.Plugins)
	out := &PluginValidateOutput{}
	out.Body.Valid = len(errs) == 0
	out.Body.Errors = errs
	if !out.Body.Valid {
		slog.Warn("plugin combination invalid", "plugins", input.Body.Plugins, "error_count", len(errs))
	}
	return out, nil
}
