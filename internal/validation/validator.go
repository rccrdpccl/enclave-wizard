package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/rh-ecosystem-edge/enclave-wizard/internal/models"
	"gopkg.in/yaml.v3"
)


type Validator struct {
	enclaveDir   string
	playbookPath string
	available    bool
}

func NewValidator(enclaveDir string) *Validator {
	if _, err := exec.LookPath("ansible-playbook"); err != nil {
		fmt.Println("WARNING: schema validation unavailable (ansible-playbook not found)")
		return &Validator{enclaveDir: enclaveDir}
	}

	playbook := filepath.Join(enclaveDir, "playbooks", "validation", "validate-schema.yaml")
	if _, err := os.Stat(playbook); err != nil {
		fmt.Println("WARNING: schema validation unavailable (playbook not found)")
		return &Validator{enclaveDir: enclaveDir}
	}

	fmt.Println("Schema validation enabled (using enclave's validate-schema.yaml)")
	return &Validator{enclaveDir: enclaveDir, playbookPath: playbook, available: true}
}

// Validate writes config to the enclave config/ directory, runs the Ansible
// schema validation playbook in-place, and rolls back to the previous config
// if validation fails.
func (v *Validator) Validate(cfg *models.EnclaveConfig) []models.ValidationError {
	if !v.available {
		return nil
	}

	configDir := filepath.Join(v.enclaveDir, "config")

	// Back up existing config files
	backupDir, err := os.MkdirTemp("", "enclave-wizard-backup-")
	if err != nil {
		return []models.ValidationError{{Message: fmt.Sprintf("failed to create backup dir: %v", err)}}
	}
	defer os.RemoveAll(backupDir)

	configFiles := []string{"global.yaml", "certificates.yaml", "cloud_infra.yaml"}
	for _, name := range configFiles {
		src := filepath.Join(configDir, name)
		if data, err := os.ReadFile(src); err == nil {
			os.WriteFile(filepath.Join(backupDir, name), data, 0640)
		}
	}

	// Write new config to the real config/ directory
	if err := writeConfigToDir(configDir, cfg); err != nil {
		return []models.ValidationError{{Message: fmt.Sprintf("failed to write config: %v", err)}}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	output, exitCode := runAnsibleValidation(ctx, v.playbookPath, v.enclaveDir)

	if exitCode == 0 {
		return nil
	}

	// Validation failed — rollback to previous config
	for _, name := range configFiles {
		backup := filepath.Join(backupDir, name)
		if data, err := os.ReadFile(backup); err == nil {
			os.WriteFile(filepath.Join(configDir, name), data, 0640)
		}
	}

	return parseAnsibleErrors(output)
}

// ValidateAndPersist validates config, then copies temp files to the real
// config directory only if validation passes.
func (v *Validator) ValidateAndPersist(cfg *models.EnclaveConfig) []models.ValidationError {
	if !v.available {
		return nil
	}

	errs := v.Validate(cfg)
	if len(errs) > 0 {
		return errs
	}

	return nil
}

func runAnsibleValidation(ctx context.Context, playbookPath, enclaveDir string) (string, int) {
	args := []string{
		playbookPath,
	}

	cmd := exec.CommandContext(ctx, "ansible-playbook", args...)
	cmd.Dir = enclaveDir
	cmd.Env = append(os.Environ(),
		"ANSIBLE_STDOUT_CALLBACK=json",
		"ANSIBLE_CALLBACKS_ENABLED=",
	)

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return stdout.String(), exitErr.ExitCode()
		}
		return stderr.String(), -1
	}
	return stdout.String(), 0
}

func parseAnsibleErrors(output string) []models.ValidationError {
	var errors []models.ValidationError

	// Try JSON callback format first (ANSIBLE_STDOUT_CALLBACK=json)
	errors = parseJSONCallbackOutput(output)
	if len(errors) > 0 {
		return errors
	}

	// Fallback: extract from "msg" in fatal lines
	fatalPattern := regexp.MustCompile(`fatal:.*FAILED!.*=>\s*(.+)`)
	for _, line := range strings.Split(output, "\n") {
		matches := fatalPattern.FindStringSubmatch(line)
		if len(matches) < 2 {
			continue
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(matches[1]), &result); err == nil {
			if msg, ok := result["msg"].(string); ok {
				for _, errLine := range strings.Split(msg, "\n") {
					errLine = strings.TrimSpace(errLine)
					if errLine != "" && !strings.HasPrefix(errLine, "Validation errors") {
						errors = append(errors, models.ValidationError{Message: errLine})
					}
				}
			}
		}
	}

	if len(errors) == 0 {
		trimmed := strings.TrimSpace(output)
		if len(trimmed) > 500 {
			trimmed = trimmed[len(trimmed)-500:]
		}
		errors = append(errors, models.ValidationError{Message: "Schema validation failed: " + trimmed})
	}
	return errors
}

type ansibleJSONOutput struct {
	Plays []struct {
		Tasks []struct {
			Hosts map[string]struct {
				Failed bool   `json:"failed"`
				Msg    string `json:"msg"`
				Errors []struct {
					Message    string `json:"message"`
					DataPath   string `json:"data_path"`
					SchemaPath string `json:"schema_path"`
				} `json:"errors"`
			} `json:"hosts"`
		} `json:"tasks"`
	} `json:"plays"`
}

func parseJSONCallbackOutput(output string) []models.ValidationError {
	var parsed ansibleJSONOutput
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		return nil
	}

	var errors []models.ValidationError
	for _, play := range parsed.Plays {
		for _, task := range play.Tasks {
			for _, result := range task.Hosts {
				if !result.Failed {
					continue
				}
				if len(result.Errors) > 0 {
					for _, e := range result.Errors {
						field := e.DataPath
						if field == "" {
							field = e.SchemaPath
						}
						errors = append(errors, models.ValidationError{
							Field:   field,
							Message: e.Message,
						})
					}
				} else if result.Msg != "" {
					for _, line := range strings.Split(result.Msg, "\n") {
						line = strings.TrimSpace(line)
						if line != "" && !strings.HasPrefix(line, "Validation errors") {
							errors = append(errors, models.ValidationError{Message: line})
						}
					}
				}
			}
		}
	}
	return errors
}

// wizardOnlyFields are fields in the Go model that don't exist in the
// enclave schema and should be stripped before validation.
var wizardOnlyFields = []string{
	"enabled_plugins", "enabledPlugins",
}

func writeConfigToDir(dir string, cfg *models.EnclaveConfig) error {
	globalMap := structToMap(cfg.Global)
	for _, key := range wizardOnlyFields {
		delete(globalMap, key)
	}

	// Map blockStorageBackend → storage_plugin (enclave renamed this field)
	if bsb, ok := globalMap["blockStorageBackend"]; ok {
		if _, hasSP := globalMap["storage_plugin"]; !hasSP {
			globalMap["storage_plugin"] = bsb
		}
		delete(globalMap, "blockStorageBackend")
	}
	delete(globalMap, "storagePlugin")

	if err := writeYAMLMap(filepath.Join(dir, "global.yaml"), globalMap); err != nil {
		return err
	}

	certsMap := structToMap(cfg.Certificates)
	if err := writeYAMLMap(filepath.Join(dir, "certificates.yaml"), certsMap); err != nil {
		return err
	}

	cloudInfraMap := structToMap(cfg.CloudInfra)
	if err := writeYAMLMap(filepath.Join(dir, "cloud_infra.yaml"), cloudInfraMap); err != nil {
		return err
	}

	return nil
}

func structToMap(v any) map[string]any {
	data, _ := json.Marshal(v)
	var m map[string]any
	json.Unmarshal(data, &m)
	// Remove null values
	for k, val := range m {
		if val == nil {
			delete(m, k)
		}
	}
	return m
}

func writeYAMLMap(path string, m map[string]any) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", filepath.Base(path), err)
	}
	return os.WriteFile(path, data, 0640)
}
