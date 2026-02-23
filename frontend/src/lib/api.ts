import type {
	Board,
	Epic,
	Ticket,
	Note,
	Task,
	TicketEvent,
	TicketRelation,
	BoardSummary,
	TicketFilter,
	Status,
	Priority,
	SortBy,
	SortOrder
} from './types.js';

const BASE = '/api/v1';

async function request<T>(path: string, init?: RequestInit): Promise<T> {
	const res = await fetch(BASE + path, {
		headers: { 'Content-Type': 'application/json' },
		...init
	});
	if (!res.ok) {
		const body = await res.json().catch(() => ({ error: res.statusText }));
		throw new Error(body.error ?? res.statusText);
	}
	if (res.status === 204) return undefined as T;
	return res.json();
}

// Boards
export const listBoards = () => request<Board[]>('/boards');
export const createBoard = (name: string, description?: string) =>
	request<Board>('/boards', { method: 'POST', body: JSON.stringify({ name, description }) });
export const getBoard = (id: string) => request<Board>(`/boards/${id}`);
export const updateBoard = (id: string, fields: Partial<Pick<Board, 'name' | 'description'>>) =>
	request<Board>(`/boards/${id}`, { method: 'PUT', body: JSON.stringify(fields) });
export const deleteBoard = (id: string) => request<void>(`/boards/${id}`, { method: 'DELETE' });
export const getBoardSummary = (id: string) => request<BoardSummary>(`/boards/${id}/summary`);

// Epics
export const listEpics = (boardId: string) => request<Epic[]>(`/boards/${boardId}/epics`);
export const createEpic = (boardId: string, title: string, description?: string) =>
	request<Epic>(`/boards/${boardId}/epics`, {
		method: 'POST',
		body: JSON.stringify({ title, description })
	});
export const getEpic = (id: string) => request<Epic>(`/epics/${id}`);
export const updateEpic = (id: string, fields: Partial<Pick<Epic, 'title' | 'description'>>) =>
	request<Epic>(`/epics/${id}`, { method: 'PUT', body: JSON.stringify(fields) });
export const deleteEpic = (id: string) => request<void>(`/epics/${id}`, { method: 'DELETE' });

// Tickets
export const listTickets = (boardId: string, filter?: TicketFilter) => {
	const params = new URLSearchParams();
	if (filter?.status) params.set('status', filter.status);
	if (filter?.priority) params.set('priority', filter.priority);
	if (filter?.epic_id) params.set('epic_id', filter.epic_id);
	if (filter?.assignee) params.set('assignee', filter.assignee);
	if (filter?.q) params.set('q', filter.q);
	if (filter?.sort_by) params.set('sort_by', filter.sort_by);
	if (filter?.sort_order) params.set('sort_order', filter.sort_order);
	const qs = params.size ? '?' + params.toString() : '';
	return request<Ticket[]>(`/boards/${boardId}/tickets${qs}`);
};
export const createTicket = (boardId: string, data: Partial<Ticket> & { title: string }) =>
	request<Ticket>(`/boards/${boardId}/tickets`, { method: 'POST', body: JSON.stringify(data) });
export const getTicket = (id: string) => request<Ticket>(`/tickets/${id}`);
export const updateTicket = (
	id: string,
	fields: Partial<
		Pick<
			Ticket,
			| 'title'
			| 'description'
			| 'status'
			| 'priority'
			| 'epic_id'
			| 'assignee'
			| 'resolution'
		>
	>
) => request<Ticket>(`/tickets/${id}`, { method: 'PUT', body: JSON.stringify(fields) });
export const deleteTicket = (id: string) => request<void>(`/tickets/${id}`, { method: 'DELETE' });

// Tasks
export const listTasks = (ticketId: string) => request<Task[]>(`/tickets/${ticketId}/tasks`);
export const createTask = (ticketId: string, title: string) =>
	request<Task>(`/tickets/${ticketId}/tasks`, { method: 'POST', body: JSON.stringify({ title }) });
export const updateTask = (id: string, fields: { title?: string; done?: boolean }) =>
	request<Task>(`/tasks/${id}`, { method: 'PUT', body: JSON.stringify(fields) });
export const deleteTask = (id: string) => request<void>(`/tasks/${id}`, { method: 'DELETE' });

// Ticket Events (audit history)
export const listTicketEvents = (ticketId: string) =>
	request<TicketEvent[]>(`/tickets/${ticketId}/events`);

// Relations
export const listRelations = (ticketId: string) =>
	request<TicketRelation[]>(`/tickets/${ticketId}/relations`);
export const addRelation = (ticketId: string, toTicketId: string) =>
	request<TicketRelation>(`/tickets/${ticketId}/relations`, {
		method: 'POST',
		body: JSON.stringify({ to_ticket_id: toTicketId, kind: 'blocks' })
	});
export const deleteRelation = (ticketId: string, toTicketId: string) =>
	request<void>(`/tickets/${ticketId}/relations/${toTicketId}`, { method: 'DELETE' });

// Notes (agent scratchpad)
export const listNotes = (ticketId: string) => request<Note[]>(`/tickets/${ticketId}/notes`);
export const createNote = (ticketId: string, body: string) =>
	request<Note>(`/tickets/${ticketId}/notes`, { method: 'POST', body: JSON.stringify({ body }) });
export const updateNote = (id: string, body: string) =>
	request<Note>(`/notes/${id}`, { method: 'PUT', body: JSON.stringify({ body }) });
export const deleteNote = (id: string) => request<void>(`/notes/${id}`, { method: 'DELETE' });
