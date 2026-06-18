// internal/cli/source_flags_test.go
package cli

import (
	"strings"
	"testing"
)

func TestPushRejectsSSHSourceURL(t *testing.T) {
	t.Setenv("API_TOKEN", "token")

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"push", "--source-url", "git@github.com:org/repo.git"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "source_url must use https://") {
		t.Fatalf("error = %q", err)
	}
}

func TestDeployAcceptsSourceFlagsBeforeEndpointWiring(t *testing.T) {
	t.Setenv("API_TOKEN", "token")

	cmd := NewRootCommand()
	cmd.SetArgs([]string{
		"deploy",
		"--source-url",
		"https://github.com/org/repo",
		"--source-ref",
		"main",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want placeholder error")
	}
	if !strings.Contains(err.Error(), "repo import and rollout creation") {
		t.Fatalf("error = %q", err)
	}
}

func TestPushRejectsPackageFileAndSourceURLTogether(t *testing.T) {
	t.Setenv("API_TOKEN", "token")

	cmd := NewRootCommand()
	cmd.SetArgs([]string{
		"push",
		"--package-file",
		"dist/app.tar.gz",
		"--source-url",
		"https://github.com/org/repo",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "provide either package_file or source_url") {
		t.Fatalf("error = %q", err)
	}
}

func TestPushAcceptsPackageFileBeforeEndpointWiring(t *testing.T) {
	t.Setenv("API_TOKEN", "token")

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"push", "--package-file", "dist/app.tar.gz"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want placeholder error")
	}
	if !strings.Contains(err.Error(), "package file upload") {
		t.Fatalf("error = %q", err)
	}
}
