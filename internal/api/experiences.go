package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/experience"
)

// ExperiencesHandler serves experience definitions loaded from the enclave directory.
type ExperiencesHandler struct {
	experiences []experience.Experience
}

// NewExperiencesHandler creates a handler that serves the given experiences.
func NewExperiencesHandler(experiences []experience.Experience) *ExperiencesHandler {
	return &ExperiencesHandler{experiences: experiences}
}

// ExperiencesOutput is the response body for GET /api/v1/experiences.
type ExperiencesOutput struct {
	Body struct {
		Experiences []experience.Experience `json:"experiences" doc:"Available experience definitions"`
	}
}

// Register adds the experiences endpoint to the API.
func (h *ExperiencesHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "list-experiences",
		Method:      http.MethodGet,
		Path:        "/api/v1/experiences",
		Summary:     "List available experiences",
		Description: "Returns experience definitions loaded from the enclave directory.",
		Tags:        []string{"Experiences"},
	}, h.listExperiences)
}

func (h *ExperiencesHandler) listExperiences(_ context.Context, _ *struct{}) (*ExperiencesOutput, error) {
	out := &ExperiencesOutput{}
	exps := h.experiences
	if exps == nil {
		exps = []experience.Experience{}
	}
	out.Body.Experiences = exps
	return out, nil
}
