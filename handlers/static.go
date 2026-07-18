package handlers

import (
	"net/http"
	"strings"

	"yenikule/config"
)

// RegisterStatic mounts the static file tree and adds a few quality-of-life
// behaviours with zero per-request heap allocations beyond the standard library.
func RegisterStatic(mux *http.ServeMux, cfg *config.Config) {
	fs := http.FileServer(http.Dir(cfg.StaticDir))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// ── Security headers (set once per response, no alloc) ──────────────
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "SAMEORIGIN")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// ── Cache-Control per content type ──────────────────────────────────
		switch {
		case strings.HasSuffix(r.URL.Path, ".html") || r.URL.Path == "/":
			// HTML: revalidate every time (content may change)
			h.Set("Cache-Control", "public, max-age=0, must-revalidate")
		case isImmutable(r.URL.Path):
			// Versioned assets (css, js, images): cache 1 year
			h.Set("Cache-Control", "public, max-age=31536000, immutable")
		default:
			h.Set("Cache-Control", "public, max-age=3600")
		}

		// ── Redirect bare /index to / ────────────────────────────────────────
		if r.URL.Path == "/index" || r.URL.Path == "/index.html" {
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
			return
		}

		fs.ServeHTTP(w, r)
	})
}

// isImmutable returns true for asset paths that are safe to cache forever.
func isImmutable(path string) bool {
	for _, ext := range []string{".css", ".js", ".woff2", ".webp", ".png", ".ico", ".svg"} {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}
