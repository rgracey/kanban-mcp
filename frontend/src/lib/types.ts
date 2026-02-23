export type Status = 'todo' | 'in_progress' | 'done';
export type Priority = 'low' | 'medium' | 'high' | 'critical';

export interface Board {
	id: string;
	name: string;
	description: string;
	created_at: string;
	updated_at: string;
}

export interface Epic {
	id: string;
	board_id: string;
	title: string;
	description: string;
	created_at: string;
	updated_at: string;
}

export interface TicketResolution {
	commit_sha?: string;
	pr_url?: string;
	notes?: string;
	resolved_at?: string;
}

export interface Ticket {
	id: string;
	board_id: string;
	epic_id: string | null;
	title: string;
	description: string;
	status: Status;
	priority: Priority;
	assignee: string;
	resolution?: TicketResolution | null;
	created_at: string;
	updated_at: string;
}

export interface Note {
	id: string;
	ticket_id: string;
	body: string;
	created_at: string;
	updated_at: string;
}

export interface Task {
	id: string;
	ticket_id: string;
	title: string;
	done: boolean;
	position: number;
	created_at: string;
	updated_at: string;
}

export type RelationKind = 'blocks';

export interface TicketRelation {
	from_ticket_id: string;
	to_ticket_id: string;
	kind: RelationKind;
	created_at: string;
	from_title: string;
	to_title: string;
}

export type TicketEventType =
	| 'created'
	| 'moved'
	| 'edited'
	| 'commented'
	| 'comment_edited'
	| 'task_added'
	| 'task_updated'
	| 'task_deleted'
	| 'relation_added'
	| 'relation_removed'
	| string;

export interface TicketEvent {
	id: string;
	ticket_id: string;
	type: TicketEventType;
	actor: string;
	payload: Record<string, unknown>;
	created_at: string;
}

export interface BoardSummary {
	board_id: string;
	ticket_counts: Record<Status, number>;
	epics: { id: string; title: string; ticket_count: number }[];
}

export interface TicketFilter {
	status?: Status;
	priority?: Priority;
	epic_id?: string;
	q?: string;
}
