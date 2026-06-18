// internal/project/edge.go
package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const ConfigFileName = "edge.toml"

type Config struct {
	Project         string   `toml:"project"`
	Tag             string   `toml:"tag"`
	ScriptPath      string   `toml:"script_path"`
	NodeIDs         []string `toml:"node_ids"`
	ClusterNames    []string `toml:"cluster_names"`
	Tags            []string `toml:"tags"`
	All             bool     `toml:"all"`
	SecretScope     string   `toml:"secret_scope"`
	DownloadBaseDir string   `toml:"download_base_dir"`
	Cleanup         *bool    `toml:"cleanup"`
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
	if c.ScriptPath == "" {
		return fmt.Errorf("edge.toml script_path is required")
	}
	if filepath.IsAbs(c.ScriptPath) {
		return fmt.Errorf("edge.toml script_path must be relative")
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
