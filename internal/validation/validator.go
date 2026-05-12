package validation

import (
	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
)

type Validator struct {
	enclaveDir string
}

func NewValidator(enclaveDir string) *Validator {
	return &Validator{enclaveDir: enclaveDir}
}

// Validate checks the config against the Enclave JSON schemas
// (schemas/variables.yaml, schemas/definitions.yaml).
// TODO: load YAML schemas, resolve $ref, validate with conditional allOf/if/then blocks
func (v *Validator) Validate(cfg *models.EnclaveConfig) []models.ValidationError {
	return nil
}
