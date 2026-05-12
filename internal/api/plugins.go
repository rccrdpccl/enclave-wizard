package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/plugins"
)

type PluginsHandler struct {
	loader *plugins.Loader
}

func NewPluginsHandler(loader *plugins.Loader) *PluginsHandler {
	return &PluginsHandler{loader: loader}
}

type PluginsOutput struct {
	Body struct {
		Plugins []models.PluginDescriptor `json:"plugins" doc:"Available plugin descriptors"`
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
		Description: "Returns all plugin descriptors discovered from plugins/*/plugin.yaml.",
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
	descriptors, err := h.loader.LoadAll()
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to load plugins", err)
	}
	out := &PluginsOutput{}
	out.Body.Plugins = descriptors
	return out, nil
}

func (h *PluginsHandler) validateCombination(_ context.Context, input *PluginValidateInput) (*PluginValidateOutput, error) {
	errs := h.loader.ValidateCombination(input.Body.Plugins)
	out := &PluginValidateOutput{}
	out.Body.Valid = len(errs) == 0
	out.Body.Errors = errs
	return out, nil
}
