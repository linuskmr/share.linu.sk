package main

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"
)

// ServeFile serves a markdown file.
func ServeFile(w http.ResponseWriter, r *http.Request, osPath string) {
	stat, err := os.Stat(osPath)
	if err != nil {
		slog.Warn("Reading markdown file", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("If-Modified-Since") == stat.ModTime().UTC().Format(http.TimeFormat) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	markdownFile, err := os.ReadFile(osPath)
	if err != nil {
		// The err could be a not found error (i.e. `os.IsNotExist(err) == true`),
		// but this function should not have been called in that case, so we do not handle it separately.
		slog.Warn("Reading markdown file", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	markdownRenderer := goldmark.New(
		goldmark.WithExtensions(extension.GFM, &frontmatter.Extender{Mode: frontmatter.SetMetadata}),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	var htmlBuf bytes.Buffer
	err = markdownRenderer.Convert(markdownFile, &htmlBuf)
	if err != nil {
		slog.Warn("Converting markdown file to html", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	markdownDocumentRoot := markdownRenderer.Parser().Parse(text.NewReader(markdownFile))
	markdownDocument := markdownDocumentRoot.OwnerDocument()
	markdownMetadata := markdownDocument.Meta()
	filenameBase := strings.TrimSuffix(filepath.Base(osPath), filepath.Ext(osPath))

	// Check if the markdown file starts with a heading that contains the filename base.
	// If not, the filename base will be printed as the title in the template.
	firstHeadingRegex := regexp.MustCompile(`\<h1 id=".*"\>(.*)\<\/h1\>`)
	firstHeadingMatch := firstHeadingRegex.FindSubmatch(htmlBuf.Bytes())
	firstHeadingContainsFilenameBase := false
	if firstHeadingMatch != nil {
		firstHeading := strings.ToLower(string(firstHeadingMatch[1]))
		metadataHeading := strings.ToLower(filenameBase)
		firstHeadingContainsFilenameBase = strings.Contains(firstHeading, metadataHeading)
	}

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Last-Modified", stat.ModTime().UTC().Format(http.TimeFormat))
	w.Header().Set("Cache-Control", "max-age=0, must-revalidate")

	err = templates.ExecuteTemplate(w, "markdown.html", struct {
		FilenameBase  string
		PrintTitle    bool
		Meta          map[string]interface{}
		Body          template.HTML
		BreadcrumbNav []DirEntry
	}{
		FilenameBase:  filenameBase,
		PrintTitle:    !firstHeadingContainsFilenameBase,
		Meta:          markdownMetadata,
		Body:          template.HTML(htmlBuf.String()),
		BreadcrumbNav: BreadcrumbNavigation(r.URL.Path),
	})
	if err != nil {
		slog.Warn("Executing markdown template", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
