package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rgracey/kanban-mcp/internal/models"
	"github.com/rgracey/kanban-mcp/internal/store"
)

func registerTools(srv *server.MCPServer, s store.Store) {

	// -------------------------------------------------------------------------
	// board — list | get | create | update | delete | summary
	// -------------------------------------------------------------------------
	srv.AddTool(
		mcpgo.NewTool("board",
			mcpgo.WithDescription("Manage boards. action: list, get, create, update, delete, summary, context, ready"),
			mcpgo.WithString("action", mcpgo.Required(), mcpgo.Description("list|get|create|update|delete|summary|context|ready")),
			mcpgo.WithString("id", mcpgo.Description("Board ID (get/update/delete/summary)")),
			mcpgo.WithString("name", mcpgo.Description("Board name (create/update)")),
			mcpgo.WithString("description", mcpgo.Description("Board description (create/update)")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			action, err := req.RequireString("action")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			args := req.GetArguments()
			getString := func(k string) string {
				v, _ := args[k].(string)
				return v
			}

			switch action {
			case "list":
				boards, err := s.ListBoards(ctx)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonListResult(boards)

			case "get":
				id := getString("id")
				if id == "" {
					return mcpgo.NewToolResultError("id required"), nil
				}
				board, err := s.GetBoard(ctx, id)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonResult(board)

			case "create":
				name := getString("name")
				if name == "" {
					return mcpgo.NewToolResultError("name required"), nil
				}
				board, err := s.CreateBoard(ctx, name, getString("description"))
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonResult(board)

			case "update":
				id := getString("id")
				if id == "" {
					return mcpgo.NewToolResultError("id required"), nil
				}
				var name, desc *string
				if v := getString("name"); v != "" {
					name = &v
				}
				if v := getString("description"); v != "" {
					desc = &v
				}
				board, err := s.UpdateBoard(ctx, id, name, desc)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonResult(board)

			case "delete":
				id := getString("id")
				if id == "" {
					return mcpgo.NewToolResultError("id required"), nil
				}
				if err := s.DeleteBoard(ctx, id); err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return mcpgo.NewToolResultText("deleted"), nil

			case "summary":
				id := getString("id")
				if id == "" {
					return mcpgo.NewToolResultError("id required"), nil
				}
				summary, err := s.GetBoardSummary(ctx, id)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonResult(summary)

			case "context":
				// Returns the full board snapshot: board metadata + epics + all tickets
				// with embedded tasks and relations. Use this instead of multiple list
				// calls when you need a complete picture of the board.
				id := getString("id")
				if id == "" {
					return mcpgo.NewToolResultError("id required"), nil
				}
				bctx, err := s.BoardContext(ctx, id)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonResult(bctx)

			case "ready":
				// Returns unblocked todo tickets ordered by priority descending.
				// Use this to get an agent's immediate work queue.
				id := getString("id")
				if id == "" {
					return mcpgo.NewToolResultError("id required"), nil
				}
				tickets, err := s.ReadyTickets(ctx, id)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonListResult(tickets)

			default:
				return mcpgo.NewToolResultError(fmt.Sprintf("unknown action %q", action)), nil
			}
		},
	)

	// -------------------------------------------------------------------------
	// epic — list | get | create | update | delete
	// -------------------------------------------------------------------------
	srv.AddTool(
		mcpgo.NewTool("epic",
			mcpgo.WithDescription("Manage epics. action: list, get, create, update, delete"),
			mcpgo.WithString("action", mcpgo.Required(), mcpgo.Description("list|get|create|update|delete")),
			mcpgo.WithString("id", mcpgo.Description("Epic ID (get/update/delete)")),
			mcpgo.WithString("board_id", mcpgo.Description("Board ID (list/create)")),
			mcpgo.WithString("title", mcpgo.Description("Epic title (create/update)")),
			mcpgo.WithString("description", mcpgo.Description("Epic description (create/update)")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			action, err := req.RequireString("action")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			args := req.GetArguments()
			getString := func(k string) string {
				v, _ := args[k].(string)
				return v
			}

			switch action {
			case "list":
				boardID := getString("board_id")
				if boardID == "" {
					return mcpgo.NewToolResultError("board_id required"), nil
				}
				epics, err := s.ListEpics(ctx, boardID)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonListResult(epics)

			case "get":
				id := getString("id")
				if id == "" {
					return mcpgo.NewToolResultError("id required"), nil
				}
				epic, err := s.GetEpic(ctx, id)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonResult(epic)

			case "create":
				boardID := getString("board_id")
				title := getString("title")
				if boardID == "" || title == "" {
					return mcpgo.NewToolResultError("board_id and title required"), nil
				}
				epic, err := s.CreateEpic(ctx, boardID, title, getString("description"))
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonResult(epic)

			case "update":
				id := getString("id")
				if id == "" {
					return mcpgo.NewToolResultError("id required"), nil
				}
				var title, desc *string
				if v := getString("title"); v != "" {
					title = &v
				}
				if v := getString("description"); v != "" {
					desc = &v
				}
				epic, err := s.UpdateEpic(ctx, id, title, desc)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonResult(epic)

			case "delete":
				id := getString("id")
				if id == "" {
					return mcpgo.NewToolResultError("id required"), nil
				}
				if err := s.DeleteEpic(ctx, id); err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return mcpgo.NewToolResultText("deleted"), nil

			default:
				return mcpgo.NewToolResultError(fmt.Sprintf("unknown action %q", action)), nil
			}
		},
	)

	// -------------------------------------------------------------------------
	// ticket — list | get | create | update | delete | move | history
	// -------------------------------------------------------------------------
	srv.AddTool(
		mcpgo.NewTool("ticket",
			mcpgo.WithDescription("Manage tickets. action: list, get, create, bulk_create, update, delete, move, history"),
			mcpgo.WithString("action", mcpgo.Required(), mcpgo.Description("list|get|create|bulk_create|update|delete|move|history")),
			mcpgo.WithString("id", mcpgo.Description("Ticket ID (get/update/delete/move/history)")),
			mcpgo.WithString("board_id", mcpgo.Description("Board ID (list/create/bulk_create)")),
			mcpgo.WithString("title", mcpgo.Description("Title (create/update)")),
			mcpgo.WithString("description", mcpgo.Description("Description (create/update)")),
			mcpgo.WithString("status", mcpgo.Description("todo|in_progress|done")),
			mcpgo.WithString("priority", mcpgo.Description("low|medium|high|critical")),
			mcpgo.WithString("epic_id", mcpgo.Description("Epic ID, empty to clear")),
			mcpgo.WithString("assignee", mcpgo.Description("Assignee name, empty to clear")),
			mcpgo.WithString("filter_status", mcpgo.Description("Filter by status (list)")),
			mcpgo.WithString("filter_priority", mcpgo.Description("Filter by priority (list)")),
			mcpgo.WithString("filter_epic_id", mcpgo.Description("Filter by epic (list)")),
			mcpgo.WithString("q", mcpgo.Description("Keyword search (list)")),
			mcpgo.WithString("sort_by", mcpgo.Description("priority|created_at (list)")),
			mcpgo.WithString("sort_order", mcpgo.Description("asc|desc (list)")),
			mcpgo.WithBoolean("include_notes", mcpgo.Description("Embed agent notes (get)")),
			mcpgo.WithBoolean("include_history", mcpgo.Description("Embed audit history (get)")),
			mcpgo.WithString("tickets_json", mcpgo.Description(`JSON array of ticket objects for bulk_create, e.g. [{"title":"T1","priority":"high"},{"title":"T2"}]`)),
			mcpgo.WithString("references_json", mcpgo.Description(`JSON array of code references for create/update, e.g. [{"kind":"file","target":"src/api/handler.go:42","label":"handler"},{"kind":"pr","target":"https://github.com/..."}]`)),
			mcpgo.WithString("resolution_json", mcpgo.Description(`JSON object to record resolution for update, e.g. {"commit_sha":"abc123","pr_url":"https://...","notes":"Fixed by ...","resolved_at":"2026-01-01T00:00:00Z"}. Pass "null" to clear.`)),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			action, err := req.RequireString("action")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			args := req.GetArguments()
			getString := func(k string) string {
				v, _ := args[k].(string)
				return v
			}

			switch action {
			case "list":
				boardID := getString("board_id")
				if boardID == "" {
					return mcpgo.NewToolResultError("board_id required"), nil
				}
				filter := models.TicketFilter{}
				if v := getString("filter_status"); v != "" {
					st := models.Status(v)
					filter.Status = &st
				}
				if v := getString("filter_priority"); v != "" {
					p := models.Priority(v)
					filter.Priority = &p
				}
				if v := getString("filter_epic_id"); v != "" {
					filter.EpicID = &v
				}
				if v := getString("q"); v != "" {
					filter.Query = &v
				}
				if v := getString("sort_by"); v != "" {
					filter.SortBy = &v
				}
				if v := getString("sort_order"); v != "" {
					filter.SortOrder = &v
				}
				tickets, err := s.ListTickets(ctx, boardID, filter)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonListResult(tickets)

			case "get":
				id := getString("id")
				if id == "" {
					return mcpgo.NewToolResultError("id required"), nil
				}
				ticket, err := s.GetTicket(ctx, id)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				type envelope struct {
					models.Ticket
					Notes   []models.Note        `json:"notes,omitempty"`
					History []models.TicketEvent `json:"history,omitempty"`
				}
				out := envelope{Ticket: ticket}
				includeNotes, _ := args["include_notes"].(bool)
				if includeNotes {
					notes, err := s.ListNotes(ctx, id)
					if err != nil {
						return mcpgo.NewToolResultError(err.Error()), nil
					}
					out.Notes = notes
				}
				if v, ok := args["include_history"].(bool); ok && v {
					events, err := s.ListTicketEvents(ctx, id)
					if err != nil {
						return mcpgo.NewToolResultError(err.Error()), nil
					}
					out.History = events
				}
				return jsonResult(out)

			case "create":
				boardID := getString("board_id")
				title := getString("title")
				if boardID == "" || title == "" {
					return mcpgo.NewToolResultError("board_id and title required"), nil
				}
				t := models.Ticket{
					Title:       title,
					Description: getString("description"),
					Status:      models.Status(req.GetString("status", string(models.StatusTodo))),
					Priority:    models.Priority(req.GetString("priority", string(models.PriorityMedium))),
					Assignee:    getString("assignee"),
				}
				if v := getString("epic_id"); v != "" {
					t.EpicID = &v
				}
				if raw := getString("references_json"); raw != "" {
					var refs []models.TicketReference
					if err := json.Unmarshal([]byte(raw), &refs); err != nil {
						return mcpgo.NewToolResultError("references_json is not valid JSON: " + err.Error()), nil
					}
					t.References = refs
				}
				if raw := getString("resolution_json"); raw != "" && raw != "null" {
					var res models.TicketResolution
					if err := json.Unmarshal([]byte(raw), &res); err != nil {
						return mcpgo.NewToolResultError("resolution_json is not valid JSON: " + err.Error()), nil
					}
					t.Resolution = &res
				}
				ticket, err := s.CreateTicket(ctx, boardID, t)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonResult(ticket)

			case "update":
				id := getString("id")
				if id == "" {
					return mcpgo.NewToolResultError("id required"), nil
				}
				fields := map[string]any{}
				for _, k := range []string{"title", "description", "status", "priority", "epic_id", "assignee"} {
					if v, ok := args[k]; ok {
						fields[k] = v
					}
				}
				if raw := getString("references_json"); raw != "" {
					fields["references"] = raw
				}
				if raw := getString("resolution_json"); raw != "" {
					if raw == "null" {
						fields["resolution"] = nil
					} else {
						fields["resolution"] = raw
					}
				}
				ticket, err := s.UpdateTicket(ctx, id, fields)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonResult(ticket)

			case "delete":
				id := getString("id")
				if id == "" {
					return mcpgo.NewToolResultError("id required"), nil
				}
				if err := s.DeleteTicket(ctx, id); err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return mcpgo.NewToolResultText("deleted"), nil

			case "move":
				id := getString("id")
				status := getString("status")
				if id == "" || status == "" {
					return mcpgo.NewToolResultError("id and status required"), nil
				}
				ticket, err := s.UpdateTicket(ctx, id, map[string]any{"status": status})
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonResult(ticket)

			case "bulk_create":
				boardID := getString("board_id")
				if boardID == "" {
					return mcpgo.NewToolResultError("board_id required"), nil
				}
				raw := getString("tickets_json")
				if raw == "" {
					return mcpgo.NewToolResultError("tickets_json required — JSON array of ticket objects"), nil
				}
				var inputs []struct {
					Title       string `json:"title"`
					Description string `json:"description"`
					Status      string `json:"status"`
					Priority    string `json:"priority"`
					EpicID      string `json:"epic_id"`
					Assignee    string `json:"assignee"`
				}
				if err := json.Unmarshal([]byte(raw), &inputs); err != nil {
					return mcpgo.NewToolResultError("tickets_json is not valid JSON: " + err.Error()), nil
				}
				if len(inputs) == 0 {
					return mcpgo.NewToolResultError("tickets_json must contain at least one ticket"), nil
				}
				tickets := make([]models.Ticket, 0, len(inputs))
				for _, inp := range inputs {
					if inp.Title == "" {
						return mcpgo.NewToolResultError("every ticket must have a title"), nil
					}
					t := models.Ticket{
						Title:       inp.Title,
						Description: inp.Description,
						Status:      models.Status(inp.Status),
						Priority:    models.Priority(inp.Priority),
						Assignee:    inp.Assignee,
					}
					if inp.EpicID != "" {
						t.EpicID = &inp.EpicID
					}
					tickets = append(tickets, t)
				}
				created, err := s.BulkCreateTickets(ctx, boardID, tickets)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonListResult(created)

			case "history":
				id := getString("id")
				if id == "" {
					return mcpgo.NewToolResultError("id required"), nil
				}
				events, err := s.ListTicketEvents(ctx, id)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonListResult(events)

			default:
				return mcpgo.NewToolResultError(fmt.Sprintf("unknown action %q", action)), nil
			}
		},
	)

	// -------------------------------------------------------------------------
	// task — list | create | update | delete
	// -------------------------------------------------------------------------
	srv.AddTool(
		mcpgo.NewTool("task",
			mcpgo.WithDescription("Manage checklist tasks on a ticket. action: list, create, update, delete"),
			mcpgo.WithString("action", mcpgo.Required(), mcpgo.Description("list|create|update|delete")),
			mcpgo.WithString("id", mcpgo.Description("Task ID (update/delete)")),
			mcpgo.WithString("ticket_id", mcpgo.Description("Ticket ID (list/create)")),
			mcpgo.WithString("title", mcpgo.Description("Task title (create/update)")),
			mcpgo.WithBoolean("done", mcpgo.Description("Mark done/undone (update)")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			action, err := req.RequireString("action")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			args := req.GetArguments()
			getString := func(k string) string {
				v, _ := args[k].(string)
				return v
			}

			switch action {
			case "list":
				ticketID := getString("ticket_id")
				if ticketID == "" {
					return mcpgo.NewToolResultError("ticket_id required"), nil
				}
				tasks, err := s.ListTasks(ctx, ticketID)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonListResult(tasks)

			case "create":
				ticketID := getString("ticket_id")
				title := getString("title")
				if ticketID == "" || title == "" {
					return mcpgo.NewToolResultError("ticket_id and title required"), nil
				}
				task, err := s.CreateTask(ctx, ticketID, title)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonResult(task)

			case "update":
				id := getString("id")
				if id == "" {
					return mcpgo.NewToolResultError("id required"), nil
				}
				var title *string
				if v := getString("title"); v != "" {
					title = &v
				}
				var done *bool
				if v, ok := args["done"].(bool); ok {
					done = &v
				}
				task, err := s.UpdateTask(ctx, id, title, done)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonResult(task)

			case "delete":
				id := getString("id")
				if id == "" {
					return mcpgo.NewToolResultError("id required"), nil
				}
				if err := s.DeleteTask(ctx, id); err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return mcpgo.NewToolResultText("deleted"), nil

			default:
				return mcpgo.NewToolResultError(fmt.Sprintf("unknown action %q", action)), nil
			}
		},
	)

	// -------------------------------------------------------------------------
	// note — list | add | update | delete  (agent scratchpad)
	// -------------------------------------------------------------------------
	srv.AddTool(
		mcpgo.NewTool("note",
			mcpgo.WithDescription("Manage agent scratchpad notes on a ticket. Use notes to record observations, intermediate reasoning, investigation logs, and any other machine-readable context. action: list, add, update, delete"),
			mcpgo.WithString("action", mcpgo.Required(), mcpgo.Description("list|add|update|delete")),
			mcpgo.WithString("id", mcpgo.Description("Note ID (update/delete)")),
			mcpgo.WithString("ticket_id", mcpgo.Description("Ticket ID (list/add)")),
			mcpgo.WithString("body", mcpgo.Description("Note text (add/update)")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			action, err := req.RequireString("action")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			args := req.GetArguments()
			getString := func(k string) string {
				v, _ := args[k].(string)
				return v
			}

			switch action {
			case "list":
				ticketID := getString("ticket_id")
				if ticketID == "" {
					return mcpgo.NewToolResultError("ticket_id required"), nil
				}
				notes, err := s.ListNotes(ctx, ticketID)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonListResult(notes)

			case "add":
				ticketID := getString("ticket_id")
				body := getString("body")
				if ticketID == "" || body == "" {
					return mcpgo.NewToolResultError("ticket_id and body required"), nil
				}
				note, err := s.CreateNote(ctx, ticketID, body)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonResult(note)

			case "update":
				id := getString("id")
				body := getString("body")
				if id == "" || body == "" {
					return mcpgo.NewToolResultError("id and body required"), nil
				}
				note, err := s.UpdateNote(ctx, id, body)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonResult(note)

			case "delete":
				id := getString("id")
				if id == "" {
					return mcpgo.NewToolResultError("id required"), nil
				}
				if err := s.DeleteNote(ctx, id); err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return mcpgo.NewToolResultText("deleted"), nil

			default:
				return mcpgo.NewToolResultError(fmt.Sprintf("unknown action %q", action)), nil
			}
		},
	)

	// -------------------------------------------------------------------------
	// relation — list | add | delete
	// -------------------------------------------------------------------------
	srv.AddTool(
		mcpgo.NewTool("relation",
			mcpgo.WithDescription("Manage blocking relations between tickets. action: list, add, delete"),
			mcpgo.WithString("action", mcpgo.Required(), mcpgo.Description("list|add|delete")),
			mcpgo.WithString("ticket_id", mcpgo.Description("Ticket ID (list/add/delete)")),
			mcpgo.WithString("to_ticket_id", mcpgo.Description("ID of the ticket being blocked (add/delete)")),
		),
		func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
			action, err := req.RequireString("action")
			if err != nil {
				return mcpgo.NewToolResultError(err.Error()), nil
			}
			args := req.GetArguments()
			getString := func(k string) string {
				v, _ := args[k].(string)
				return v
			}

			switch action {
			case "list":
				ticketID := getString("ticket_id")
				if ticketID == "" {
					return mcpgo.NewToolResultError("ticket_id required"), nil
				}
				relations, err := s.ListRelations(ctx, ticketID)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonListResult(relations)

			case "add":
				fromID := getString("ticket_id")
				toID := getString("to_ticket_id")
				if fromID == "" || toID == "" {
					return mcpgo.NewToolResultError("ticket_id and to_ticket_id required"), nil
				}
				if fromID == toID {
					return mcpgo.NewToolResultError("a ticket cannot block itself"), nil
				}
				rel, err := s.AddRelation(ctx, fromID, toID, models.RelationBlocks)
				if err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return jsonResult(rel)

			case "delete":
				fromID := getString("ticket_id")
				toID := getString("to_ticket_id")
				if fromID == "" || toID == "" {
					return mcpgo.NewToolResultError("ticket_id and to_ticket_id required"), nil
				}
				if err := s.DeleteRelation(ctx, fromID, toID, models.RelationBlocks); err != nil {
					return mcpgo.NewToolResultError(err.Error()), nil
				}
				return mcpgo.NewToolResultText("deleted"), nil

			default:
				return mcpgo.NewToolResultError(fmt.Sprintf("unknown action %q", action)), nil
			}
		},
	)
}
