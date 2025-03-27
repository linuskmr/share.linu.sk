package main

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

// ServeDir lists the directory entries.
func ServeDir(w http.ResponseWriter, r *http.Request, osPath string) {
	if r.URL.Path == "" || r.URL.Path == "." || r.URL.Path == "/" {
		http.Error(w, "No directory listing for base directory", http.StatusForbidden)
	}

	osEntries, err := os.ReadDir(osPath)
	if err != nil {
		// The err could be a not found error (i.e. `os.IsNotExist(err) == true`),
		// but this function should not have been called in that case, so we do not handle it separately.
		slog.Warn("Reading directory", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var dirEntries []DirEntry
	for _, entry := range osEntries {
		dirEntry := DirEntry{
			Name:  entry.Name(),
			Path:  filepath.Join(r.URL.Path, entry.Name()),
			IsDir: entry.IsDir(),
		}
		dirEntries = append(dirEntries, dirEntry)
	}

	w.Header().Set("Content-Type", "text/html")
	err = templates.ExecuteTemplate(w, "directory.html", struct {
		DirectoryName string
		Entries       []DirEntry
		BreadcrumbNav []DirEntry
	}{
		DirectoryName: filepath.Base(osPath),
		Entries:       dirEntries,
		BreadcrumbNav: BreadcrumbNavigation(r.URL.Path),
	})
	if err != nil {
		slog.Warn("Executing markdown template", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
