package main

import (
	"path"
	"strings"
)

// BreadcrumbNavigation splits path_ (i.e. `r.URL.Path`) into its components
func BreadcrumbNavigation(path_ string) []DirEntry {
	if path_ == "." {
		path_ = ""
	}

	filenames := strings.Split(path_, "/")

	dirEntries := make([]DirEntry, 0, len(filenames))

	currentPath := "/"
	for _, filename := range filenames {
		if filename == "" {
			continue
		}
		currentPath = path.Join(currentPath, filename)
		dirEntries = append(dirEntries, DirEntry{
			Name:  filename,
			Path:  currentPath,
			IsDir: true, // Not known, but not important for the breadcrumb navigation
		})
	}

	return dirEntries
}
