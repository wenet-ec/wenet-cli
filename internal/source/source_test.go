// internal/source/source_test.go
package source

import "testing"

func TestValidateLocalModeAllowsEmptySource(t *testing.T) {
	opts := Options{}
	if err := opts.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidatePackageFileMode(t *testing.T) {
	opts := Options{PackageFile: "dist/app.tar.gz"}
	if err := opts.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !opts.IsPackageFile() {
		t.Fatal("IsPackageFile() = false, want true")
	}
}

func TestValidateRejectsPackageFileWithSourceURL(t *testing.T) {
	opts := Options{
		PackageFile: "dist/app.tar.gz",
		URL:         "https://github.com/org/repo",
	}
	if err := opts.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestValidateRejectsRefWithoutURL(t *testing.T) {
	opts := Options{Ref: "main"}
	if err := opts.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestValidateRejectsTokenWithoutURL(t *testing.T) {
	opts := Options{Token: "secret"}
	if err := opts.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestValidateRequiresHTTPS(t *testing.T) {
	opts := Options{URL: "git@github.com:org/repo.git"}
	if err := opts.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}

	opts = Options{URL: "http://github.com/org/repo"}
	if err := opts.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestValidateRejectsCredentialsInURL(t *testing.T) {
	opts := Options{URL: "https://token@github.com/org/repo"}
	if err := opts.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestValidateAcceptsHTTPSRepo(t *testing.T) {
	opts := Options{
		URL:   "https://github.com/org/repo",
		Ref:   "main",
		Token: "secret",
	}
	if err := opts.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !opts.IsRemote() {
		t.Fatal("IsRemote() = false, want true")
	}
}

func TestMergeEnvUsesFlagsFirst(t *testing.T) {
	opts := Options{
		PackageFile: "flag.tar.gz",
		URL:         "https://github.com/org/flag",
		Ref:         "flag",
	}
	env := Options{
		PackageFile: "env.tar.gz",
		URL:         "https://github.com/org/env",
		Ref:         "env",
		Token:       "env-token",
	}
	merged := opts.MergeEnv(env)

	if merged.PackageFile != "flag.tar.gz" {
		t.Fatalf("PackageFile = %q", merged.PackageFile)
	}
	if merged.URL != "https://github.com/org/flag" {
		t.Fatalf("URL = %q", merged.URL)
	}
	if merged.Ref != "flag" {
		t.Fatalf("Ref = %q", merged.Ref)
	}
	if merged.Token != "env-token" {
		t.Fatalf("Token = %q", merged.Token)
	}
}
