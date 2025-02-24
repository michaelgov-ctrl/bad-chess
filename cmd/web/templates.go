package main

import (
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"sort"

	"github.com/justinas/nosurf"
	"github.com/michaelgov-ctrl/bad-chess/ui"
)

type templateData struct {
	IsAuthenticated bool
	CSRFToken       string
	Flash           string
	Form            any
	TimeControls    []TimeControl
}

func (app *application) newTemplateData(r *http.Request) templateData {
	var td = templateData{
		IsAuthenticated: app.isAuthenticated(r),
		CSRFToken:       nosurf.Token(r),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		TimeControls:    []TimeControl{},
	}

	for k := range SupportedTimeControls {
		td.TimeControls = append(td.TimeControls, k)
	}
	sort.Slice(td.TimeControls, func(i, j int) bool {
		return td.TimeControls[i] < td.TimeControls[j]
	})

	return td
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
