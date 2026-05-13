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
	Requires         *PluginRequires  `json:"requires,omitempty" yaml:"requires,omitempty" doc:"Prerequisites for the plugin"`
	Registries       []PluginRegistry `json:"registries,omitempty" yaml:"registries,omitempty" doc:"Registry mirror entries for MCE patching"`
}

type PluginOperator struct {
	Name        string `json:"name" yaml:"name" doc:"OLM package name"`
	Version     string `json:"version" yaml:"version" doc:"Operator version"`
	Channel     string `json:"channel" yaml:"channel" doc:"OLM update channel"`
	InitVersion string `json:"initVersion,omitempty" yaml:"init_version,omitempty" doc:"Bootstrap version for seed phase"`
	Namespace   string `json:"namespace,omitempty" yaml:"namespace,omitempty" doc:"Install namespace"`
	Source      string `json:"source,omitempty" yaml:"source,omitempty" doc:"CatalogSource name"`
}

type PluginRequires struct {
	Vars  []PluginRequiresVar  `json:"vars,omitempty" yaml:"vars,omitempty" doc:"Required Ansible variables"`
	Files []PluginRequiresFile `json:"files,omitempty" yaml:"files,omitempty" doc:"Required files within plugin directory"`
}

type PluginRequiresVar struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	When        string `json:"when,omitempty" yaml:"when,omitempty"`
}

type PluginRequiresFile struct {
	Path        string `json:"path" yaml:"path"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	When        string `json:"when,omitempty" yaml:"when,omitempty"`
}

type PluginRegistry struct {
	Location string `json:"location" yaml:"location" doc:"Source registry address"`
	Mirror   string `json:"mirror" yaml:"mirror" doc:"Path under internal Quay registry"`
}

type ValidationError struct {
	Field   string `json:"field" doc:"JSON path to the invalid field"`
	Message string `json:"message" doc:"Human-readable error description"`
}
