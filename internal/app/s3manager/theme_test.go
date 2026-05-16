package s3manager_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudlena/s3manager/internal/app/s3manager"
	"github.com/cloudlena/s3manager/internal/app/s3manager/mocks"
	"github.com/matryer/is"
	"github.com/minio/minio-go/v7"
)

// TestDarkModeWiring guards the pieces of dark-mode plumbing that are easy to
// accidentally delete: the inline FOUC <script>, the dark.css <link>, the
// toggle button, and theme.js. If any of these are removed, dark mode silently
// regresses without anything else failing.
func TestDarkModeWiring(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	s3 := &mocks.S3Mock{
		ListBucketsFunc: func(context.Context) ([]minio.BucketInfo, error) {
			return []minio.BucketInfo{}, nil
		},
	}
	templates := os.DirFS(filepath.Join("..", "..", "..", "web", "template"))

	req, err := http.NewRequest(http.MethodGet, "/buckets", nil)
	is.NoErr(err)

	rr := httptest.NewRecorder()
	handler := s3manager.HandleBucketsView(s3, templates, true, "", "")
	handler.ServeHTTP(rr, req)
	resp := rr.Result()
	defer func() {
		err = resp.Body.Close()
		is.NoErr(err)
	}()
	bodyBytes, err := io.ReadAll(resp.Body)
	is.NoErr(err)
	body := string(bodyBytes)

	is.Equal(http.StatusOK, resp.StatusCode)

	// Inline FOUC script applies the theme before any stylesheet renders, so
	// the user never sees a light flash before dark loads.
	is.True(strings.Contains(body, "data-theme"))                  // FOUC script sets data-theme attr
	is.True(strings.Contains(body, "prefers-color-scheme"))        // ...reading system preference
	is.True(strings.Contains(body, "localStorage.getItem('theme')")) // ...with manual-toggle override

	// Stylesheet + JS that implement the dark theme.
	is.True(strings.Contains(body, "/static/css/dark.css"))
	is.True(strings.Contains(body, "/static/js/theme.js"))

	// The toggle button itself.
	is.True(strings.Contains(body, "toggleTheme()"))
	is.True(strings.Contains(body, `aria-label="Toggle dark mode"`))
}

// TestDarkCSSKeyTokens ensures the dark.css palette tokens that the rest of
// the file references are actually defined. Catches "I deleted --danger-soft
// by mistake" without needing a full headless-browser run.
func TestDarkCSSKeyTokens(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	path := filepath.Join("..", "..", "..", "web", "static", "css", "dark.css")
	contents, err := os.ReadFile(path)
	is.NoErr(err)

	css := string(contents)
	for _, token := range []string{
		"--bg:", "--surface:", "--text:", "--muted:", "--border:",
		"--accent:", "--accent-hover:", "--accent-soft:", "--accent-active-text:",
		"--hover-bg:", "--teal:", "--teal-hover:", "--danger:", "--danger-hover:",
		"--danger-soft:", "--success-soft:", "--warning-soft:",
	} {
		is.True(strings.Contains(css, token)) // token must be defined
	}

	// The dark theme must scope its overrides under [data-theme="dark"] so
	// light mode renders unaffected.
	is.True(strings.Contains(css, `:root[data-theme="dark"]`))

	// Print reset: prevents the dark palette from leaking onto paper.
	is.True(strings.Contains(css, "@media print"))
}
