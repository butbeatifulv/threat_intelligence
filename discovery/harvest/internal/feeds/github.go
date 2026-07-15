package feeds

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultGitHubUserAgent = "veil-scrape/1.0"
	maxCodeloadZipBytes    = 220 << 20 // 220 MiB (atomic-red-team ~161 MiB)
)

// gitHubCodeloadSkip — repos too large for zip download in scrape/smoke (use API or skip).
var gitHubCodeloadSkip = map[string]bool{
	"rapid7/metasploit-framework": true,
}

func gitHubRepoKey(owner, repo string) string {
	return owner + "/" + repo
}

func gitHubRefs() []string {
	return []string{"main", "master"}
}

// GitHubRefs returns branch names tried for codeload/raw GitHub fetches (main first).
func GitHubRefs() []string {
	return gitHubRefs()
}

// GHContent is a file or directory entry under a GitHub repo path.
type GHContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

// GitHubRawURL builds a raw.githubusercontent.com URL (no API token required).
func GitHubRawURL(owner, repo, ref, path string) string {
	path = strings.TrimPrefix(path, "/")
	ref = strings.TrimSpace(ref)
	if ref == "" {
		ref = "master"
	}
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, ref, path)
}

func gitHubApplyHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", defaultGitHubUserAgent)
	if tok := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
}

func isGitHubListFallback(err error) bool {
	if err == nil {
		return false
	}
	var he *HTTPStatusError
	if ok := errorAsHTTPStatus(err, &he); ok {
		return he.Code == http.StatusForbidden || he.Code == http.StatusTooManyRequests ||
			he.Code == http.StatusUnauthorized
	}
	s := err.Error()
	return strings.Contains(s, " 403 ") || strings.Contains(s, " 429 ") ||
		strings.Contains(s, "Forbidden") || strings.Contains(s, "rate limit")
}

func errorAsHTTPStatus(err error, he **HTTPStatusError) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*HTTPStatusError); ok {
		*he = e
		return true
	}
	return false
}

// GitHubListDir lists immediate children of path in a public repo (no token required).
// Prefers codeload.github.com (no REST quota); falls back to API when zip unavailable.
func GitHubListDir(ctx context.Context, c *Client, owner, repo, path string) ([]GHContent, error) {
	path = strings.Trim(path, "/")
	if !gitHubCodeloadSkip[gitHubRepoKey(owner, repo)] {
		var codeloadErr error
		for _, ref := range gitHubRefs() {
			items, err := gitHubListDirFromCodeload(ctx, c, owner, repo, ref, path)
			if err == nil {
				return items, nil
			}
			if codeloadErr == nil || strings.Contains(err.Error(), "exceeds") {
				codeloadErr = err
			}
		}
		// API only when codeload failed for a reason other than size/network
		if codeloadErr != nil && !strings.Contains(codeloadErr.Error(), "exceeds") {
			for _, ref := range gitHubRefs() {
				items, err := gitHubListTreeChildren(ctx, c, owner, repo, ref, path)
				if err == nil {
					return items, nil
				}
				if !isGitHubListFallback(err) {
					return nil, codeloadErr
				}
			}
			items, err := gitHubListContents(ctx, c, owner, repo, path)
			if err == nil {
				return items, nil
			}
		}
		return nil, codeloadErr
	}
	for _, ref := range gitHubRefs() {
		items, err := gitHubListTreeChildren(ctx, c, owner, repo, ref, path)
		if err == nil {
			return items, nil
		}
		if !isGitHubListFallback(err) {
			return nil, err
		}
	}
	return gitHubListContents(ctx, c, owner, repo, path)
}

func gitHubListContents(ctx context.Context, c *Client, owner, repo, path string) ([]GHContent, error) {
	path = strings.Trim(path, "/")
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	gitHubApplyHeaders(req)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("github list %s: %s %s", path, resp.Status, string(b))
	}
	var out []GHContent
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

type gitTreeEntry struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
}

type gitTreeResponse struct {
	Tree      []gitTreeEntry `json:"tree"`
	Truncated bool           `json:"truncated"`
}

// gitHubListTreeChildren returns direct children under pathPrefix using recursive tree (one API call).
func gitHubListTreeChildren(ctx context.Context, c *Client, owner, repo, ref, pathPrefix string) ([]GHContent, error) {
	pathPrefix = strings.Trim(pathPrefix, "/")
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", owner, repo, ref)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	gitHubApplyHeaders(req)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("github tree %s@%s: %s %s", repo, ref, resp.Status, string(b))
	}
	var tr gitTreeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}
	if tr.Truncated {
		return nil, fmt.Errorf("github tree %s@%s truncated", repo, ref)
	}

	prefix := pathPrefix
	if prefix != "" {
		prefix += "/"
	}
	seen := make(map[string]GHContent)
	for _, ent := range tr.Tree {
		if ent.Type != "blob" && ent.Type != "tree" {
			continue
		}
		if prefix != "" && !strings.HasPrefix(ent.Path, prefix) {
			continue
		}
		rel := strings.TrimPrefix(ent.Path, prefix)
		if rel == "" || strings.Contains(rel, "/") {
			continue
		}
		typ := "file"
		if ent.Type == "tree" || ent.Mode == "040000" {
			typ = "dir"
		}
		name := rel
		fullPath := ent.Path
		if pathPrefix != "" && !strings.HasPrefix(fullPath, pathPrefix) {
			fullPath = pathPrefix + "/" + rel
		}
		seen[rel] = GHContent{
			Name:        name,
			Path:        fullPath,
			Type:        typ,
			DownloadURL: GitHubRawURL(owner, repo, ref, fullPath),
		}
	}
	out := make([]GHContent, 0, len(seen))
	for _, v := range seen {
		out = append(out, v)
	}
	return out, nil
}

// GitHubListTreePaths returns repo file paths under pathPrefix (blobs only).
func GitHubListTreePaths(ctx context.Context, c *Client, owner, repo, ref, pathPrefix string) ([]string, error) {
	if ref == "" {
		ref = "master"
	}
	pathPrefix = strings.Trim(pathPrefix, "/")
	if !gitHubCodeloadSkip[gitHubRepoKey(owner, repo)] {
		for _, r := range gitHubRefs() {
			out, err := gitHubListTreePathsFromCodeload(ctx, c, owner, repo, r, pathPrefix)
			if err == nil {
				return out, nil
			}
		}
	}
	out, err := gitHubListTreePathsAPI(ctx, c, owner, repo, ref, pathPrefix)
	if err == nil {
		return out, nil
	}
	if !isGitHubListFallback(err) {
		return nil, err
	}
	for _, r := range gitHubRefs() {
		if r == ref {
			continue
		}
		out, err2 := gitHubListTreePathsAPI(ctx, c, owner, repo, r, pathPrefix)
		if err2 == nil {
			return out, nil
		}
	}
	return nil, err
}

func gitHubListTreePathsAPI(ctx context.Context, c *Client, owner, repo, ref, pathPrefix string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", owner, repo, ref)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	gitHubApplyHeaders(req)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("github tree paths: %s", string(b))
	}
	var tr gitTreeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}
	prefix := pathPrefix
	if prefix != "" {
		prefix += "/"
	}
	var out []string
	for _, ent := range tr.Tree {
		if ent.Type != "blob" {
			continue
		}
		if prefix != "" && !strings.HasPrefix(ent.Path, prefix) {
			continue
		}
		out = append(out, ent.Path)
	}
	return out, nil
}

func gitHubDownloadCodeloadZip(ctx context.Context, c *Client, owner, repo, ref string) ([]byte, error) {
	cacheRel := filepath.Join("codeload", owner, repo, ref+".zip")
	if b, ok := c.ReadCache(cacheRel); ok {
		return b, nil
	}
	zipURL := fmt.Sprintf("https://codeload.github.com/%s/%s/zip/refs/heads/%s", owner, repo, ref)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, zipURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", defaultGitHubUserAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("codeload %s/%s@%s: %s %s", owner, repo, ref, resp.Status, string(b))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxCodeloadZipBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxCodeloadZipBytes {
		return nil, fmt.Errorf("codeload zip %s/%s@%s exceeds %d bytes", owner, repo, ref, maxCodeloadZipBytes)
	}
	_ = c.WriteCache(cacheRel, body)
	return body, nil
}

func zipPathsUnderPrefix(zr *zip.Reader, pathPrefix string) []string {
	pathPrefix = strings.Trim(pathPrefix, "/")
	var pfx string
	if pathPrefix != "" {
		pfx = pathPrefix + "/"
	}
	var out []string
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := strings.ReplaceAll(f.Name, "\\", "/")
		if strings.HasPrefix(name, "/") || strings.Contains(name, "../") || name == ".." {
			// Zip Slip: codeload archives are fetched from GitHub over HTTPS, but a
			// compromised/malicious repo could still ship a crafted entry name; never
			// trust it enough to let it leave this function unsanitized.
			continue
		}
		slash := strings.Index(name, "/")
		if slash < 0 {
			continue
		}
		rel := name[slash+1:]
		if strings.Contains(rel, "../") || strings.HasPrefix(rel, "/") {
			continue
		}
		if pfx != "" && !strings.HasPrefix(rel, pfx) {
			continue
		}
		out = append(out, rel)
	}
	return out
}

func gitHubListDirFromCodeload(ctx context.Context, c *Client, owner, repo, ref, pathPrefix string) ([]GHContent, error) {
	body, err := gitHubDownloadCodeloadZip(ctx, c, owner, repo, ref)
	if err != nil {
		return nil, err
	}
	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, err
	}
	pathPrefix = strings.Trim(pathPrefix, "/")
	var pfx string
	if pathPrefix != "" {
		pfx = pathPrefix + "/"
	}
	seen := make(map[string]GHContent)
	for _, rel := range zipPathsUnderPrefix(zr, pathPrefix) {
		rest := rel
		if pfx != "" {
			if !strings.HasPrefix(rest, pfx) {
				continue
			}
			rest = strings.TrimPrefix(rest, pfx)
		}
		if rest == "" {
			continue
		}
		parts := strings.SplitN(rest, "/", 2)
		child := parts[0]
		if child == "" {
			continue
		}
		if _, ok := seen[child]; ok {
			continue
		}
		fullPath := child
		if pathPrefix != "" {
			fullPath = pathPrefix + "/" + child
		}
		typ := "file"
		if len(parts) > 1 {
			typ = "dir"
		}
		seen[child] = GHContent{
			Name:        child,
			Path:        fullPath,
			Type:        typ,
			DownloadURL: GitHubRawURL(owner, repo, ref, fullPath),
		}
	}
	out := make([]GHContent, 0, len(seen))
	for _, v := range seen {
		out = append(out, v)
	}
	return out, nil
}

func gitHubListTreePathsFromCodeload(ctx context.Context, c *Client, owner, repo, ref, pathPrefix string) ([]string, error) {
	body, err := gitHubDownloadCodeloadZip(ctx, c, owner, repo, ref)
	if err != nil {
		return nil, err
	}
	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, err
	}
	return zipPathsUnderPrefix(zr, pathPrefix), nil
}

// GitHubFetchRaw downloads a file via raw.githubusercontent.com (open, no token).
func GitHubFetchRaw(ctx context.Context, c *Client, owner, repo, ref, path string) ([]byte, error) {
	rawURL := GitHubRawURL(owner, repo, ref, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", defaultGitHubUserAgent)
	return c.DoGET(req, "")
}
