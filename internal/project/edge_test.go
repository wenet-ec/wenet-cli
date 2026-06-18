// internal/project/edge_test.go
package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppliesDefaults(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "edge.toml"), `
project = "site"
tag = "1.0.0"
script_path = "deploy.sh"
all = true
`)

	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DownloadBaseDir != "/tmp" {
		t.Fatalf("DownloadBaseDir = %q", cfg.DownloadBaseDir)
	}
	if cfg.Cleanup == nil || !*cfg.Cleanup {
		t.Fatalf("Cleanup = %v, want true", cfg.Cleanup)
	}
}

func TestValidateRequiresExactlyOneTarget(t *testing.T) {
	cfg := Config{
		Project:      "site",
		Tag:          "1.0.0",
		ScriptPath:   "deploy.sh",
		NodeIDs:      []string{"node-a"},
		ClusterNames: []string{"cluster-a"},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestValidateRejectsAbsoluteScriptPath(t *testing.T) {
	cfg := Config{
		Project:    "site",
		Tag:        "1.0.0",
		ScriptPath: "/deploy.sh",
		All:        true,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func writeFile(t *testing.T, path string, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}
