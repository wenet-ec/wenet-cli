// internal/project/edge.go
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/pelletier/go-toml/v2"
)

const ConfigFileName = "edge.toml"

type Config struct {
	Project         string   `toml:"project"`
	Tag             string   `toml:"tag"`
	Scripts         Scripts  `toml:"scripts"`
	NodeIDs         []string `toml:"node_ids"`
	ClusterNames    []string `toml:"cluster_names"`
	Tags            []string `toml:"tags"`
	All             bool     `toml:"all"`
	SecretScope     string   `toml:"secret_scope"`
	DownloadBaseDir string   `toml:"download_base_dir"`
	Cleanup         *bool    `toml:"cleanup"`
}

type Scripts struct {
	Linux   string `toml:"linux"`
	Darwin  string `toml:"darwin"`
	Windows string `toml:"windows"`
}

func (s Scripts) Paths() map[string]string {
	paths := map[string]string{}
	if s.Linux != "" {
		paths["linux"] = s.Linux
	}
	if s.Darwin != "" {
		paths["darwin"] = s.Darwin
	}
	if s.Windows != "" {
		paths["windows"] = s.Windows
	}
	return paths
}

func (s Scripts) Platforms() []string {
	paths := s.Paths()
	platforms := make([]string, 0, len(paths))
	for platform := range paths {
		platforms = append(platforms, platform)
	}
	sort.Strings(platforms)
	return platforms
}

func Load(root string) (*Config, error) {
	path := filepath.Join(root, ConfigFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", ConfigFileName, err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", ConfigFileName, err)
	}
	if cfg.Cleanup == nil {
		cleanup := true
		cfg.Cleanup = &cleanup
	}
	if cfg.DownloadBaseDir == "" {
		cfg.DownloadBaseDir = "/tmp"
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c Config) Validate() error {
	if c.Project == "" {
		return fmt.Errorf("edge.toml project is required")
	}
	if c.Tag == "" {
		return fmt.Errorf("edge.toml tag is required")
	}
	scripts := c.Scripts.Paths()
	if len(scripts) == 0 {
		return fmt.Errorf("edge.toml [scripts] must include at least one of linux, darwin, or windows")
	}
	for platform, scriptPath := range scripts {
		if filepath.IsAbs(scriptPath) {
			return fmt.Errorf("edge.toml scripts.%s must be relative", platform)
		}
	}

	targets := 0
	if len(c.NodeIDs) > 0 {
		targets++
	}
	if len(c.ClusterNames) > 0 {
		targets++
	}
	if len(c.Tags) > 0 {
		targets++
	}
	if c.All {
		targets++
	}
	if targets != 1 {
		return fmt.Errorf("edge.toml must specify exactly one targeting form: node_ids, cluster_names, tags, or all")
	}
	return nil
}
