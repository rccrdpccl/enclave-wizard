package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type DefaultsHandler struct {
	enclaveDir string
}

func NewDefaultsHandler(enclaveDir string) *DefaultsHandler {
	return &DefaultsHandler{enclaveDir: enclaveDir}
}

type DefaultsOutput struct {
	Body struct {
		Disconnected     bool     `json:"disconnected" doc:"Default air-gapped mode"`
		MasterMaxPods    int      `json:"masterMaxPods" doc:"Default max pods per node"`
		DiskEncryption   bool     `json:"diskEncryption" doc:"Default disk encryption setting"`
		OCMirrorLogLevel string   `json:"ocMirrorLogLevel" doc:"Default oc-mirror log level"`
		StoragePlugin    string   `json:"storagePlugin" doc:"Default storage plugin"`
		EnabledPlugins   []string `json:"enabledPlugins" doc:"Default enabled plugins"`
	}
}

func (h *DefaultsHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-defaults",
		Method:      http.MethodGet,
		Path:        "/api/v1/defaults",
		Summary:     "Get default configuration values",
		Description: "Returns default values from defaults/deployment.yaml and plugin descriptors.",
		Tags:        []string{"Defaults"},
	}, h.getDefaults)
}

func (h *DefaultsHandler) getDefaults(_ context.Context, _ *struct{}) (*DefaultsOutput, error) {
	// TODO: load from defaults/deployment.yaml
	out := &DefaultsOutput{}
	out.Body.Disconnected = true
	out.Body.MasterMaxPods = 500
	out.Body.DiskEncryption = false
	out.Body.OCMirrorLogLevel = "info"
	out.Body.StoragePlugin = "lvms"
	out.Body.EnabledPlugins = []string{"lvms"}
	return out, nil
}
