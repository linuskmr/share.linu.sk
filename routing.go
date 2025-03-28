package main

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var firstRequestTimestamp *time.Time = nil

func Handler(w http.ResponseWriter, r *http.Request) {
	// To maintain a flat directory structure (i.e. that the http paths directly map to paths inside the served directory),
	// we need a trick to serve static files, which could collide with the paths of the served files.
	if r.URL.Query().Get("staticDirBypass") == "true" {
		// The normal `http.ServeFile()` would set the `Last-Modified` header, but we are using the embed filesystem, which has no metadata information of files (see https://github.com/golang/go/issues/44854), so we have to set it manually to the program startup time.
		if firstRequestTimestamp == nil {
			firstRequestTimestamp = new(time.Time)
			*firstRequestTimestamp = time.Now()
		}
		w.Header().Set("Last-Modified", firstRequestTimestamp.UTC().Format(http.TimeFormat))
		w.Header().Set("Cache-Control", "max-age=0, must-revalidate")
		http.FileServerFS(staticFs).ServeHTTP(w, r)
		return
	}

	if r.URL.Path == "/" {
		http.ServeFileFS(w, r, staticFs, "static/index.html")
		return
	}

	// Convert request URL path to OS path
	osPath := r.URL.Path
	osPath, _ = strings.CutPrefix(osPath, "/")
	osPath = filepath.Clean(osPath)
	if strings.Contains(osPath, "..") || strings.HasPrefix(osPath, "/") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	osPath = filepath.Join(baseFileDir, osPath)

	stat, err := os.Stat(osPath)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Debug("File not found", "path", osPath)
			NotFoundHandler(w, r)
			return
		}
		slog.Warn("Getting file info", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if stat.IsDir() {
		ServeDir(w, r, osPath)
	} else {
		ServeFile(w, r, osPath)
	}
}
