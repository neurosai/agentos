//go:build contracts

package contracts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
}

func loadSchema(t *testing.T, root, name string) *jsonschema.Schema {
	t.Helper()
	path := filepath.Join(root, "api", "jsonschema", name)
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open schema %s: %v", name, err)
	}
	defer f.Close()

	doc, err := jsonschema.UnmarshalJSON(f)
	if err != nil {
		t.Fatalf("unmarshal schema %s: %v", name, err)
	}

	url := "file://" + path
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(url, doc); err != nil {
		t.Fatalf("add resource: %v", err)
	}
	compiled, err := compiler.Compile(url)
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}
	return compiled
}

func validateJSONFile(t *testing.T, schema *jsonschema.Schema, root, rel string) {
	t.Helper()
	path := filepath.Join(root, rel)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	var doc any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("unmarshal %s: %v", rel, err)
	}
	if err := schema.Validate(doc); err != nil {
		t.Fatalf("validate %s: %v", rel, err)
	}
}

func TestMemoryRecordExample(t *testing.T) {
	root := repoRoot(t)
	schema := loadSchema(t, root, "memory-record.json")
	validateJSONFile(t, schema, root, "examples/memory/fact.json")
}

func TestCatalogEntityExample(t *testing.T) {
	root := repoRoot(t)
	schema := loadSchema(t, root, "catalog-entity.json")
	validateJSONFile(t, schema, root, "examples/catalog/component-payment-api.json")
}

func TestAgentManifestExample(t *testing.T) {
	root := repoRoot(t)
	schema := loadSchema(t, root, "agent-manifest.json")
	validateJSONFile(t, schema, root, "examples/agents/hermes-dev.json")
}
