package server

import (
	"html/template"
	"net/http"
	"path/filepath"
	"wingspan-ops/internal/esi"
	"wingspan-ops/internal/routing"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

// functions provides custom functions to be used within the HTML templates.
var functions = template.FuncMap{
	"HTML": func(s string) template.HTML {
		return template.HTML(s)
	},
	"add": func(a, b int) int {
		return a + b
	},
}

// Server holds all the dependencies required for the web application.
type Server struct {
	templates    map[string]*template.Template
	wingspanURL  string
	feedbackURL  string
	esiClient    *esi.ESIClient
	graph        *routing.Graph
	oauthConfig  *oauth2.Config
	sessionStore *sessions.CookieStore
}

// New creates and initializes a new Server instance.
func New(
	wingspanAPIURL, feedbackURL string,
	esiClient *esi.ESIClient,
	graph *routing.Graph,
	oauthConfig *oauth2.Config,
	sessionStore *sessions.CookieStore,
) (*Server, error) {
	// Initialize the template cache.
	cache, err := newTemplateCache("./templates")
	if err != nil {
		return nil, err
	}

	// Create and return the Server instance with all dependencies.
	return &Server{
		templates:    cache,
		wingspanURL:  wingspanAPIURL,
		feedbackURL:  feedbackURL,
		esiClient:    esiClient,
		graph:        graph,
		oauthConfig:  oauthConfig,
		sessionStore: sessionStore,
	}, nil
}

// RegisterRoutes sets up all the HTTP routes for the application.
func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// --- Public Routes ---
	// These routes are accessible to everyone, logged in or not.

	// 1. Static file server for CSS, images, etc.
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// 2. Authentication handlers for the SSO login and callback flow.
	mux.HandleFunc("/login", s.loginPageHandler) // Changed
	mux.HandleFunc("/callback", s.callbackHandler)
	mux.HandleFunc("/logout", s.logoutHandler)

	// --- Protected Routes ---
	// These routes are wrapped in the authMiddleware. If a user is not
	// logged in, the middleware will redirect them to the login page.
	mux.Handle("/", s.authMiddleware(http.HandlerFunc(s.homeHandler)))
	mux.Handle("/short-circuit", s.authMiddleware(http.HandlerFunc(s.shortCircuitHandler)))
	mux.Handle("/lookup", s.authMiddleware(http.HandlerFunc(s.lookupHandler)))
	mux.Handle("/about", s.authMiddleware(http.HandlerFunc(s.aboutHandler)))

	return mux
}

// newTemplateCache parses all templates and stores them in a map for efficient rendering.
func newTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	// Find all "page" templates (e.g., index.html, about.html).
	pages, err := filepath.Glob(filepath.Join(dir, "*.html"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		// Skip files that are not meant to be rendered as standalone pages.
		if name == "layout.html" {
			continue
		}

		// For each page, create a new template set that includes the main layout.
		ts, err := template.New(name).Funcs(functions).ParseFiles(filepath.Join(dir, "layout.html"), page)
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}

	return cache, nil
}
