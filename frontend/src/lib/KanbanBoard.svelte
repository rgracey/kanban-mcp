<script lang="ts">
	import { toast } from 'svelte-sonner';
	import type { Ticket, Epic, Status, TicketFilter } from './types.js';
	import { listTickets, listEpics, updateTicket } from './api.js';
	import EpicFilter from './EpicFilter.svelte';
	import FilterBar from './FilterBar.svelte';
	import KanbanColumn from './KanbanColumn.svelte';
	import CreateTicket from './modals/CreateTicket.svelte';
	import CreateEpic from './modals/CreateEpic.svelte';
	import EditEpic from './modals/EditEpic.svelte';

	interface Props {
		boardId: string;
	}

	let { boardId }: Props = $props();

	let tickets = $state<Ticket[]>([]);
	let epics = $state<Epic[]>([]);
	let selectedEpicId = $state<string | null>(null);
	let ticketFilter = $state<TicketFilter>({});
	let loading = $state(true);
	let error = $state('');
	let showCreateTicket = $state(false);
	let showCreateEpic = $state(false);
	let editingEpic = $state<Epic | null>(null);

	const columns: { status: Status; label: string }[] = [
		{ status: 'todo', label: 'To Do' },
		{ status: 'in_progress', label: 'In Progress' },
		{ status: 'blocked', label: 'Blocked' },
		{ status: 'done', label: 'Done' }
	];

	async function load() {
		loading = true;
		error = '';
		try {
			const filter: TicketFilter = {
				...ticketFilter,
				...(selectedEpicId ? { epic_id: selectedEpicId } : {})
			};
			[tickets, epics] = await Promise.all([
				listTickets(boardId, filter),
				listEpics(boardId)
			]);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load board';
			toast.error(error);
		} finally {
			loading = false;
		}
	}

	// Reset filters and epic selection when board switches
	$effect(() => {
		// Tracking boardId — reset local filter state when board changes
		void boardId;
		ticketFilter = {};
		selectedEpicId = null;
	});

	$effect(() => {
		// Re-fetch whenever boardId, selectedEpicId, or ticketFilter changes.
		void boardId;
		void ticketFilter;
		void selectedEpicId;
		load();
	});

	// SSE: reload board when a change event arrives for this board.
	// The $effect returns a cleanup so the connection is torn down on re-run/destroy.
	$effect(() => {
		const currentBoardId = boardId;

		const source = new EventSource('/api/v1/events');

		const handler = (e: MessageEvent) => {
			try {
				const ev = JSON.parse(e.data) as { type: string; board_id: string };
				if (ev.board_id === currentBoardId) {
					load();
				}
			} catch {
				/* ignore parse errors */
			}
		};

		source.addEventListener('board_change', handler);

		// Return cleanup — runs when boardId changes or component is destroyed
		return () => {
			source.removeEventListener('board_change', handler);
			source.close();
		};
	});

	// Tickets are already filtered server-side; just split by status for columns
	function ticketsForStatus(status: Status): Ticket[] {
		return tickets.filter((t) => t.status === status);
	}

	function handleConsider(status: Status, items: Ticket[]) {
		const others = tickets.filter((t) => t.status !== status);
		tickets = [...others, ...items.map((t) => ({ ...t, status }))];
	}

	async function handleFinalize(status: Status, items: Ticket[]) {
		const others = tickets.filter((t) => t.status !== status);
		tickets = [...others, ...items.map((t) => ({ ...t, status }))];

		for (const t of items) {
			if (t.status !== status) {
				try {
					const saved = await updateTicket(t.id, { status });
					tickets = tickets.map((existing) => (existing.id === saved.id ? saved : existing));
					const label = { todo: 'To Do', in_progress: 'In Progress', done: 'Done', blocked: 'Blocked' }[status];
					toast.success(`Moved "${saved.title}" to ${label}`);
				} catch (e) {
					toast.error(e instanceof Error ? e.message : 'Failed to move ticket');
					await load();
					return;
				}
			}
		}
	}

	function handleTicketUpdate(updated: Ticket) {
		tickets = tickets.map((t) => (t.id === updated.id ? updated : t));
	}

	function handleTicketDelete(ticketId: string) {
		tickets = tickets.filter((t) => t.id !== ticketId);
	}

	function handleTicketCreate(ticket: Ticket) {
		tickets = [...tickets, ticket];
		showCreateTicket = false;
		toast.success(`Ticket "${ticket.title}" created`);
	}

	function handleEpicCreate(epic: Epic) {
		epics = [...epics, epic];
		showCreateEpic = false;
		toast.success(`Epic "${epic.title}" created`);
	}

	function handleEpicUpdate(updated: Epic) {
		epics = epics.map((e) => (e.id === updated.id ? updated : e));
		editingEpic = null;
	}

	function handleEpicDelete(epicId: string) {
		epics = epics.filter((e) => e.id !== epicId);
		if (selectedEpicId === epicId) selectedEpicId = null;
		editingEpic = null;
	}
</script>

<div class="flex h-full flex-col gap-4">
	<!-- Toolbar row 1: epic filter + action buttons -->
	<div class="flex flex-wrap items-center justify-between gap-3">
		<!-- Left: epic filter -->
		<div class="flex min-w-0 items-center gap-2">
			<span class="shrink-0 text-xs font-medium text-gray-400 dark:text-gray-500">Epic</span>
			<EpicFilter {epics} {selectedEpicId} onchange={(id) => (selectedEpicId = id)} onedit={(epic) => (editingEpic = epic)} />
		</div>
		<!-- Right: action buttons -->
		<div class="flex shrink-0 items-center gap-2">
			<button
				class="flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-sm font-medium text-gray-600 transition-colors hover:border-gray-300 hover:bg-gray-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300 dark:hover:border-gray-600 dark:hover:bg-gray-700/70"
				onclick={() => (showCreateEpic = true)}
			>
				<svg
					class="h-3.5 w-3.5"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					stroke-width="2.5"
				>
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
				</svg>
				Epic
			</button>
			<button
				class="flex items-center gap-1.5 rounded-lg bg-indigo-600 px-3 py-1.5 text-sm font-medium text-white shadow-sm transition-colors hover:bg-indigo-700"
				onclick={() => (showCreateTicket = true)}
			>
				<svg
					class="h-3.5 w-3.5"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					stroke-width="2.5"
				>
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
				</svg>
				New Ticket
			</button>
		</div>
	</div>

	<!-- Toolbar row 2: filter/search bar -->
	<FilterBar filter={ticketFilter} onchange={(f) => (ticketFilter = f)} />

	{#if loading}
		<div class="flex h-64 items-center justify-center text-sm text-gray-400">Loading...</div>
	{:else if error}
		<div class="flex h-64 items-center justify-center text-sm text-red-500">{error}</div>
	{:else}
		<div class="grid flex-1 grid-cols-4 gap-4">
			{#each columns as col (col.status)}
				<KanbanColumn
					status={col.status}
					label={col.label}
					tickets={ticketsForStatus(col.status)}
					{epics}
					{boardId}
					onconsider={handleConsider}
					onfinalize={handleFinalize}
					onticketupdate={handleTicketUpdate}
					onticketdelete={handleTicketDelete}
				/>
			{/each}
		</div>
	{/if}
</div>

{#if showCreateTicket}
	<CreateTicket
		{boardId}
		onclose={() => (showCreateTicket = false)}
		oncreated={handleTicketCreate}
	/>
{/if}

{#if showCreateEpic}
	<CreateEpic {boardId} onclose={() => (showCreateEpic = false)} oncreated={handleEpicCreate} />
{/if}

{#if editingEpic}
	<EditEpic
		epic={editingEpic}
		onclose={() => (editingEpic = null)}
		onupdated={handleEpicUpdate}
		ondeleted={handleEpicDelete}
	/>
{/if}
