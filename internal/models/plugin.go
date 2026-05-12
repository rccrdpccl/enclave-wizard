package models

type PluginDescriptor struct {
	Name             string           `json:"name" yaml:"name" doc:"Plugin identifier matching directory name"`
	Type             string           `json:"type" yaml:"type" doc:"Plugin type" enum:"foundation,addon"`
	Order            *int             `json:"order,omitempty" yaml:"order,omitempty" doc:"Deployment order (lower deploys first)"`
	Mirror           *string          `json:"mirror,omitempty" yaml:"mirror,omitempty" doc:"Image mirroring strategy" enum:"core,plugin,none"`
	Catalog          *string          `json:"catalog,omitempty" yaml:"catalog,omitempty" doc:"Operator catalog" enum:"redhat,certified"`
	Operators        []PluginOperator `json:"operators,omitempty" yaml:"operators,omitempty" doc:"OLM operators to install"`
	InstallOperators *bool            `json:"installOperators,omitempty" yaml:"installOperators,omitempty" doc:"Install OLM operators during deploy (default: true)"`
	Defaults         any              `json:"defaults,omitempty" yaml:"defaults,omitempty" doc:"Default Ansible variables"`
}

type PluginOperator struct {
	Name    string `json:"name" yaml:"name" doc:"OLM package name"`
	Version string `json:"version" yaml:"version" doc:"Operator version"`
	Channel string `json:"channel" yaml:"channel" doc:"OLM update channel"`
}

type ValidationError struct {
	Field   string `json:"field" doc:"JSON path to the invalid field"`
	Message string `json:"message" doc:"Human-readable error description"`
}
