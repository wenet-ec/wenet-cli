// internal/cli/logout_test.go
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/wenet-ec/wenet-cli/internal/config"
)

func TestLogoutRemovesSelectedProfile(t *testing.T) {
	home := isolatedHome(t)
	path, err := config.DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error = %v", err)
	}
	if err := config.SaveFile(path, &config.File{Profiles: map[string]config.Profile{
		"default": {Server: config.DefaultServer, Token: "default-token"},
		"work":    {Server: config.DefaultServer, Token: "work-token"},
	}}); err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--profile", "work", "logout"})
	cmd.SetOut(&bytes.Buffer{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	file, err := config.LoadFile(filepath.Join(home, ".config", "wenet", "config.toml"))
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	if _, ok := file.Profiles["work"]; ok {
		t.Fatal("work profile still exists")
	}
	if _, ok := file.Profiles["default"]; !ok {
		t.Fatal("default profile was removed")
	}
}

func TestLogoutMissingProfileReturnsError(t *testing.T) {
	isolatedHome(t)

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"logout"})
	cmd.SetOut(&bytes.Buffer{})

	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
}

func isolatedHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", home)
		t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	} else {
		t.Setenv("HOME", home)
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	}
	if err := os.MkdirAll(filepath.Join(home, ".config"), 0o755); err != nil {
		t.Fatal(err)
	}
	return home
}
