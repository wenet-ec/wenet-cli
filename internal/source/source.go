// internal/source/source.go
package source

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Options struct {
	PackageFile string
	URL         string
	Ref         string
	Token       string
}

func FromEnv() Options {
	return Options{
		PackageFile: strings.TrimSpace(os.Getenv("PACKAGE_FILE")),
		URL:         strings.TrimSpace(os.Getenv("SOURCE_URL")),
		Ref:         strings.TrimSpace(os.Getenv("SOURCE_REF")),
		Token:       strings.TrimSpace(os.Getenv("SOURCE_TOKEN")),
	}
}

func (o Options) MergeEnv(env Options) Options {
	if o.PackageFile == "" {
		o.PackageFile = env.PackageFile
	}
	if o.URL == "" {
		o.URL = env.URL
	}
	if o.Ref == "" {
		o.Ref = env.Ref
	}
	if o.Token == "" {
		o.Token = env.Token
	}
	return o
}

func (o Options) IsRemote() bool {
	return o.URL != ""
}

func (o Options) IsPackageFile() bool {
	return o.PackageFile != ""
}

func (o Options) Validate() error {
	if o.PackageFile != "" && o.URL != "" {
		return fmt.Errorf("provide either package_file or source_url, not both")
	}

	if o.URL == "" {
		if o.Ref != "" {
			return fmt.Errorf("source_ref requires source_url")
		}
		if o.Token != "" {
			return fmt.Errorf("source_token requires source_url")
		}
		return nil
	}

	parsed, err := url.Parse(o.URL)
	if err != nil {
		return fmt.Errorf("source_url must use https://")
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("source_url must use https://")
	}
	if parsed.Host == "" {
		return fmt.Errorf("source_url must include a host")
	}
	if parsed.User != nil {
		return fmt.Errorf("source_url must not include credentials; use source_token")
	}
	return nil
}
