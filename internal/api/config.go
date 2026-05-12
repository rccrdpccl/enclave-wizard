package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/config"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/validation"
)

type ConfigHandler struct {
	reader    *config.Reader
	writer    *config.Writer
	validator *validation.Validator
}

func NewConfigHandler(reader *config.Reader, writer *config.Writer, validator *validation.Validator) *ConfigHandler {
	return &ConfigHandler{reader: reader, writer: writer, validator: validator}
}

// --- Input / Output types ---

type GetConfigOutput struct {
	Body models.EnclaveConfig
}

type WriteConfigInput struct {
	Body models.EnclaveConfig
}

type ValidateConfigInput struct {
	Body models.EnclaveConfig
}

type ValidateConfigOutput struct {
	Body struct {
		Valid  bool                     `json:"valid" doc:"Whether the config passes all validation checks"`
		Errors []models.ValidationError `json:"errors,omitempty" doc:"Validation errors, if any"`
	}
}

type PreviewConfigInput struct {
	Body models.EnclaveConfig
}

type PreviewConfigOutput struct {
	Body struct {
		GlobalYaml       string `json:"globalYaml" doc:"Rendered global.yaml content"`
		CertificatesYaml string `json:"certificatesYaml" doc:"Rendered certificates.yaml content"`
		CloudInfraYaml   string `json:"cloudInfraYaml" doc:"Rendered cloud_infra.yaml content"`
	}
}

// --- Registration ---

func (h *ConfigHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-config",
		Method:      http.MethodGet,
		Path:        "/api/v1/config",
		Summary:     "Load existing configuration",
		Description: "Reads config/global.yaml, config/certificates.yaml, and config/cloud_infra.yaml from the Enclave directory and returns the merged configuration.",
		Tags:        []string{"Config"},
	}, h.getConfig)

	huma.Register(api, huma.Operation{
		OperationID: "write-config",
		Method:      http.MethodPut,
		Path:        "/api/v1/config",
		Summary:     "Write configuration to disk",
		Description: "Accepts wizard state, serializes to YAML, and writes to the Enclave config directory.",
		Tags:        []string{"Config"},
	}, h.writeConfig)

	huma.Register(api, huma.Operation{
		OperationID:   "validate-config",
		Method:        http.MethodPost,
		Path:          "/api/v1/config/validate",
		Summary:       "Validate configuration",
		Description:   "Validates the candidate configuration against Enclave JSON schemas and returns structured errors.",
		Tags:          []string{"Config"},
	}, h.validateConfig)

	huma.Register(api, huma.Operation{
		OperationID: "preview-config",
		Method:      http.MethodPost,
		Path:        "/api/v1/config/preview",
		Summary:     "Preview rendered YAML",
		Description: "Returns the rendered YAML content for each config file without writing to disk.",
		Tags:        []string{"Config"},
	}, h.previewConfig)
}

// --- Handlers ---

func (h *ConfigHandler) getConfig(_ context.Context, _ *struct{}) (*GetConfigOutput, error) {
	cfg, err := h.reader.ReadAll()
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to read config", err)
	}
	return &GetConfigOutput{Body: *cfg}, nil
}

func (h *ConfigHandler) writeConfig(_ context.Context, input *WriteConfigInput) (*struct{}, error) {
	if errs := h.validator.Validate(&input.Body); len(errs) > 0 {
		return nil, huma.Error422UnprocessableEntity("config validation failed")
	}
	if err := h.writer.WriteAll(&input.Body); err != nil {
		return nil, huma.Error500InternalServerError("failed to write config", err)
	}
	return nil, nil
}

func (h *ConfigHandler) validateConfig(_ context.Context, input *ValidateConfigInput) (*ValidateConfigOutput, error) {
	errs := h.validator.Validate(&input.Body)
	out := &ValidateConfigOutput{}
	out.Body.Valid = len(errs) == 0
	out.Body.Errors = errs
	return out, nil
}

func (h *ConfigHandler) previewConfig(_ context.Context, _ *PreviewConfigInput) (*PreviewConfigOutput, error) {
	// TODO: serialize each config section to YAML and return as strings
	out := &PreviewConfigOutput{}
	out.Body.GlobalYaml = "# TODO: render global.yaml"
	out.Body.CertificatesYaml = "# TODO: render certificates.yaml"
	out.Body.CloudInfraYaml = "# TODO: render cloud_infra.yaml"
	return out, nil
}
