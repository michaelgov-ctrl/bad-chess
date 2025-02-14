package main

import (
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/michaelgov-ctrl/badchess/ui"
)

type templateData struct {
	IsAuthenticated bool
}

func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		IsAuthenticated: false,
	}
}

func newTemplateCache() (map[string]*template.Template, error) {
	var cache = make(map[string]*template.Template)

	pages, err := fs.Glob(ui.Files, "html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/base.tmpl.html",
			"html/partials/*.html",
			page,
		}

		ts, err := template.New(name).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
