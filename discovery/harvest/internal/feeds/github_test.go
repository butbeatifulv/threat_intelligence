package feeds

import (
	"archive/zip"
	"bytes"
	"errors"
	"net/http"
	"testing"
)

func TestGitHubRawURL(t *testing.T) {
	got := GitHubRawURL("o", "r", "main", "/path/file.yml")
	want := "https://raw.githubusercontent.com/o/r/main/path/file.yml"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestGitHubRepoKey(t *testing.T) {
	if gitHubRepoKey("a", "b") != "a/b" {
		t.Fatal()
	}
}

func TestIsGitHubListFallback(t *testing.T) {
	if isGitHubListFallback(nil) {
		t.Fatal("nil err")
	}
	if !isGitHubListFallback(&HTTPStatusError{Code: http.StatusForbidden}) {
		t.Fatal("403")
	}
	if !isGitHubListFallback(errors.New("GET failed: 429 Too Many Requests")) {
		t.Fatal("429 string")
	}
	if isGitHubListFallback(errors.New("connection reset")) {
		t.Fatal("unrelated")
	}
}

func TestZipPathsUnderPrefix(t *testing.T) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	w, err := zw.Create("repo-main/foo/bar.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}
	if _, err := zw.Create("repo-main/other/out.txt"); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	paths := zipPathsUnderPrefix(zr, "foo")
	if len(paths) != 1 || paths[0] != "foo/bar.txt" {
		t.Fatalf("got %v", paths)
	}
}
