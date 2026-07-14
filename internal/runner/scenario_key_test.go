package runner

import "testing"

func TestScenarioKey(t *testing.T) {
	tests := []struct {
		playbook string
		tags     []string
		want     string
	}{
		{"playbooks/main.yaml", nil, "playbooks-main.yaml"},
		{"playbooks/03-deploy.yaml", nil, "playbooks-03-deploy.yaml"},
		{"playbooks/validation/validate-schema.yaml", []string{"validate-config"},
			"playbooks-validation-validate-schema.yaml--validate-config"},
		{"playbooks/validation/validate-schema.yaml", []string{"b-tag", "a-tag"},
			"playbooks-validation-validate-schema.yaml--a-tag,b-tag"},
		{"validations.sh", nil, "validations.sh"},
		{"playbooks/main.yaml", []string{}, "playbooks-main.yaml"},
	}
	for _, tt := range tests {
		got := ScenarioKey(tt.playbook, tt.tags)
		if got != tt.want {
			t.Errorf("ScenarioKey(%q, %v) = %q, want %q", tt.playbook, tt.tags, got, tt.want)
		}
	}
}
