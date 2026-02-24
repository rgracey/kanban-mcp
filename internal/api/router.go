package api

import (
	"net/http"
	"strings"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rgracey/kanban-mcp/frontend"
	"github.com/rgracey/kanban-mcp/internal/store"
)

// NewRouter creates a new HTTP handler with all API routes mounted under /api/v1.
// If mcpHandler is non-nil it is mounted at /mcp (Streamable HTTP MCP transport).
func NewRouter(s store.Store, hub *Hub, mcpHandler http.Handler) http.Handler {
	r := chi.NewRouter()

	// Standard middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Mount MCP handler at /mcp when HTTP transport is enabled.
	if mcpHandler != nil {
		r.Mount("/mcp", mcpHandler)
	}

	// Mount API routes under /api/v1
	api := chi.NewRouter()
	NewAPIRouter(api, s, hub)
	r.Mount("/api/v1", api)

	// Serve embedded SPA for all non-API routes.
	// http.FileServer handles Content-Type detection and ETag/Range support.
	// Sub the FS root to "dist" so paths like /assets/index.js resolve correctly.
	distFS, err := frontend.SubFS("dist")
	if err != nil {
		panic("frontend: failed to sub dist FS: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(distFS))

	r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
		// fs.FS paths must not have a leading slash.
		// If the file doesn't exist in the embedded FS, serve index.html so the
		// SPA router can handle the path client-side.
		path := strings.TrimPrefix(req.URL.Path, "/")
		if path != "" {
			if _, err := distFS.Open(path); err != nil {
				req.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, req)
	})

	return r
}

// NewAPIRouter mounts all API routes on the given chi router
func NewAPIRouter(r chi.Router, s store.Store, hub *Hub) {
	// SSE board-change stream
	r.Get("/events", SSEHandler(hub))
	// Boards
	r.Route("/boards", func(r chi.Router) {
		r.Get("/", ListBoards(s))
		r.Post("/", CreateBoard(s))
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", GetBoard(s))
			r.Put("/", UpdateBoard(s))
			r.Delete("/", DeleteBoard(s))
			r.Get("/summary", GetBoardSummary(s))
			r.Get("/context", GetBoardContext(s))
			r.Get("/ready", ReadyTickets(s))
			r.Get("/epics", ListEpics(s))
			r.Post("/epics", CreateEpic(s))
			r.Get("/tickets", ListTickets(s))
			r.Post("/tickets", CreateTicket(s, hub))
			r.Post("/tickets/bulk", BulkCreateTickets(s, hub))
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
			r.Put("/", UpdateTicket(s, hub))
			r.Delete("/", DeleteTicket(s, hub))
			r.Get("/notes", ListNotes(s))
			r.Post("/notes", CreateNote(s, hub))
			r.Get("/events", ListTicketEvents(s))
			r.Get("/tasks", ListTasks(s))
			r.Post("/tasks", CreateTask(s, hub))
			r.Post("/tasks/bulk", BulkCreateTasks(s, hub))
			r.Get("/relations", ListRelations(s))
			r.Post("/relations", AddRelation(s))
			r.Delete("/relations/{toId}", DeleteRelation(s))
		})
	})

	// Tasks (standalone routes for direct task access)
	r.Route("/tasks", func(r chi.Router) {
		r.Route("/{id}", func(r chi.Router) {
			r.Put("/", UpdateTask(s))
			r.Delete("/", DeleteTask(s))
		})
	})

	// Notes (standalone routes for direct note access)
	r.Route("/notes", func(r chi.Router) {
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", GetNote(s))
			r.Put("/", UpdateNote(s))
			r.Delete("/", DeleteNote(s))
		})
	})
}
