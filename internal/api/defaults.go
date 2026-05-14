package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/defaults"
)

type DefaultsHandler struct {
	reader *defaults.Reader
}

func NewDefaultsHandler(enclaveDir string) *DefaultsHandler {
	return &DefaultsHandler{reader: defaults.NewReader(enclaveDir)}
}

type DefaultsOutput struct {
	Body defaults.Defaults
}

func (h *DefaultsHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-defaults",
		Method:      http.MethodGet,
		Path:        "/api/v1/defaults",
		Summary:     "Get default configuration values",
		Description: "Returns default values read from defaults/deployment.yaml and plugin definitions.",
		Tags:        []string{"Defaults"},
	}, h.getDefaults)
}

func (h *DefaultsHandler) getDefaults(_ context.Context, _ *struct{}) (*DefaultsOutput, error) {
	d, err := h.reader.ReadAll()
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to read defaults", err)
	}
	return &DefaultsOutput{Body: *d}, nil
}
