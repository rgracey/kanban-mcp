package models

import "time"

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
	ID          string    `json:"id"`
	BoardID     string    `json:"board_id"`
	EpicID      *string   `json:"epic_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	Priority    Priority  `json:"priority"`
	Assignee    string    `json:"assignee"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Comment struct {
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
