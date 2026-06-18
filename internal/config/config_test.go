// internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveAndLoadFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	file := &File{Profiles: map[string]Profile{
		"default": {
			Server: DefaultServer,
			Token:  "tok_default",
		},
		"staging": {
			Server: "https://staging.example.test",
			Token:  "tok_staging",
		},
	}}

	if err := SaveFile(path, file); err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("config mode = %o, want 0600", got)
	}

	loaded, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	if got := loaded.Profiles["staging"].Token; got != "tok_staging" {
		t.Fatalf("staging token = %q", got)
	}
}

func TestResolveCredentialUsesEnvironment(t *testing.T) {
	t.Setenv("API_TOKEN", "env_token")
	t.Setenv("API_SERVER", "https://api.example.test")

	cred, err := ResolveCredential("missing")
	if err != nil {
		t.Fatalf("ResolveCredential() error = %v", err)
	}
	if cred.Token != "env_token" {
		t.Fatalf("token = %q", cred.Token)
	}
	if cred.Server != "https://api.example.test" {
		t.Fatalf("server = %q", cred.Server)
	}
}

func TestRenderQuotesProfileWhenNeeded(t *testing.T) {
	data := render(&File{Profiles: map[string]Profile{
		"work profile": {
			Token: "tok",
		},
	}})

	if !strings.Contains(string(data), `["work profile"]`) {
		t.Fatalf("rendered config did not quote profile name:\n%s", data)
	}
	if !strings.Contains(string(data), `server = "https://api.wenet-ec.com"`) {
		t.Fatalf("rendered config did not include default server:\n%s", data)
	}
}
