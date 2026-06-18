// internal/archive/archive.go
package archive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
	"github.com/wenet-ec/wenet-cli/internal/project"
)

type Result struct {
	Path      string
	FileCount int
	Config    *project.Config
}

func Build(root string, output string) (*Result, error) {
	cfg, err := project.Load(root)
	if err != nil {
		return nil, err
	}

	if output == "" {
		output = filepath.Join(root, ".wenet", archiveName(cfg))
	}
	if !filepath.IsAbs(output) {
		output = filepath.Join(root, output)
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return nil, fmt.Errorf("create archive directory: %w", err)
	}

	matcher, err := buildIgnoreMatcher(root)
	if err != nil {
		return nil, err
	}

	if err := validateScript(root, cfg.ScriptPath, matcher); err != nil {
		return nil, err
	}

	tmp := output + ".tmp"
	file, err := os.Create(tmp)
	if err != nil {
		return nil, fmt.Errorf("create archive: %w", err)
	}
	success := false
	defer func() {
		_ = file.Close()
		if !success {
			_ = os.Remove(tmp)
		}
	}()

	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)

	count, err := writeArchive(root, output, matcher, tw)
	closeErr := tw.Close()
	gzErr := gz.Close()
	fileErr := file.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close tar archive: %w", closeErr)
	}
	if gzErr != nil {
		return nil, fmt.Errorf("close gzip archive: %w", gzErr)
	}
	if fileErr != nil {
		return nil, fmt.Errorf("close archive file: %w", fileErr)
	}
	if err := os.Rename(tmp, output); err != nil {
		return nil, fmt.Errorf("move archive into place: %w", err)
	}
	success = true

	return &Result{Path: output, FileCount: count, Config: cfg}, nil
}

func archiveName(cfg *project.Config) string {
	projectName := safeName(cfg.Project)
	tag := safeName(cfg.Tag)
	return fmt.Sprintf("%s-%s.tar.gz", projectName, tag)
}

func safeName(value string) string {
	value = strings.TrimSpace(value)
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	if b.Len() == 0 {
		return "package"
	}
	return b.String()
}

func buildIgnoreMatcher(root string) (*ignore.GitIgnore, error) {
	patterns := []string{".git", ".git/**", ".wenet", ".wenet/**"}
	for _, name := range []string{".gitignore", ".edgeignore"} {
		path := filepath.Join(root, name)
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			patterns = append(patterns, line)
		}
	}
	return ignore.CompileIgnoreLines(patterns...), nil
}

func validateScript(root string, scriptPath string, matcher *ignore.GitIgnore) error {
	clean := filepath.Clean(scriptPath)
	if clean == "." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
		return fmt.Errorf("edge.toml script_path must stay inside the project")
	}
	if matcher.MatchesPath(filepath.ToSlash(clean)) {
		return fmt.Errorf("edge.toml script_path %q is ignored by .gitignore or .edgeignore", scriptPath)
	}
	info, err := os.Stat(filepath.Join(root, clean))
	if err != nil {
		return fmt.Errorf("edge.toml script_path %q does not exist", scriptPath)
	}
	if info.IsDir() {
		return fmt.Errorf("edge.toml script_path %q is a directory", scriptPath)
	}
	return nil
}

func writeArchive(root string, output string, matcher *ignore.GitIgnore, tw *tar.Writer) (int, error) {
	count := 0
	outputAbs, err := filepath.Abs(output)
	if err != nil {
		return 0, err
	}

	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		pathAbs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if pathAbs == outputAbs || pathAbs == outputAbs+".tmp" {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if matcher.MatchesPath(rel) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}

		if err := addFile(root, path, rel, tw); err != nil {
			return err
		}
		count++
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("build archive: %w", err)
	}
	return count, nil
}

func addFile(root string, path string, rel string, tw *tar.Writer) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = rel
	header.ModTime = info.ModTime()
	header.AccessTime = info.ModTime()
	header.ChangeTime = info.ModTime()

	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	if _, err := io.Copy(tw, file); err != nil {
		return err
	}
	return nil
}
