# 0010 — Frontend API Client and Types

## Goal

Create typed TypeScript definitions mirroring the Go models and a thin `fetch`-based API client used by all frontend components.

## Dependencies

- Ticket 0009 (frontend scaffold)

## Tasks

### `frontend/src/lib/types.ts`

```ts
export type Status = 'todo' | 'in_progress' | 'done'
export type Priority = 'low' | 'medium' | 'high' | 'critical'

export interface Board {
  id: string
  name: string
  description: string
  created_at: string
  updated_at: string
}

export interface Epic {
  id: string
  board_id: string
  title: string
  description: string
  created_at: string
  updated_at: string
}

export interface Ticket {
  id: string
  board_id: string
  epic_id: string | null
  title: string
  description: string
  status: Status
  priority: Priority
  created_at: string
  updated_at: string
}

export interface Comment {
  id: string
  ticket_id: string
  body: string
  created_at: string
  updated_at: string
}

export interface BoardSummary {
  board_id: string
  ticket_counts: Record<Status, number>
  epics: { id: string; title: string; ticket_count: number }[]
}

export interface TicketFilter {
  status?: Status
  priority?: Priority
  epic_id?: string
  q?: string
}
```

### `frontend/src/lib/api.ts`

Base URL: `/api/v1` (relative — works in dev proxy and production).

Implement typed async functions for every REST endpoint. All functions throw an `Error` with the server's `error` message on non-2xx responses.

```ts
const BASE = '/api/v1'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(body.error ?? res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

// Boards
export const listBoards = () => request<Board[]>('/boards')
export const createBoard = (name: string, description?: string) =>
  request<Board>('/boards', { method: 'POST', body: JSON.stringify({ name, description }) })
export const getBoard = (id: string) => request<Board>(`/boards/${id}`)
export const updateBoard = (id: string, fields: Partial<Pick<Board, 'name' | 'description'>>) =>
  request<Board>(`/boards/${id}`, { method: 'PUT', body: JSON.stringify(fields) })
export const deleteBoard = (id: string) => request<void>(`/boards/${id}`, { method: 'DELETE' })
export const getBoardSummary = (id: string) => request<BoardSummary>(`/boards/${id}/summary`)

// Epics
export const listEpics = (boardId: string) => request<Epic[]>(`/boards/${boardId}/epics`)
export const createEpic = (boardId: string, title: string, description?: string) =>
  request<Epic>(`/boards/${boardId}/epics`, { method: 'POST', body: JSON.stringify({ title, description }) })
export const getEpic = (id: string) => request<Epic>(`/epics/${id}`)
export const updateEpic = (id: string, fields: Partial<Pick<Epic, 'title' | 'description'>>) =>
  request<Epic>(`/epics/${id}`, { method: 'PUT', body: JSON.stringify(fields) })
export const deleteEpic = (id: string) => request<void>(`/epics/${id}`, { method: 'DELETE' })

// Tickets
export const listTickets = (boardId: string, filter?: TicketFilter) => {
  const params = new URLSearchParams()
  if (filter?.status)   params.set('status', filter.status)
  if (filter?.priority) params.set('priority', filter.priority)
  if (filter?.epic_id)  params.set('epic_id', filter.epic_id)
  if (filter?.q)        params.set('q', filter.q)
  const qs = params.size ? '?' + params.toString() : ''
  return request<Ticket[]>(`/boards/${boardId}/tickets${qs}`)
}
export const createTicket = (boardId: string, data: Partial<Ticket> & { title: string }) =>
  request<Ticket>(`/boards/${boardId}/tickets`, { method: 'POST', body: JSON.stringify(data) })
export const getTicket = (id: string) => request<Ticket>(`/tickets/${id}`)
export const updateTicket = (id: string, fields: Partial<Pick<Ticket, 'title' | 'description' | 'status' | 'priority' | 'epic_id'>>) =>
  request<Ticket>(`/tickets/${id}`, { method: 'PUT', body: JSON.stringify(fields) })
export const deleteTicket = (id: string) => request<void>(`/tickets/${id}`, { method: 'DELETE' })

// Comments
export const listComments = (ticketId: string) => request<Comment[]>(`/tickets/${ticketId}/comments`)
export const createComment = (ticketId: string, body: string) =>
  request<Comment>(`/tickets/${ticketId}/comments`, { method: 'POST', body: JSON.stringify({ body }) })
export const updateComment = (id: string, body: string) =>
  request<Comment>(`/comments/${id}`, { method: 'PUT', body: JSON.stringify({ body }) })
export const deleteComment = (id: string) => request<void>(`/comments/${id}`, { method: 'DELETE' })
```

### Vite dev proxy

Add to `frontend/vite.config.ts` so `fetch('/api/v1/...')` is proxied to the Go server during development:

```ts
server: {
  proxy: {
    '/api': 'http://localhost:8080',
  },
},
```

## Acceptance Criteria

- `cd frontend && npm run build` exits 0 with no TypeScript errors.
- `cd frontend && npx tsc --noEmit` exits 0.
- Every REST endpoint listed in `docs/api_contracts.md` has a corresponding exported function in `api.ts`.
- Types in `types.ts` match the JSON field names in `docs/api_contracts.md` exactly.
