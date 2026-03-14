package main

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"log/slog"
)

//go:embed static
var embeddedFrontend embed.FS

func registerEmbeddedFrontendRoutes(r *gin.Engine) {
	distFS, err := fs.Sub(embeddedFrontend, "static")
	if err != nil {
		slog.Error("не вдалося отримати директорію static", "помилка", err)
		panic(err)
	}

	indexHTML, err := fs.ReadFile(distFS, "index.html")
	if err != nil {
		slog.Error("не вдалося прочитати index.html", "помилка", err)
		panic(err)
	}

	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Ресурс не знайдено"})
			return
		}

		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Status(http.StatusMethodNotAllowed)
			return
		}

		if serveEmbeddedFile(c, distFS) {
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})
}

func serveEmbeddedFile(c *gin.Context, distFS fs.FS) bool {
	path := strings.TrimPrefix(c.Request.URL.Path, "/")
	if path == "" {
		return false
	}

	if stat, err := fs.Stat(distFS, path); err == nil && !stat.IsDir() {
		c.FileFromFS(path, http.FS(distFS))
		return true
	}

	return false
}
