package mcp

import (
	"context"
	"encoding/json"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rgracey/kanban-mcp/internal/models"
	"github.com/rgracey/kanban-mcp/internal/store"
)

func registerTools(srv *server.MCPServer, s store.Store) {
	// --- Boards ---

	srv.AddTool(
		mcpgo.NewTool("list_boards",
			mcpgo.WithDescription("List all kanban boards"),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			boards, err := s.ListBoards(ctx)
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return jsonResult(boards)
		},
	)

	srv.AddTool(
		mcpgo.NewTool("create_board",
			mcpgo.WithDescription("Create a new kanban board"),
			mcpgo.WithString("name", mcpgo.Required(), mcpgo.Description("Board name")),
			mcpgo.WithString("description", mcpgo.Description("Board description")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			name, err := req.RequireString("name")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			description := req.GetString("description", "")
			board, err := s.CreateBoard(ctx, name, description)
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return jsonResult(board)
		},
	)

	srv.AddTool(
		mcpgo.NewTool("update_board",
			mcpgo.WithDescription("Update an existing board"),
			mcpgo.WithString("id", mcpgo.Required(), mcpgo.Description("Board ID")),
			mcpgo.WithString("name", mcpgo.Description("New board name")),
			mcpgo.WithString("description", mcpgo.Description("New board description")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			id, err := req.RequireString("id")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			args := req.GetArguments()
			var name, description *string
			if v, ok := args["name"].(string); ok {
				name = &v
			}
			if v, ok := args["description"].(string); ok {
				description = &v
			}
			board, err := s.UpdateBoard(ctx, id, name, description)
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return jsonResult(board)
		},
	)

	srv.AddTool(
		mcpgo.NewTool("delete_board",
			mcpgo.WithDescription("Delete a board and all its contents"),
			mcpgo.WithString("id", mcpgo.Required(), mcpgo.Description("Board ID")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			id, err := req.RequireString("id")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			if err := s.DeleteBoard(ctx, id); err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return mcpgo.NewToolResultText("deleted"), nil
		},
	)

	srv.AddTool(
		mcpgo.NewTool("get_board_summary",
			mcpgo.WithDescription("Get ticket counts and epic summary for a board"),
			mcpgo.WithString("id", mcpgo.Required(), mcpgo.Description("Board ID")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			id, err := req.RequireString("id")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			summary, err := s.GetBoardSummary(ctx, id)
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return jsonResult(summary)
		},
	)

	// --- Epics ---

	srv.AddTool(
		mcpgo.NewTool("list_epics",
			mcpgo.WithDescription("List all epics on a board"),
			mcpgo.WithString("board_id", mcpgo.Required(), mcpgo.Description("Board ID")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			boardID, err := req.RequireString("board_id")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			epics, err := s.ListEpics(ctx, boardID)
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return jsonResult(epics)
		},
	)

	srv.AddTool(
		mcpgo.NewTool("create_epic",
			mcpgo.WithDescription("Create a new epic on a board"),
			mcpgo.WithString("board_id", mcpgo.Required(), mcpgo.Description("Board ID")),
			mcpgo.WithString("title", mcpgo.Required(), mcpgo.Description("Epic title")),
			mcpgo.WithString("description", mcpgo.Description("Epic description")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			boardID, err := req.RequireString("board_id")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			title, err := req.RequireString("title")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			description := req.GetString("description", "")
			epic, err := s.CreateEpic(ctx, boardID, title, description)
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return jsonResult(epic)
		},
	)

	srv.AddTool(
		mcpgo.NewTool("update_epic",
			mcpgo.WithDescription("Update an existing epic"),
			mcpgo.WithString("id", mcpgo.Required(), mcpgo.Description("Epic ID")),
			mcpgo.WithString("title", mcpgo.Description("New title")),
			mcpgo.WithString("description", mcpgo.Description("New description")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			id, err := req.RequireString("id")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			args := req.GetArguments()
			var title, description *string
			if v, ok := args["title"].(string); ok {
				title = &v
			}
			if v, ok := args["description"].(string); ok {
				description = &v
			}
			epic, err := s.UpdateEpic(ctx, id, title, description)
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return jsonResult(epic)
		},
	)

	srv.AddTool(
		mcpgo.NewTool("delete_epic",
			mcpgo.WithDescription("Delete an epic"),
			mcpgo.WithString("id", mcpgo.Required(), mcpgo.Description("Epic ID")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			id, err := req.RequireString("id")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			if err := s.DeleteEpic(ctx, id); err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return mcpgo.NewToolResultText("deleted"), nil
		},
	)

	// --- Tickets ---

	srv.AddTool(
		mcpgo.NewTool("list_tickets",
			mcpgo.WithDescription("List tickets on a board, with optional filters and sorting"),
			mcpgo.WithString("board_id", mcpgo.Required(), mcpgo.Description("Board ID")),
			mcpgo.WithString("status", mcpgo.Description("Filter by status: todo, in_progress, done")),
			mcpgo.WithString("priority", mcpgo.Description("Filter by priority: low, medium, high, critical")),
			mcpgo.WithString("epic_id", mcpgo.Description("Filter by epic ID")),
			mcpgo.WithString("q", mcpgo.Description("Keyword search")),
			mcpgo.WithString("sort_by", mcpgo.Description("Sort field: priority | created_at (default: created_at)")),
			mcpgo.WithString("sort_order", mcpgo.Description("Sort direction: asc | desc (default: desc)")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			boardID, err := req.RequireString("board_id")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			args := req.GetArguments()
			filter := models.TicketFilter{}
			if v, ok := args["status"].(string); ok && v != "" {
				st := models.Status(v)
				filter.Status = &st
			}
			if v, ok := args["priority"].(string); ok && v != "" {
				p := models.Priority(v)
				filter.Priority = &p
			}
			if v, ok := args["epic_id"].(string); ok && v != "" {
				filter.EpicID = &v
			}
			if v, ok := args["q"].(string); ok && v != "" {
				filter.Query = &v
			}
			if v, ok := args["sort_by"].(string); ok && v != "" {
				filter.SortBy = &v
			}
			if v, ok := args["sort_order"].(string); ok && v != "" {
				filter.SortOrder = &v
			}
			tickets, err := s.ListTickets(ctx, boardID, filter)
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return jsonResult(tickets)
		},
	)

	srv.AddTool(
		mcpgo.NewTool("create_ticket",
			mcpgo.WithDescription("Create a new ticket on a board"),
			mcpgo.WithString("board_id", mcpgo.Required(), mcpgo.Description("Board ID")),
			mcpgo.WithString("title", mcpgo.Required(), mcpgo.Description("Ticket title")),
			mcpgo.WithString("description", mcpgo.Description("Ticket description")),
			mcpgo.WithString("status", mcpgo.Description("todo | in_progress | done (default: todo)")),
			mcpgo.WithString("priority", mcpgo.Description("low | medium | high | critical (default: medium)")),
			mcpgo.WithString("epic_id", mcpgo.Description("Epic ID to attach to")),
			mcpgo.WithString("assignee", mcpgo.Description("Name of the assignee")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			boardID, err := req.RequireString("board_id")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			title, err := req.RequireString("title")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			args := req.GetArguments()
			t := models.Ticket{
				Title:       title,
				Description: req.GetString("description", ""),
				Status:      models.Status(req.GetString("status", string(models.StatusTodo))),
				Priority:    models.Priority(req.GetString("priority", string(models.PriorityMedium))),
				Assignee:    req.GetString("assignee", ""),
			}
			if v, ok := args["epic_id"].(string); ok && v != "" {
				t.EpicID = &v
			}
			ticket, err := s.CreateTicket(ctx, boardID, t)
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return jsonResult(ticket)
		},
	)

	srv.AddTool(
		mcpgo.NewTool("update_ticket",
			mcpgo.WithDescription("Update fields on an existing ticket"),
			mcpgo.WithString("id", mcpgo.Required(), mcpgo.Description("Ticket ID")),
			mcpgo.WithString("title", mcpgo.Description("New title")),
			mcpgo.WithString("description", mcpgo.Description("New description")),
			mcpgo.WithString("status", mcpgo.Description("New status")),
			mcpgo.WithString("priority", mcpgo.Description("New priority")),
			mcpgo.WithString("epic_id", mcpgo.Description("New epic ID (empty to clear)")),
			mcpgo.WithString("assignee", mcpgo.Description("New assignee (empty to clear)")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			id, err := req.RequireString("id")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			args := req.GetArguments()
			fields := map[string]any{}
			for _, key := range []string{"title", "description", "status", "priority", "epic_id", "assignee"} {
				if v, ok := args[key]; ok {
					fields[key] = v
				}
			}
			ticket, err := s.UpdateTicket(ctx, id, fields)
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return jsonResult(ticket)
		},
	)

	srv.AddTool(
		mcpgo.NewTool("delete_ticket",
			mcpgo.WithDescription("Delete a ticket"),
			mcpgo.WithString("id", mcpgo.Required(), mcpgo.Description("Ticket ID")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			id, err := req.RequireString("id")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			if err := s.DeleteTicket(ctx, id); err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return mcpgo.NewToolResultText("deleted"), nil
		},
	)

	srv.AddTool(
		mcpgo.NewTool("move_ticket",
			mcpgo.WithDescription("Move a ticket to a different status column"),
			mcpgo.WithString("id", mcpgo.Required(), mcpgo.Description("Ticket ID")),
			mcpgo.WithString("status", mcpgo.Required(), mcpgo.Description("Target status: todo, in_progress, done")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			id, err := req.RequireString("id")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			status, err := req.RequireString("status")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			ticket, err := s.UpdateTicket(ctx, id, map[string]any{"status": status})
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return jsonResult(ticket)
		},
	)

	// --- Events ---

	srv.AddTool(
		mcpgo.NewTool("list_ticket_events",
			mcpgo.WithDescription("List the audit trail / history for a ticket"),
			mcpgo.WithString("ticket_id", mcpgo.Required(), mcpgo.Description("Ticket ID")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			ticketID, err := req.RequireString("ticket_id")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			events, err := s.ListTicketEvents(ctx, ticketID)
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return jsonResult(events)
		},
	)

	srv.AddTool(
		mcpgo.NewTool("add_ticket_event",
			mcpgo.WithDescription("Manually append an event to a ticket's audit trail"),
			mcpgo.WithString("ticket_id", mcpgo.Required(), mcpgo.Description("Ticket ID")),
			mcpgo.WithString("type", mcpgo.Required(), mcpgo.Description("Event type: created, moved, edited, commented")),
			mcpgo.WithString("actor", mcpgo.Description("Who performed the action")),
			mcpgo.WithString("payload", mcpgo.Description("JSON object with event details")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			ticketID, err := req.RequireString("ticket_id")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			eventType, err := req.RequireString("type")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			actor := req.GetString("actor", "")
			payload := map[string]any{}
			if raw := req.GetString("payload", ""); raw != "" {
				if err := json.Unmarshal([]byte(raw), &payload); err != nil {
					return mcpgo.NewToolResultError("payload must be valid JSON: " + err.Error()), nil
				}
			}
			event, err := s.CreateTicketEvent(ctx, ticketID, models.TicketEventType(eventType), actor, payload)
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return jsonResult(event)
		},
	)

	// --- Comments ---

	srv.AddTool(
		mcpgo.NewTool("list_comments",
			mcpgo.WithDescription("List comments on a ticket"),
			mcpgo.WithString("ticket_id", mcpgo.Required(), mcpgo.Description("Ticket ID")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			ticketID, err := req.RequireString("ticket_id")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			comments, err := s.ListComments(ctx, ticketID)
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return jsonResult(comments)
		},
	)

	srv.AddTool(
		mcpgo.NewTool("add_comment",
			mcpgo.WithDescription("Add a comment to a ticket"),
			mcpgo.WithString("ticket_id", mcpgo.Required(), mcpgo.Description("Ticket ID")),
			mcpgo.WithString("body", mcpgo.Required(), mcpgo.Description("Comment text")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			ticketID, err := req.RequireString("ticket_id")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			body, err := req.RequireString("body")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			comment, err := s.CreateComment(ctx, ticketID, body)
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return jsonResult(comment)
		},
	)

	srv.AddTool(
		mcpgo.NewTool("update_comment",
			mcpgo.WithDescription("Update a comment's body"),
			mcpgo.WithString("id", mcpgo.Required(), mcpgo.Description("Comment ID")),
			mcpgo.WithString("body", mcpgo.Required(), mcpgo.Description("New comment text")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			id, err := req.RequireString("id")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			body, err := req.RequireString("body")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			comment, err := s.UpdateComment(ctx, id, body)
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return jsonResult(comment)
		},
	)

	srv.AddTool(
		mcpgo.NewTool("delete_comment",
			mcpgo.WithDescription("Delete a comment"),
			mcpgo.WithString("id", mcpgo.Required(), mcpgo.Description("Comment ID")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			id, err := req.RequireString("id")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			if err := s.DeleteComment(ctx, id); err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			return mcpgo.NewToolResultText("deleted"), nil
		},
	)
}
