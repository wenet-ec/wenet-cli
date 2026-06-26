// internal/cli/source_flags_test.go
package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestDeployWithRemoteSourceCreatesRollout(t *testing.T) {
	root := testProject(t)
	server := deploymentAPIServer(t)
	t.Setenv("API_TOKEN", "token")
	t.Setenv("API_SERVER", server.URL)
	chdir(t, root)

	cmd := NewRootCommand()
	cmd.SetArgs([]string{
		"deploy",
		"--source-url",
		"https://github.com/org/repo",
		"--source-ref",
		"main",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
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

func TestPushAcceptsPackageFile(t *testing.T) {
	root := testProject(t)
	packagePath := filepath.Join(root, "dist", "app.tar.gz")
	writeFile(t, packagePath, "fake package")
	server := deploymentAPIServer(t)
	t.Setenv("API_TOKEN", "token")
	t.Setenv("API_SERVER", server.URL)
	chdir(t, root)

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"push", "--package-file", packagePath})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func deploymentAPIServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/public/v1/projects/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeEnvelope(t, w, []map[string]any{})
		case http.MethodPost:
			writeEnvelopeStatus(t, w, http.StatusCreated, map[string]any{
				"id":   "project-1",
				"name": "site",
			})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/public/v1/packages/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeEnvelopeStatus(t, w, http.StatusCreated, map[string]any{
			"id":      "package-1",
			"project": "project-1",
			"tag":     "1.0.0",
		})
	})
	mux.HandleFunc("/api/public/v1/rollouts/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeEnvelopeStatus(t, w, http.StatusCreated, map[string]any{
			"id":               "rollout-1",
			"package":          "package-1",
			"deployment_ids":   []string{"deployment-1"},
			"deployment_count": 1,
		})
	})
	return httptest.NewServer(mux)
}

func testProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "edge.toml"), `
project = "site"
tag = "1.0.0"
all = true

[scripts]
linux = "deploy.sh"
`)
	writeFile(t, filepath.Join(root, "deploy.sh"), "#!/bin/sh\necho deploy\n")
	return root
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previous)
	})
}

func writeEnvelope(t *testing.T, w http.ResponseWriter, data any) {
	t.Helper()
	writeEnvelopeStatus(t, w, http.StatusOK, data)
}

func writeEnvelopeStatus(t *testing.T, w http.ResponseWriter, status int, data any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]any{"data": data}); err != nil {
		t.Fatal(err)
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
