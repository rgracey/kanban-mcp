# 0011 — Frontend Kanban Board UI

## Goal

Build the main kanban board interface: board switcher, three-column drag-and-drop board, epic filter, and ticket cards.

## Dependencies

- Ticket 0010 (API client + types)

## Components

### `App.svelte`

- On mount, call `listBoards()`. Store result in a reactive `boards` variable.
- Render `<BoardSwitcher>` passing the boards list and the selected board ID.
- When a board is selected, render `<KanbanBoard boardId={selectedId} />`.
- If no boards exist, show a prompt to create the first board.

### `BoardSwitcher.svelte`

Props: `boards: Board[]`, `selectedId: string | null`

- Renders a horizontal tab strip or dropdown listing board names.
- Emits a `select` event with the board ID when a board is clicked.
- Includes a "New board" button that opens `<CreateBoard>` modal.

### `KanbanBoard.svelte`

Props: `boardId: string`

- On mount (and when `boardId` changes), call `listTickets(boardId)` and `listEpics(boardId)`.
- Renders `<EpicFilter>` to allow filtering by epic.
- Renders three `<KanbanColumn>` components — one per status: `todo`, `in_progress`, `done`.
- Passes each column the filtered subset of tickets matching that status.
- Handles the `drop` event from `svelte-dnd-action`: when a ticket is dropped into a different column, call `updateTicket(id, { status: newStatus })` then update local state optimistically.
- Includes a "New ticket" button that opens `<CreateTicket boardId={boardId}>` modal.

### `KanbanColumn.svelte`

Props: `status: Status`, `label: string`, `tickets: Ticket[]`, `epics: Epic[]`

- Uses `svelte-dnd-action`'s `dndzone` action on the column container.
- Renders a `<TicketCard>` for each ticket.
- Displays the column label and ticket count in the header.

### `TicketCard.svelte`

Props: `ticket: Ticket`, `epics: Epic[]`

- Displays: ticket title, priority badge (colour-coded), epic name (if set).
- Clicking the card opens `<TicketDetail>`.

### `EpicFilter.svelte`

Props: `epics: Epic[]`, `selectedEpicId: string | null`

- Renders "All" option plus one option per epic.
- Emits `change` with the selected epic ID (or `null` for All).

### Priority badge colours (Tailwind classes)

| Priority | Classes |
|---|---|
| `low` | `bg-gray-100 text-gray-600` |
| `medium` | `bg-blue-100 text-blue-700` |
| `high` | `bg-orange-100 text-orange-700` |
| `critical` | `bg-red-100 text-red-700` |

## Acceptance Criteria

- `cd frontend && npm run build` exits 0 with no TypeScript or Svelte errors.
- Opening the app in a browser shows the board switcher.
- Selecting a board shows three columns with tickets correctly distributed by status.
- Dragging a ticket to another column calls `updateTicket` with the new status and the card appears in the new column without a page refresh.
- The epic filter hides tickets not belonging to the selected epic.
- A board with no tickets shows three empty columns (not an error state).
