# REST API Contracts

Base path: `/api/v1`  
All requests/responses: `Content-Type: application/json`  
All IDs: UUID v4 strings.  
Timestamps: RFC3339 UTC strings.

---

## Error Response

All errors return a consistent envelope:

```json
{
  "error": "human-readable message"
}
```

| HTTP status | When |
|---|---|
| 400 | Validation failure, malformed JSON |
| 404 | Resource not found |
| 500 | Internal server error |

---

## Boards

### `GET /api/v1/boards`

Response `200`:
```json
[
  {
    "id": "uuid",
    "name": "string",
    "description": "string",
    "created_at": "RFC3339",
    "updated_at": "RFC3339"
  }
]
```

### `POST /api/v1/boards`

Request:
```json
{ "name": "string (required)", "description": "string (optional)" }
```
Response `201`: Board object.

### `GET /api/v1/boards/:id`

Response `200`: Board object. `404` if not found.

### `PUT /api/v1/boards/:id`

Request: any subset of `{ "name", "description" }`. Only provided fields are updated.  
Response `200`: Updated board object.

### `DELETE /api/v1/boards/:id`

Response `204`. Cascades: deletes all epics, tickets, and comments on that board.

### `GET /api/v1/boards/:id/summary`

Response `200`:
```json
{
  "board_id": "uuid",
  "ticket_counts": {
    "todo": 3,
    "in_progress": 1,
    "done": 7
  },
  "epics": [
    { "id": "uuid", "title": "string", "ticket_count": 4 }
  ]
}
```

---

## Epics

### `GET /api/v1/boards/:id/epics`

Response `200`: Array of epic objects.

Epic object:
```json
{
  "id": "uuid",
  "board_id": "uuid",
  "title": "string",
  "description": "string",
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

### `POST /api/v1/boards/:id/epics`

Request: `{ "title": "string (required)", "description": "string (optional)" }`  
Response `201`: Epic object.

### `GET /api/v1/epics/:id`

Response `200`: Epic object. `404` if not found.

### `PUT /api/v1/epics/:id`

Request: any subset of `{ "title", "description" }`.  
Response `200`: Updated epic object.

### `DELETE /api/v1/epics/:id`

Response `204`. Tickets that belonged to this epic have `epic_id` set to `null` — they are not deleted.

---

## Tickets

### `GET /api/v1/boards/:id/tickets`

Query params (all optional):
- `status` — `todo` | `in_progress` | `done`
- `priority` — `low` | `medium` | `high` | `critical`
- `epic_id` — UUID
- `q` — keyword search against `title` and `description` (case-insensitive, substring match)

Response `200`: Array of ticket objects.

Ticket object:
```json
{
  "id": "uuid",
  "board_id": "uuid",
  "epic_id": "uuid or null",
  "title": "string",
  "description": "string",
  "status": "todo | in_progress | done",
  "priority": "low | medium | high | critical",
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

### `POST /api/v1/boards/:id/tickets`

Request:
```json
{
  "title": "string (required)",
  "description": "string (optional)",
  "status": "todo | in_progress | done (optional, default: todo)",
  "priority": "low | medium | high | critical (optional, default: medium)",
  "epic_id": "uuid (optional)"
}
```
Response `201`: Ticket object.

### `GET /api/v1/tickets/:id`

Response `200`: Ticket object. `404` if not found.

### `PUT /api/v1/tickets/:id`

Request: any subset of `{ "title", "description", "status", "priority", "epic_id" }`.  
Response `200`: Updated ticket object.

### `DELETE /api/v1/tickets/:id`

Response `204`. Cascades: deletes all comments on this ticket.

---

## Comments

### `GET /api/v1/tickets/:id/comments`

Response `200`: Array of comment objects, ordered by `created_at` ascending.

Comment object:
```json
{
  "id": "uuid",
  "ticket_id": "uuid",
  "body": "string",
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

### `POST /api/v1/tickets/:id/comments`

Request: `{ "body": "string (required)" }`  
Response `201`: Comment object.

### `PUT /api/v1/comments/:id`

Request: `{ "body": "string (required)" }`  
Response `200`: Updated comment object.

### `DELETE /api/v1/comments/:id`

Response `204`.
