# 0012 — Frontend Ticket Detail Panel and Modals

## Goal

Implement the ticket detail slide-over panel and all create/edit/delete modals for boards, epics, tickets, and comments.

## Dependencies

- Ticket 0011 (board UI — opens these components)

## Components

### `TicketDetail.svelte`

Props: `ticketId: string`, `boardId: string`

- On mount, call `getTicket(ticketId)` and `listComments(ticketId)` and `listEpics(boardId)`.
- Renders as a slide-over panel (fixed right side, overlay backdrop).
- Emits `close` event when the backdrop or close button is clicked.
- **Editable fields** (inline edit on click, confirmed on blur or Enter):
  - Title (text input)
  - Description (textarea, supports markdown — render as plain textarea, no live preview required)
  - Status (select: `todo`, `in_progress`, `done`)
  - Priority (select: `low`, `medium`, `high`, `critical`)
  - Epic (select: epics on this board + "None" option)
- On field change, call `updateTicket(id, { field: newValue })`.
- **Comments section**:
  - Lists comments ordered oldest-first.
  - Each comment shows body and timestamp.
  - Edit button on each comment: inline textarea + save/cancel.
  - Delete button on each comment: calls `deleteComment(id)`, removes from list.
  - "Add comment" textarea at the bottom with a submit button: calls `createComment(ticketId, body)`.
- **Delete ticket** button at the bottom: calls `deleteTicket(id)`, emits `deleted` event, closes panel.

### `modals/CreateBoard.svelte`

- Form: `name` (required text input), `description` (optional textarea).
- Submit calls `createBoard(name, description)`.
- On success, emits `created` event with the new `Board` object.
- On error, displays the error message inline.
- Cancel button emits `close`.

### `modals/CreateEpic.svelte`

Props: `boardId: string`

- Form: `title` (required), `description` (optional).
- Submit calls `createEpic(boardId, title, description)`.
- On success, emits `created` with the new `Epic`.

### `modals/CreateTicket.svelte`

Props: `boardId: string`

- Form: `title` (required), `description` (optional), `priority` (select, default `medium`), `epic` (select from board epics + "None").
- Status defaults to `todo` and is not shown in the creation form.
- Submit calls `createTicket(boardId, data)`.
- On success, emits `created` with the new `Ticket`.

### Modal wrapper behaviour (apply to all modals)

- Rendered in a centered overlay with a dark backdrop.
- Pressing `Escape` closes the modal.
- Clicking the backdrop closes the modal.
- While submitting, the submit button is disabled and shows a loading state.

## Acceptance Criteria

- `cd frontend && npm run build` exits 0.
- Clicking a ticket card opens `TicketDetail`; editing a field and blurring calls the API and reflects the updated value without page refresh.
- Adding a comment via the detail panel appends it to the list immediately.
- Deleting a ticket closes the panel and removes the card from the board.
- Creating a board via the modal adds it to the board switcher immediately.
- Creating an epic via the modal makes it available in the epic filter and ticket creation form without page refresh.
- Creating a ticket via the modal adds the card to the `To Do` column immediately.
- All modals close on `Escape` or backdrop click.
- Form submit buttons are disabled while the API call is in flight.
