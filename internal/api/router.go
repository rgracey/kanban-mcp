package api

import (
	"io"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rgracey/kanban-mcp/internal/frontend"
	"github.com/rgracey/kanban-mcp/internal/store"
)

// NewRouter creates a new HTTP handler with all API routes mounted under /api/v1
func NewRouter(s store.Store) http.Handler {
	r := chi.NewRouter()

	// Standard middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Mount API routes under /api/v1
	api := chi.NewRouter()
	NewAPIRouter(api, s)
	r.Mount("/api/v1", api)

	// Serve embedded SPA for all non-API routes
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		var file string
		if r.URL.Path == "/" || r.URL.Path == "" {
			file = "dist/index.html"
		} else {
			// Try to serve the requested file
			_, err := frontend.FS.ReadFile("dist" + r.URL.Path)
			if err != nil {
				// File not found, serve index.html for SPA client-side routing
				file = "dist/index.html"
			} else {
				file = "dist" + r.URL.Path
			}
		}

		f, err := frontend.FS.Open(file)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		defer f.Close()

		// Copy file content to response
		io.Copy(w, f)
	})

	return r
}

// NewAPIRouter mounts all API routes on the given chi router
func NewAPIRouter(r chi.Router, s store.Store) {
	// Boards
	r.Route("/boards", func(r chi.Router) {
		r.Get("/", ListBoards(s))
		r.Post("/", CreateBoard(s))
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", GetBoard(s))
			r.Put("/", UpdateBoard(s))
			r.Delete("/", DeleteBoard(s))
			r.Get("/summary", GetBoardSummary(s))
			r.Get("/epics", ListEpics(s))
			r.Post("/epics", CreateEpic(s))
			r.Get("/tickets", ListTickets(s))
			r.Post("/tickets", CreateTicket(s))
		})
	})

	// Epics (standalone routes for direct epic access)
	r.Route("/epics", func(r chi.Router) {
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", GetEpic(s))
			r.Put("/", UpdateEpic(s))
			r.Delete("/", DeleteEpic(s))
		})
	})

	// Tickets (standalone routes for direct ticket access)
	r.Route("/tickets", func(r chi.Router) {
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", GetTicket(s))
			r.Put("/", UpdateTicket(s))
			r.Delete("/", DeleteTicket(s))
			r.Get("/comments", ListComments(s))
			r.Post("/comments", CreateComment(s))
		})
	})

	// Comments (standalone routes for direct comment access)
	r.Route("/comments", func(r chi.Router) {
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", GetComment(s))
			r.Put("/", UpdateComment(s))
			r.Delete("/", DeleteComment(s))
		})
	})
}
