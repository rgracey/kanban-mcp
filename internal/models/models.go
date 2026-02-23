package models

import "time"

// TicketResolution is written by an agent when it closes a ticket.
type TicketResolution struct {
	CommitSHA  string `json:"commit_sha,omitempty"`
	PRURL      string `json:"pr_url,omitempty"`
	Notes      string `json:"notes,omitempty"`
	ResolvedAt string `json:"resolved_at,omitempty"` // RFC3339
}

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

type Board struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Epic struct {
	ID          string    `json:"id"`
	BoardID     string    `json:"board_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Ticket struct {
	ID          string            `json:"id"`
	BoardID     string            `json:"board_id"`
	EpicID      *string           `json:"epic_id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      Status            `json:"status"`
	Priority    Priority          `json:"priority"`
	Assignee    string            `json:"assignee"`
	Resolution  *TicketResolution `json:"resolution,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Note is an agent scratchpad entry attached to a ticket.
// Notes replace the old comments system and are intended for machine-readable
// observations, investigation logs, and intermediate reasoning.
type Note struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticket_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BoardSummary struct {
	BoardID      string         `json:"board_id"`
	TicketCounts map[string]int `json:"ticket_counts"`
	Epics        []EpicSummary  `json:"epics"`
}

type EpicSummary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	TicketCount int    `json:"ticket_count"`
}

// TicketContext embeds a Ticket with its associated tasks and relations,
// used in the board context snapshot.
type TicketContext struct {
	Ticket
	Tasks     []Task           `json:"tasks"`
	Relations []TicketRelation `json:"relations"`
}

// BoardContext is a complete machine-readable snapshot of a board in one shot.
// It is designed for LLM agents that need full context without multiple round-trips.
type BoardContext struct {
	Board   Board           `json:"board"`
	Epics   []Epic          `json:"epics"`
	Tickets []TicketContext `json:"tickets"`
}

type TicketFilter struct {
	Status    *Status
	Priority  *Priority
	EpicID    *string
	Query     *string // keyword search against title + description
	SortBy    *string // "priority" | "created_at" (default)
	SortOrder *string // "asc" | "desc" (default)
}

type Task struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticket_id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RelationKind string

const (
	RelationBlocks RelationKind = "blocks"
)

// TicketRelation represents a directional relation between two tickets.
// from_ticket_id blocks to_ticket_id.
type TicketRelation struct {
	FromTicketID string       `json:"from_ticket_id"`
	ToTicketID   string       `json:"to_ticket_id"`
	Kind         RelationKind `json:"kind"`
	CreatedAt    time.Time    `json:"created_at"`
	// Denormalised title fields populated on list queries
	FromTitle string `json:"from_title,omitempty"`
	ToTitle   string `json:"to_title,omitempty"`
}

type TicketEventType string

const (
	EventCreated         TicketEventType = "created"
	EventMoved           TicketEventType = "moved"
	EventEdited          TicketEventType = "edited"
	EventCommented       TicketEventType = "commented"
	EventCommentEdited   TicketEventType = "comment_edited"
	EventTaskAdded       TicketEventType = "task_added"
	EventTaskUpdated     TicketEventType = "task_updated"
	EventTaskDeleted     TicketEventType = "task_deleted"
	EventRelationAdded   TicketEventType = "relation_added"
	EventRelationRemoved TicketEventType = "relation_removed"
)

type TicketEvent struct {
	ID        string          `json:"id"`
	TicketID  string          `json:"ticket_id"`
	Type      TicketEventType `json:"type"`
	Actor     string          `json:"actor"`
	Payload   map[string]any  `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}
