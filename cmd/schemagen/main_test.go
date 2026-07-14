package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// --- Name conversion tests ---

func TestSnakeToPascal(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"deploy_database", "DeployDatabase"},
		{"instances", "Instances"},
		{"db_size", "DBSize"},
		{"admin_password", "AdminPassword"},
		{"log_level", "LogLevel"},
		{"simple", "Simple"},
		{"rhbk", "Rhbk"},
		{"trust_manager", "TrustManager"},
		{"api_url", "APIURL"},
		{"bmc_ip", "BMCIP"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := SnakeToPascal(tc.input)
			if got != tc.want {
				t.Errorf("SnakeToPascal(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSnakeToCamel(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"rhbk_instances", "rhbkInstances"},
		{"rhbk_deploy_database", "rhbkDeployDatabase"},
		{"rhbk_db_size", "rhbkDBSize"},
		{"simple_enabled", "simpleEnabled"},
		{"simple_name", "simpleName"},
		{"osac_api_url", "osacAPIURL"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := SnakeToCamel(tc.input)
			if got != tc.want {
				t.Errorf("SnakeToCamel(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestStripPrefix(t *testing.T) {
	cases := []struct {
		propName string
		prefix   string
		want     string
	}{
		{"rhbk_instances", "rhbk", "instances"},
		{"rhbk_deploy_database", "rhbk", "deploy_database"},
		{"rhbk_db_size", "rhbk", "db_size"},
		{"simple_enabled", "simple", "enabled"},
		{"unrelated_field", "rhbk", "unrelated_field"},
		{"rhbk", "rhbk", "rhbk"}, // no underscore after prefix
	}
	for _, tc := range cases {
		name := tc.propName + "/" + tc.prefix
		t.Run(name, func(t *testing.T) {
			got := StripPrefix(tc.propName, tc.prefix)
			if got != tc.want {
				t.Errorf("StripPrefix(%q, %q) = %q, want %q", tc.propName, tc.prefix, got, tc.want)
			}
		})
	}
}

// --- Schema parsing tests ---

func TestParseSchema(t *testing.T) {
	data := `
type: object
required:
  - foo_name
properties:
  foo_name:
    type: string
    description: The name
  foo_count:
    type: integer
    description: How many
    minimum: 0
    maximum: 100
`
	var schema JSONSchema
	if err := yaml.Unmarshal([]byte(data), &schema); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if schema.Type != "object" {
		t.Errorf("type = %q, want %q", schema.Type, "object")
	}
	if len(schema.Properties) != 2 {
		t.Errorf("properties count = %d, want 2", len(schema.Properties))
	}
	if len(schema.Required) != 1 || schema.Required[0] != "foo_name" {
		t.Errorf("required = %v, want [foo_name]", schema.Required)
	}

	nameProp := schema.Properties["foo_name"]
	if nameProp.Type != "string" {
		t.Errorf("foo_name type = %q, want %q", nameProp.Type, "string")
	}
	if nameProp.Description != "The name" {
		t.Errorf("foo_name description = %q, want %q", nameProp.Description, "The name")
	}

	countProp := schema.Properties["foo_count"]
	if countProp.Minimum == nil || *countProp.Minimum != 0 {
		t.Errorf("foo_count minimum = %v, want 0", countProp.Minimum)
	}
	if countProp.Maximum == nil || *countProp.Maximum != 100 {
		t.Errorf("foo_count maximum = %v, want 100", countProp.Maximum)
	}
}

func TestParseSchemaWithEnum(t *testing.T) {
	data := `
type: object
properties:
  level:
    type: string
    enum:
      - low
      - medium
      - high
`
	var schema JSONSchema
	if err := yaml.Unmarshal([]byte(data), &schema); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	prop := schema.Properties["level"]
	if len(prop.Enum) != 3 {
		t.Fatalf("enum length = %d, want 3", len(prop.Enum))
	}
	if prop.Enum[0] != "low" || prop.Enum[1] != "medium" || prop.Enum[2] != "high" {
		t.Errorf("enum = %v, want [low medium high]", prop.Enum)
	}
}

// --- Pointer vs value type tests ---

func TestPointerTypesOptionalVsRequired(t *testing.T) {
	schema := &JSONSchema{
		Type:     "object",
		Required: []string{"test_name"},
		Properties: map[string]SchemaProperty{
			"test_name":  {Type: "string", Description: "Required field"},
			"test_count": {Type: "integer", Description: "Optional field"},
		},
	}

	code, err := generatePluginFile("test", schema)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	src := string(code)

	// Required field: value type (string, not *string)
	if !containsField(src, "Name", "string") {
		t.Errorf("expected required field to be value type (string), got:\n%s", src)
	}

	// Optional field: pointer type (*int)
	if !containsField(src, "Count", "*int") {
		t.Errorf("expected optional field to be pointer type (*int), got:\n%s", src)
	}
}

// --- Validation tag generation tests ---

func TestValidationTags(t *testing.T) {
	schema := &JSONSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"v_min_field":     {Type: "integer", Minimum: ptrFloat(1)},
			"v_max_field":     {Type: "integer", Maximum: ptrFloat(100)},
			"v_pattern_field": {Type: "string", Pattern: "^[a-z]+$"},
			"v_enum_field":    {Type: "string", Enum: []string{"a", "b", "c"}},
			"v_minlen_field":  {Type: "string", MinLength: ptrInt(5)},
			"v_minitems":      {Type: "array", MinItems: ptrInt(1)},
		},
	}

	code, err := generatePluginFile("v", schema)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	src := string(code)

	checks := []struct {
		name string
		want string
	}{
		{"minimum tag", `minimum:"1"`},
		{"maximum tag", `maximum:"100"`},
		{"pattern tag", `pattern:"^[a-z]+$"`},
		{"enum tag", `enum:"a,b,c"`},
		{"minLength tag", `minLength:"5"`},
		{"minItems tag", `minItems:"1"`},
	}

	for _, tc := range checks {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(src, tc.want) {
				t.Errorf("expected %s in output, got:\n%s", tc.want, src)
			}
		})
	}
}

// --- Doc tag tests ---

func TestDocTagFromDescription(t *testing.T) {
	schema := &JSONSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"d_name": {Type: "string", Description: "The name of the thing"},
		},
	}

	code, err := generatePluginFile("d", schema)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	src := string(code)

	if !strings.Contains(src, `doc:"The name of the thing"`) {
		t.Errorf("expected doc tag in output, got:\n%s", src)
	}
}

// --- Full generation tests ---

func TestGenerateRHBKFromFixture(t *testing.T) {
	schema := loadFixtureSchema(t, "rhbk")

	code, err := generatePluginFile("rhbk", schema)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	src := string(code)

	// Struct name
	if !strings.Contains(src, "type RhbkConfig struct") {
		t.Errorf("expected 'type RhbkConfig struct', got:\n%s", src)
	}

	// Prefix stripping: rhbk_instances -> Instances (not RhbkInstances)
	if !containsField(src, "Instances", "int") {
		t.Errorf("expected 'Instances int' (required, value type, prefix stripped), got:\n%s", src)
	}

	// YAML tag preserves original
	if !strings.Contains(src, `yaml:"rhbk_instances"`) {
		t.Errorf("expected yaml tag to preserve original name, got:\n%s", src)
	}

	// JSON tag is camelCase
	if !strings.Contains(src, `json:"rhbkInstances"`) {
		t.Errorf("expected json tag in camelCase, got:\n%s", src)
	}

	// Optional field is pointer
	if !containsField(src, "DeployDatabase", "*bool") {
		t.Errorf("expected 'DeployDatabase *bool' (optional), got:\n%s", src)
	}

	// Doc tag
	if !strings.Contains(src, `doc:"Number of Keycloak replicas"`) {
		t.Errorf("expected doc tag for instances, got:\n%s", src)
	}

	// Minimum validation
	if !strings.Contains(src, `minimum:"1"`) {
		t.Errorf("expected minimum tag, got:\n%s", src)
	}

	// DO NOT EDIT header
	if !strings.Contains(src, "Code generated by schemagen. DO NOT EDIT.") {
		t.Errorf("expected DO NOT EDIT header, got:\n%s", src)
	}
}

func TestGenerateSimpleFromFixture(t *testing.T) {
	schema := loadFixtureSchema(t, "simple")

	code, err := generatePluginFile("simple", schema)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	src := string(code)

	if !strings.Contains(src, "type SimpleConfig struct") {
		t.Errorf("expected 'type SimpleConfig struct', got:\n%s", src)
	}

	// Required field: value type
	if !containsField(src, "Enabled", "bool") {
		t.Errorf("expected 'Enabled bool' (required), got:\n%s", src)
	}

	// Optional field: pointer type
	if !containsField(src, "Name", "*string") {
		t.Errorf("expected 'Name *string' (optional), got:\n%s", src)
	}
}

// --- Registry generation tests ---

func TestGenerateRegistry(t *testing.T) {
	plugins := []pluginInfo{
		{Name: "rhbk", Schema: &JSONSchema{}},
		{Name: "simple", Schema: &JSONSchema{}},
	}

	code, err := generateRegistryFile(plugins)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	src := string(code)

	if !strings.Contains(src, "Code generated by schemagen. DO NOT EDIT.") {
		t.Errorf("expected DO NOT EDIT header")
	}
	if !strings.Contains(src, "PluginConfigTypes") {
		t.Errorf("expected PluginConfigTypes map")
	}
	if !strings.Contains(src, `"rhbk"`) {
		t.Errorf("expected rhbk entry")
	}
	if !strings.Contains(src, `"simple"`) {
		t.Errorf("expected simple entry")
	}
	if !strings.Contains(src, "RhbkConfig{}") {
		t.Errorf("expected RhbkConfig constructor")
	}
	if !strings.Contains(src, "SimpleConfig{}") {
		t.Errorf("expected SimpleConfig constructor")
	}
}

// --- Generated code compiles ---

func TestGeneratedCodeCompiles(t *testing.T) {
	rhbkSchema := loadFixtureSchema(t, "rhbk")
	simpleSchema := loadFixtureSchema(t, "simple")

	rhbkCode, err := generatePluginFile("rhbk", rhbkSchema)
	if err != nil {
		t.Fatalf("generate rhbk: %v", err)
	}

	simpleCode, err := generatePluginFile("simple", simpleSchema)
	if err != nil {
		t.Fatalf("generate simple: %v", err)
	}

	plugins := []pluginInfo{
		{Name: "rhbk", Schema: rhbkSchema},
		{Name: "simple", Schema: simpleSchema},
	}
	registryCode, err := generateRegistryFile(plugins)
	if err != nil {
		t.Fatalf("generate registry: %v", err)
	}

	// Write all generated files to a temp directory and parse them
	dir := t.TempDir()
	files := map[string][]byte{
		"plugin_rhbk.go":   rhbkCode,
		"plugin_simple.go": simpleCode,
		"registry.go":      registryCode,
	}

	fset := token.NewFileSet()
	for name, code := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, code, 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}

		_, err := parser.ParseFile(fset, path, code, parser.AllErrors)
		if err != nil {
			t.Errorf("file %s does not parse: %v\nsource:\n%s", name, err, string(code))
		}
	}
}

// --- Plugin discovery tests ---

func TestDiscoverPlugins(t *testing.T) {
	testdataDir := filepath.Join("testdata")
	plugins, err := discoverPlugins(testdataDir)
	if err != nil {
		t.Fatalf("discoverPlugins: %v", err)
	}

	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(plugins))
	}

	// Sorted alphabetically
	if plugins[0].Name != "rhbk" {
		t.Errorf("first plugin = %q, want %q", plugins[0].Name, "rhbk")
	}
	if plugins[1].Name != "simple" {
		t.Errorf("second plugin = %q, want %q", plugins[1].Name, "simple")
	}
}

// --- Type mapping tests ---

func TestSchemaTypeToGo(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"string", "string"},
		{"integer", "int"},
		{"number", "float64"},
		{"boolean", "bool"},
		{"array", "[]string"},
		{"object", "map[string]any"},
		{"unknown", "any"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := schemaTypeToGo(tc.input)
			if got != tc.want {
				t.Errorf("schemaTypeToGo(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// --- Format number tests ---

func TestFormatNumber(t *testing.T) {
	cases := []struct {
		input float64
		want  string
	}{
		{1, "1"},
		{100, "100"},
		{1.5, "1.5"},
		{0, "0"},
		{-1, "-1"},
	}
	for _, tc := range cases {
		got := formatNumber(tc.input)
		if got != tc.want {
			t.Errorf("formatNumber(%v) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// --- Dual tag tests ---

func TestDualTags(t *testing.T) {
	schema := &JSONSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"test_some_field": {Type: "string", Description: "A field"},
		},
	}

	code, err := generatePluginFile("test", schema)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	src := string(code)

	// YAML tag preserves original snake_case
	if !strings.Contains(src, `yaml:"test_some_field,omitempty"`) {
		t.Errorf("expected yaml tag with original name, got:\n%s", src)
	}

	// JSON tag uses camelCase
	if !strings.Contains(src, `json:"testSomeField,omitempty"`) {
		t.Errorf("expected json tag with camelCase, got:\n%s", src)
	}
}

// --- Enum tag test ---

func TestEnumTag(t *testing.T) {
	schema := &JSONSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"e_level": {
				Type: "string",
				Enum: []string{"DEBUG", "INFO", "WARN", "ERROR"},
			},
		},
	}

	code, err := generatePluginFile("e", schema)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	src := string(code)

	if !strings.Contains(src, `enum:"DEBUG,INFO,WARN,ERROR"`) {
		t.Errorf("expected enum tag, got:\n%s", src)
	}
}

// --- MaxItems tag test ---

func TestMaxItemsTag(t *testing.T) {
	maxItems := 5
	schema := &JSONSchema{
		Type: "object",
		Properties: map[string]SchemaProperty{
			"m_items": {
				Type:     "array",
				MaxItems: &maxItems,
			},
		},
	}

	code, err := generatePluginFile("m", schema)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	src := string(code)

	if !strings.Contains(src, `maxItems:"5"`) {
		t.Errorf("expected maxItems tag, got:\n%s", src)
	}
}

// --- Helpers ---

// containsField checks whether the generated source contains a field with the
// given name and type, ignoring gofmt alignment whitespace.
func containsField(src, fieldName, fieldType string) bool {
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		// Field lines look like: FieldName  Type  `tags...`
		parts := strings.Fields(trimmed)
		if len(parts) >= 2 && parts[0] == fieldName && parts[1] == fieldType {
			return true
		}
	}
	return false
}

func loadFixtureSchema(t *testing.T, pluginName string) *JSONSchema {
	t.Helper()
	path := filepath.Join("testdata", "plugins", pluginName, "schemas", "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture %s: %v", path, err)
	}
	var schema JSONSchema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		t.Fatalf("parsing fixture %s: %v", path, err)
	}
	return &schema
}

func ptrFloat(f float64) *float64 { return &f }
func ptrInt(i int) *int           { return &i }
