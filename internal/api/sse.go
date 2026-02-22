package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// SSEEvent is a board-change notification sent over the SSE stream.
type SSEEvent struct {
	Type    string `json:"type"`     // e.g. "ticket.created", "ticket.updated", "task.updated"
	BoardID string `json:"board_id"` // board that changed
}

// Hub manages SSE client subscriptions and broadcasts events.
type Hub struct {
	mu      sync.RWMutex
	clients map[chan SSEEvent]struct{}
	done    chan struct{}
}

// NewHub creates a new broadcast hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[chan SSEEvent]struct{}),
		done:    make(chan struct{}),
	}
}

// Close signals all active SSE handlers to exit, allowing a clean shutdown.
func (h *Hub) Close() {
	close(h.done)
}

func (h *Hub) subscribe() chan SSEEvent {
	ch := make(chan SSEEvent, 16)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) unsubscribe(ch chan SSEEvent) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
	close(ch)
}

// Publish broadcasts an event to all connected SSE clients.
func (h *Hub) Publish(ev SSEEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- ev:
		default:
			// slow client — skip rather than block
		}
	}
}

// SSEHandler returns an http.HandlerFunc that streams events to the client.
func SSEHandler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		// Send an initial ping so the client knows the connection is live.
		fmt.Fprintf(w, "event: ping\ndata: {}\n\n")
		flusher.Flush()

		ch := hub.subscribe()
		defer hub.unsubscribe(ch)

		for {
			select {
			case <-hub.done:
				return
			case <-r.Context().Done():
				return
			case ev, ok := <-ch:
				if !ok {
					return
				}
				data, _ := json.Marshal(ev)
				fmt.Fprintf(w, "event: board_change\ndata: %s\n\n", data)
				flusher.Flush()
			}
		}
	}
}
