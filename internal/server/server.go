package server

import (
	"html/template"
	"net/http"
	"path/filepath"
	"wingspan-ops/internal/esi"
	"wingspan-ops/internal/routing"
)

var functions = template.FuncMap{
	"HTML": func(s string) template.HTML {
		return template.HTML(s)
	},
	"add": func(a, b int) int {
		return a + b
	},
}

// The Server struct now holds a map of templates.
type Server struct {
	templates   map[string]*template.Template
	wingspanURL string
	feedbackURL string
	esiClient   *esi.ESIClient
	graph       *routing.Graph
}

func New(wingspanAPIURL, feedbackURL string, esiClient *esi.ESIClient, graph *routing.Graph) (*Server, error) {
	// The new cache function returns a map.
	cache, err := newTemplateCache("./templates")
	if err != nil {
		return nil, err
	}

	return &Server{
		templates:   cache,
		wingspanURL: wingspanAPIURL,
		feedbackURL: feedbackURL,
		esiClient:   esiClient,
		graph:       graph,
	}, nil
}

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.homeHandler)
	mux.HandleFunc("/short-circuit", s.shortCircuitHandler)
	mux.HandleFunc("/lookup", s.lookupHandler)
	mux.HandleFunc("/about", s.aboutHandler) // ADD THIS
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
	return mux
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	// Find all "page" templates (anything that isn't a layout/partial).
	pages, err := filepath.Glob(filepath.Join(dir, "*.html"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		// Skip files that are not "pages" (like layouts).
		if name == "layout.html" {
			continue
		}

		// Create a new template set for each page.
		// Start with the base layout, then add the specific page file.
		ts, err := template.New(name).Funcs(functions).ParseFiles(filepath.Join(dir, "layout.html"), page)
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}
	return cache, nil
}
