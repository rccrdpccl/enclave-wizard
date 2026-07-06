package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type VersionHandler struct {
	wizardVersion  string
	enclaveVersion string
}

func NewVersionHandler(wizardVersion, enclaveVersion string) *VersionHandler {
	return &VersionHandler{wizardVersion: wizardVersion, enclaveVersion: enclaveVersion}
}

type VersionOutput struct {
	Body struct {
		WizardVersion  string `json:"wizardVersion" doc:"Enclave Wizard version or git hash"`
		EnclaveVersion string `json:"enclaveVersion" doc:"Enclave version or git hash"`
	}
}

func (h *VersionHandler) version(_ context.Context, _ *struct{}) (*VersionOutput, error) {
	out := &VersionOutput{}
	out.Body.WizardVersion = h.wizardVersion
	out.Body.EnclaveVersion = h.enclaveVersion
	return out, nil
}

func (h *VersionHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "getVersion",
		Method:      http.MethodGet,
		Path:        "/api/v1/version",
		Summary:     "Get wizard and enclave versions",
		Tags:        []string{"version"},
	}, h.version)
}
