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

// VAST CSI plugin configuration

type VASTIPRange struct {
	Start string `json:"start" yaml:"start" doc:"Start IP address of the range"`
	End   string `json:"end" yaml:"end" doc:"End IP address of the range"`
}

type VASTVipPool struct {
	SubnetCIDR int           `json:"subnet_cidr" yaml:"subnet_cidr" doc:"Subnet prefix length" minimum:"1" maximum:"32"`
	IPRanges   []VASTIPRange `json:"ip_ranges" yaml:"ip_ranges" doc:"IP address ranges for VIP allocation" minItems:"1"`
}

type VASTTier struct {
	Name     string `json:"name" yaml:"name" doc:"Storage tier name (becomes StorageClass suffix)"`
	Protocol string `json:"protocol" yaml:"protocol" doc:"Storage protocol" enum:"nfs,block"`
}

type VASTConfig struct {
	InfraTenant  *string    `json:"infraTenant,omitempty" yaml:"infraTenant,omitempty" doc:"Tenant name for infrastructure resources"`
	StoragePath  *string    `json:"storagePath,omitempty" yaml:"storagePath,omitempty" doc:"Root path on VAST cluster"`
	ViewPolicyID *int       `json:"viewPolicyId,omitempty" yaml:"viewPolicyId,omitempty" doc:"VAST view policy ID"`
	QuayPvcSize  *string    `json:"quayPvcSize,omitempty" yaml:"quayPvcSize,omitempty" doc:"PVC size for Quay storage tier (e.g. 1000Gi)"`
	Tiers        []VASTTier `json:"tiers,omitempty" yaml:"tiers,omitempty" doc:"Storage tiers to create on the VAST cluster"`
}

type ValidationError struct {
	Field   string `json:"field" doc:"JSON path to the invalid field"`
	Message string `json:"message" doc:"Human-readable error description"`
}
