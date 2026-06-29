// internal/api/client.go
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(baseURL string, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) Token() string {
	return c.token
}

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Package struct {
	ID        string         `json:"id"`
	Project   string         `json:"project"`
	Tag       string         `json:"tag"`
	Scripts   map[string]any `json:"scripts"`
	SizeBytes int64          `json:"size_bytes"`
	SourceURL string         `json:"source_url"`
	SourceRef string         `json:"source_ref"`
}

type SecretScope struct {
	ID      string `json:"id"`
	Project string `json:"project"`
	Name    string `json:"name"`
}

type Rollout struct {
	ID              string   `json:"id"`
	Package         string   `json:"package"`
	DeploymentIDs   []string `json:"deployment_ids"`
	DeploymentCount int      `json:"deployment_count"`
}

type PackageSource struct {
	FilePath    string
	SourceURL   string
	SourceRef   string
	SourceToken string
}

type RolloutInput struct {
	PackageID       string
	SecretScopeID   string
	DownloadBaseDir string
	Cleanup         bool
	Targeting       map[string]any
}

type envelope[T any] struct {
	Data T `json:"data"`
}

func (c *Client) EnsureProject(name string) (*Project, error) {
	project, err := c.FindProjectByName(name)
	if err != nil {
		return nil, err
	}
	if project != nil {
		return project, nil
	}
	var created Project
	if err := c.postJSON("/projects/", map[string]any{"name": name}, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

func (c *Client) FindProjectByName(name string) (*Project, error) {
	var projects []Project
	if err := c.getJSON("/projects/?name="+url.QueryEscape(name)+"&page_size=100", &projects); err != nil {
		return nil, err
	}
	for _, project := range projects {
		if project.Name == name {
			return &project, nil
		}
	}
	return nil, nil
}

func (c *Client) FindSecretScope(projectID string, name string) (*SecretScope, error) {
	var scopes []SecretScope
	path := "/secret-scopes/?project=" + url.QueryEscape(projectID) +
		"&name=" + url.QueryEscape(name) + "&page_size=100"
	if err := c.getJSON(path, &scopes); err != nil {
		return nil, err
	}
	for _, scope := range scopes {
		if scope.Project == projectID && scope.Name == name {
			return &scope, nil
		}
	}
	return nil, nil
}

func (c *Client) PushPackage(projectID string, tag string, source PackageSource) (*Package, error) {
	if source.FilePath != "" {
		return c.pushPackageFile(projectID, tag, source.FilePath)
	}
	payload := map[string]any{
		"project": projectID,
		"tag":     tag,
	}
	if source.SourceURL != "" {
		payload["source_url"] = source.SourceURL
	}
	if source.SourceRef != "" {
		payload["source_ref"] = source.SourceRef
	}
	if source.SourceToken != "" {
		payload["source_token"] = source.SourceToken
	}
	var pkg Package
	if err := c.postJSON("/packages/", payload, &pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (c *Client) CreateRollout(input RolloutInput) (*Rollout, error) {
	payload := map[string]any{
		"package":           input.PackageID,
		"download_base_dir": input.DownloadBaseDir,
		"cleanup":           input.Cleanup,
		"targeting":         input.Targeting,
	}
	if input.SecretScopeID != "" {
		payload["secret_scope"] = input.SecretScopeID
	}
	var rollout Rollout
	if err := c.postJSON("/rollouts/", payload, &rollout); err != nil {
		return nil, err
	}
	return &rollout, nil
}

func (c *Client) pushPackageFile(projectID string, tag string, filePath string) (*Package, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open package file: %w", err)
	}
	defer func() { _ = file.Close() }()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("project", projectID)
	if tag != "" {
		_ = writer.WriteField("tag", tag)
	}
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("create multipart file field: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("read package file: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart body: %w", err)
	}

	var pkg Package
	if err := c.do("POST", "/packages/", writer.FormDataContentType(), &body, &pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (c *Client) getJSON(path string, out any) error {
	return c.do("GET", path, "", nil, out)
}

func (c *Client) postJSON(path string, payload any, out any) error {
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return err
	}
	return c.do("POST", path, "application/json", &body, out)
}

func (c *Client) do(method string, path string, contentType string, body io.Reader, out any) error {
	req, err := http.NewRequest(method, c.publicURL(path), body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return apiError(method, path, resp.Status, data)
	}
	if out == nil {
		return nil
	}
	var wrapped envelope[json.RawMessage]
	if err := json.NewDecoder(resp.Body).Decode(&wrapped); err != nil {
		return err
	}
	if len(wrapped.Data) == 0 {
		return nil
	}
	return json.Unmarshal(wrapped.Data, out)
}

// apiError extracts the human-readable message from the platform's error
// envelope {"error": {"message": "...", "details": {...}}} and falls back to
// the raw body when the envelope is absent or unparseable.
func apiError(method, path, status string, body []byte) error {
	var env struct {
		Error *struct {
			Message string `json:"message"`
			Details any    `json:"details"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &env) == nil && env.Error != nil {
		msg := env.Error.Message
		if env.Error.Details != nil {
			if raw, err := json.Marshal(env.Error.Details); err == nil && string(raw) != "null" {
				msg += ": " + string(raw)
			}
		}
		if msg != "" {
			return fmt.Errorf("%s %s: %s", method, path, msg)
		}
	}
	return fmt.Errorf("%s %s failed: %s: %s", method, path, status, strings.TrimSpace(string(body)))
}

func (c *Client) publicURL(path string) string {
	base := strings.TrimRight(c.baseURL, "/")
	if !strings.HasSuffix(base, "/api/public/v1") {
		base += "/api/public/v1"
	}
	return base + path
}
