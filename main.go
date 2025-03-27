package main

import (
	"embed"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
)

//go:embed templates/*
var templatesFs embed.FS

//go:embed static/*
var staticFs embed.FS

var templates = template.Must(template.ParseFS(templatesFs, "templates/*.html"))

// Directory to serve files from
var baseFileDir string

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	addr := flag.String("addr", ":8080", "Network address to listen on")
	flag.StringVar(&baseFileDir, "dir", ".", "Directory to serve files from")
	flag.Parse()

	http.HandleFunc("GET /", Handler)

	slog.Info("Starting server", "addr", *addr)
	http.ListenAndServe(*addr, nil)
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFileFS(w, r, staticFs, "static/404.html")
}

type DirEntry struct {
	Name  string
	Path  string
	IsDir bool
}
