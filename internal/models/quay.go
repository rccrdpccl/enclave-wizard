package models

type QuayBackendRGWConfiguration struct {
	AccessKey          string  `json:"access_key" yaml:"access_key" doc:"S3/RGW access key ID" minLength:"1"`
	SecretKey          string  `json:"secret_key" yaml:"secret_key" doc:"S3/RGW secret access key" minLength:"1"`
	BucketName         string  `json:"bucket_name" yaml:"bucket_name" doc:"S3 bucket name for Quay image storage" minLength:"1"`
	Hostname           string  `json:"hostname" yaml:"hostname" doc:"RadosGW/S3 endpoint hostname" minLength:"1"`
	IsSecure           *bool   `json:"is_secure,omitempty" yaml:"is_secure,omitempty" doc:"Use HTTPS for backend storage (default: true)"`
	Port               *int    `json:"port,omitempty" yaml:"port,omitempty" doc:"Backend storage endpoint port (default: 443)" minimum:"1" maximum:"65535"`
	MinimumChunkSizeMB *int    `json:"minimum_chunk_size_mb,omitempty" yaml:"minimum_chunk_size_mb,omitempty" doc:"Minimum multipart upload chunk size in MB" minimum:"1"`
	MaximumChunkSizeMB *int    `json:"maximum_chunk_size_mb,omitempty" yaml:"maximum_chunk_size_mb,omitempty" doc:"Maximum multipart upload chunk size in MB" minimum:"1"`
	ServerSideAssembly *bool   `json:"server_side_assembly,omitempty" yaml:"server_side_assembly,omitempty" doc:"Enable server-side assembly for multipart uploads"`
	StoragePath        *string `json:"storage_path,omitempty" yaml:"storage_path,omitempty" doc:"Object storage path prefix (must start with /)" pattern:"^/"`
}
