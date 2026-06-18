// internal/config/config.go
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const DefaultServer = "https://api.wenet-ec.com"

type Profile struct {
	Server string `toml:"server"`
	Token  string `toml:"token"`
}

type File struct {
	Profiles map[string]Profile
}

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config directory: %w", err)
	}
	return filepath.Join(dir, "wenet", "config.toml"), nil
}

func LoadFile(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &File{Profiles: map[string]Profile{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	profiles := map[string]Profile{}
	if len(bytes.TrimSpace(data)) > 0 {
		if err := toml.Unmarshal(data, &profiles); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	}
	return &File{Profiles: profiles}, nil
}

func SaveFile(path string, file *File) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data := render(file)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func ResolveCredential(profile string) (Profile, error) {
	if token := os.Getenv("API_TOKEN"); token != "" {
		server := os.Getenv("API_SERVER")
		if server == "" {
			server = DefaultServer
		}
		return Profile{Server: server, Token: token}, nil
	}

	path, err := DefaultPath()
	if err != nil {
		return Profile{}, err
	}
	file, err := LoadFile(path)
	if err != nil {
		return Profile{}, err
	}
	cred, ok := file.Profiles[profile]
	if !ok {
		return Profile{}, fmt.Errorf("profile %q not found; run wenet login <token>", profile)
	}
	if cred.Server == "" {
		cred.Server = DefaultServer
	}
	if cred.Token == "" {
		return Profile{}, fmt.Errorf("profile %q is missing token", profile)
	}
	return cred, nil
}

func render(file *File) []byte {
	names := make([]string, 0, len(file.Profiles))
	for name := range file.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	for i, name := range names {
		if i > 0 {
			b.WriteString("\n")
		}
		profile := file.Profiles[name]
		if profile.Server == "" {
			profile.Server = DefaultServer
		}
		fmt.Fprintf(&b, "[%s]\n", quoteKey(name))
		fmt.Fprintf(&b, "server = %q\n", profile.Server)
		fmt.Fprintf(&b, "token = %q\n", profile.Token)
	}
	return []byte(b.String())
}

func quoteKey(key string) string {
	if key == "" {
		return `""`
	}
	for _, r := range key {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' && r != '-' {
			return fmt.Sprintf("%q", key)
		}
	}
	return key
}
