package models

type PluginType string

const (
	PluginTypeFoundation PluginType = "foundation"
	PluginTypeAddon      PluginType = "addon"
)

type Plugin struct {
	Name        string     `json:"name" doc:"Plugin identifier"`
	Type        PluginType `json:"type" doc:"Plugin type" enum:"foundation,addon"`
	Description string     `json:"description" doc:"Human-readable description"`
}

// LVMS plugin configuration

type LVMSThinPoolConfig struct {
	Name               string `json:"name" yaml:"name" doc:"Thin pool name"`
	SizePercent        int    `json:"sizePercent" yaml:"sizePercent" doc:"Pool size as percentage of volume group" minimum:"1" maximum:"100"`
	OverprovisionRatio int    `json:"overprovisionRatio" yaml:"overprovisionRatio" doc:"Thin pool overprovisioning ratio" minimum:"1"`
}

type LVMSConfig struct {
	DeviceClassName     string             `json:"deviceClassName" yaml:"deviceClassName" doc:"LVM volume group device class name"`
	DefaultStorageClass bool               `json:"defaultStorageClass" yaml:"defaultStorageClass" doc:"Set as default StorageClass"`
	ThinPoolConfig      LVMSThinPoolConfig `json:"thinPoolConfig" yaml:"thinPoolConfig" doc:"Thin pool settings"`
}

// ODF plugin configuration

type ODFConfig struct {
	DefaultStorageClass bool `json:"defaultStorageClass" yaml:"defaultStorageClass" doc:"Set as default StorageClass"`
}

type ValidationError struct {
	Field   string `json:"field" doc:"JSON path to the invalid field"`
	Message string `json:"message" doc:"Human-readable error description"`
}
