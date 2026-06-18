// internal/archive/archive_test.go
package archive

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestBuildCreatesArchiveWithIgnoreRules(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "edge.toml"), `
project = "site"
tag = "1.0.0"
all = true

[scripts]
linux = "deploy.sh"
windows = "deploy.ps1"
`)
	writeFile(t, filepath.Join(root, "deploy.sh"), "#!/bin/sh\necho deploy\n")
	writeFile(t, filepath.Join(root, "deploy.ps1"), "Write-Host deploy\n")
	writeFile(t, filepath.Join(root, "app.txt"), "app")
	writeFile(t, filepath.Join(root, ".gitignore"), "ignored.txt\n")
	writeFile(t, filepath.Join(root, ".edgeignore"), "secret.txt\n")
	writeFile(t, filepath.Join(root, "ignored.txt"), "ignored")
	writeFile(t, filepath.Join(root, "secret.txt"), "secret")

	result, err := Build(root, "")
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.FileCount != 6 {
		t.Fatalf("FileCount = %d, want 6", result.FileCount)
	}

	names := archiveNames(t, result.Path)
	want := []string{".edgeignore", ".gitignore", "app.txt", "deploy.ps1", "deploy.sh", "edge.toml"}
	if diffStrings(names, want) {
		t.Fatalf("archive names = %#v, want %#v", names, want)
	}
}

func TestBuildRejectsIgnoredScript(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "edge.toml"), `
project = "site"
tag = "1.0.0"
all = true

[scripts]
linux = "deploy.sh"
`)
	writeFile(t, filepath.Join(root, "deploy.sh"), "#!/bin/sh\n")
	writeFile(t, filepath.Join(root, ".edgeignore"), "deploy.sh\n")

	if _, err := Build(root, ""); err == nil {
		t.Fatal("Build() error = nil, want error")
	}
}

func archiveNames(t *testing.T, path string) []string {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()

	gz, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	names := []string{}
	for {
		header, err := tr.Next()
		if err != nil {
			break
		}
		names = append(names, header.Name)
	}
	sort.Strings(names)
	return names
}

func diffStrings(a []string, b []string) bool {
	if len(a) != len(b) {
		return true
	}
	for i := range a {
		if a[i] != b[i] {
			return true
		}
	}
	return false
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
