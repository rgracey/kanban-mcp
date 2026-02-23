<script lang="ts">
	import type {
		Ticket,
		Epic,
		Note,
		Task,
		TicketEvent,
		TicketRelation,
		TicketReference,
		TicketResolution
	} from './types.js';
	import {
		getTicket,
		updateTicket,
		deleteTicket,
		listEpics,
		listNotes,
		createNote,
		updateNote,
		deleteNote,
		listTasks,
		createTask,
		updateTask,
		deleteTask,
		listTicketEvents,
		listTickets,
		listRelations,
		addRelation,
		deleteRelation
	} from './api.js';
	import { toast } from 'svelte-sonner';
	import { marked } from 'marked';
	import { untrack } from 'svelte';
	import CopyId from './CopyId.svelte';

	interface Props {
		ticketId: string;
		boardId: string;
		onclose: () => void;
		onupdate?: (ticket: Ticket) => void;
		ondelete?: (ticketId: string) => void;
	}

	let { ticketId, boardId, onclose, onupdate, ondelete }: Props = $props();

	// --- server data ---
	let ticket = $state<Ticket | null>(null);
	let epics = $state<Epic[]>([]);
	let notes = $state<Note[]>([]);
	let tasks = $state<Task[]>([]);
	let events = $state<TicketEvent[]>([]);
	let relations = $state<TicketRelation[]>([]);
	let boardTickets = $state<Ticket[]>([]);
	let loading = $state(true);
	let loadError = $state('');

	// --- history panel ---
	let showHistory = $state(false);

	// --- draft state (all fields go through Save) ---
	let draftTitle = $state('');
	let draftDescription = $state('');
	let draftAssignee = $state('');
	let draftStatus = $state('');
	let draftPriority = $state('');
	let draftEpicId = $state('');
	let draftReferences = $state<TicketReference[]>([]);
	let draftResolution = $state<TicketResolution | null>(null);
	let editingDescription = $state(false);

	// --- reference editing ---
	let newRefKind = $state<TicketReference['kind']>('file');
	let newRefTarget = $state('');
	let newRefLabel = $state('');

	// --- pending new tasks (saved on Save button) ---
	let pendingTasks = $state<string[]>([]); // titles not yet persisted
	let newTaskTitle = $state('');

	// --- save / delete state ---
	let saving = $state(false);
	let fieldError = $state('');
	let deletingTicket = $state(false);

	// --- relation state ---
	let newRelationId = $state('');
	let newRelationDirection = $state<'blocks' | 'blocked_by'>('blocks');
	let addingRelation = $state(false);

	const relatedIds = $derived(
		new Set([...relations.map((r) => r.from_ticket_id), ...relations.map((r) => r.to_ticket_id)])
	);
	const selectableTickets = $derived(
		boardTickets.filter((t) => t.id !== ticketId && !relatedIds.has(t.id))
	);

	// --- task progress ---
	const taskDoneCount = $derived(tasks.filter((t) => t.done).length);
	const taskTotal = $derived(tasks.length + pendingTasks.length);
	const taskProgress = $derived(
		taskTotal > 0 ? Math.round((taskDoneCount / tasks.length) * 100) : 0
	);

	// --- note state ---
	let newNoteBody = $state('');
	let addingNote = $state(false);
	let editingNoteId = $state<string | null>(null);
	let editingNoteBody = $state('');
	let savingNote = $state(false);

	async function load() {
		loading = true;
		loadError = '';
		try {
			const [t, e, n, tk, ev, rels, bt] = await Promise.all([
				getTicket(ticketId),
				listEpics(boardId),
				listNotes(ticketId),
				listTasks(ticketId),
				listTicketEvents(ticketId),
				listRelations(ticketId),
				listTickets(boardId)
			]);
			ticket = t;
			epics = e;
			notes = n;
			tasks = tk;
			events = ev;
			relations = rels;
			boardTickets = bt;
			pendingTasks = [];
			draftTitle = untrack(() => t.title);
			draftDescription = untrack(() => t.description ?? '');
			draftAssignee = untrack(() => t.assignee ?? '');
			draftStatus = untrack(() => t.status);
			draftPriority = untrack(() => t.priority);
			draftEpicId = untrack(() => t.epic_id ?? '');
			draftReferences = untrack(() => [...(t.references ?? [])]);
			draftResolution = untrack(() => (t.resolution ? { ...t.resolution } : null));
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load ticket';
			toast.error(loadError);
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		ticketId;
		load();
	});

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && !editingDescription) onclose();
	}

	const hasDraftChanges = $derived(
		ticket !== null &&
			(draftTitle !== ticket.title ||
				draftDescription !== (ticket.description ?? '') ||
				draftAssignee !== (ticket.assignee ?? '') ||
				draftStatus !== ticket.status ||
				draftPriority !== ticket.priority ||
				draftEpicId !== (ticket.epic_id ?? '') ||
				JSON.stringify(draftReferences) !== JSON.stringify(ticket.references ?? []) ||
				JSON.stringify(draftResolution) !== JSON.stringify(ticket.resolution ?? null) ||
				pendingTasks.length > 0)
	);

	async function saveAll() {
		if (!ticket) return;
		saving = true;
		fieldError = '';
		try {
			const patch: Parameters<typeof updateTicket>[1] = {};
			const trimmedTitle = draftTitle.trim();
			if (trimmedTitle && trimmedTitle !== ticket.title) patch.title = trimmedTitle;
			if (draftDescription !== (ticket.description ?? '')) patch.description = draftDescription;
			if (draftAssignee !== (ticket.assignee ?? '')) patch.assignee = draftAssignee;
			if (draftStatus !== ticket.status) patch.status = draftStatus as Ticket['status'];
			if (draftPriority !== ticket.priority) patch.priority = draftPriority as Ticket['priority'];
			if ((draftEpicId || null) !== ticket.epic_id) patch.epic_id = draftEpicId || null;
			if (JSON.stringify(draftReferences) !== JSON.stringify(ticket.references ?? []))
				patch.references = draftReferences;
			if (JSON.stringify(draftResolution) !== JSON.stringify(ticket.resolution ?? null))
				patch.resolution = draftResolution;

			const updated = Object.keys(patch).length > 0 ? await updateTicket(ticket.id, patch) : ticket;
			ticket = updated;
			draftTitle = updated.title;
			draftDescription = updated.description ?? '';
			draftAssignee = updated.assignee ?? '';
			draftStatus = updated.status;
			draftPriority = updated.priority;
			draftEpicId = updated.epic_id ?? '';
			draftReferences = [...(updated.references ?? [])];
			draftResolution = updated.resolution ? { ...updated.resolution } : null;
			editingDescription = false;
			onupdate?.(updated);

			if (pendingTasks.length > 0) {
				const created = await Promise.all(
					pendingTasks.map((title) => createTask(ticket!.id, title))
				);
				tasks = [...tasks, ...created];
				pendingTasks = [];
			}

			toast.success('Saved');
		} catch (e) {
			const msg = e instanceof Error ? e.message : 'Save failed';
			fieldError = msg;
			toast.error(msg);
		} finally {
			saving = false;
		}
	}

	async function toggleTask(task: Task) {
		try {
			const updated = await updateTask(task.id, { done: !task.done });
			tasks = tasks.map((t) => (t.id === updated.id ? updated : t));
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to update task');
		}
	}

	async function removeTask(id: string) {
		try {
			await deleteTask(id);
			tasks = tasks.filter((t) => t.id !== id);
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to delete task');
		}
	}

	function addPendingTask() {
		const title = newTaskTitle.trim();
		if (!title) return;
		pendingTasks = [...pendingTasks, title];
		newTaskTitle = '';
	}

	function removePendingTask(i: number) {
		pendingTasks = pendingTasks.filter((_, idx) => idx !== i);
	}

	function onNewTaskKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') addPendingTask();
	}

	async function submitRelation() {
		const otherId = newRelationId.trim();
		if (!otherId || !ticket) return;
		addingRelation = true;
		try {
			const [fromId, toId] =
				newRelationDirection === 'blocks' ? [ticket.id, otherId] : [otherId, ticket.id];
			const rel = await addRelation(fromId, toId);
			relations = [...relations, rel];
			newRelationId = '';
			toast.success('Relation added');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to add relation');
		} finally {
			addingRelation = false;
		}
	}

	async function removeRelation(rel: TicketRelation) {
		if (!ticket) return;
		try {
			await deleteRelation(rel.from_ticket_id, rel.to_ticket_id);
			relations = relations.filter(
				(r) => !(r.from_ticket_id === rel.from_ticket_id && r.to_ticket_id === rel.to_ticket_id)
			);
			toast.success('Relation removed');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to remove relation');
		}
	}

	async function submitNote() {
		if (!newNoteBody.trim() || !ticket) return;
		addingNote = true;
		try {
			const note = await createNote(ticket.id, newNoteBody.trim());
			notes = [...notes, note];
			newNoteBody = '';
			toast.success('Note posted');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to post note');
		} finally {
			addingNote = false;
		}
	}

	function startEditNote(note: Note) {
		editingNoteId = note.id;
		editingNoteBody = note.body;
	}

	async function saveNoteEdit() {
		if (!editingNoteId) return;
		savingNote = true;
		try {
			const updated = await updateNote(editingNoteId, editingNoteBody);
			notes = notes.map((n) => (n.id === updated.id ? updated : n));
			editingNoteId = null;
			toast.success('Note updated');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to update note');
		} finally {
			savingNote = false;
		}
	}

	async function removeNote(id: string) {
		try {
			await deleteNote(id);
			notes = notes.filter((n) => n.id !== id);
			toast.success('Note deleted');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to delete note');
		}
	}

	async function removeTicket() {
		if (!ticket || !confirm('Delete this ticket? This cannot be undone.')) return;
		deletingTicket = true;
		try {
			await deleteTicket(ticket.id);
			toast.success('Ticket deleted');
			ondelete?.(ticket.id);
			onclose();
		} catch (e) {
			const msg = e instanceof Error ? e.message : 'Delete failed';
			fieldError = msg;
			toast.error(msg);
		} finally {
			deletingTicket = false;
		}
	}

	function formatDate(iso: string) {
		return new Date(iso).toLocaleString(undefined, {
			dateStyle: 'medium',
			timeStyle: 'short'
		});
	}

	const eventIcons: Record<string, string> = {
		created: '✦',
		moved: '→',
		edited: '✎',
		commented: '📝',
		comment_edited: '📝',
		task_added: '☐',
		task_updated: '☑',
		task_deleted: '✕',
		relation_added: '⛓',
		relation_removed: '✂'
	};

	function eventDescription(ev: TicketEvent): string {
		const p = ev.payload ?? {};
		switch (ev.type) {
			case 'created':
				return 'Ticket created';
			case 'moved': {
				const from = (p.from as string | undefined)?.replace(/_/g, ' ');
				const to = (p.to as string | undefined)?.replace(/_/g, ' ');
				return from && to ? `Moved from ${from} → ${to}` : 'Status changed';
			}
			case 'edited': {
				const fieldLabels: Record<string, string> = {
					title: 'title',
					description: 'description',
					priority: 'priority',
					assignee: 'assignee',
					epic_id: 'epic'
				};
				const parts: string[] = [];
				for (const [k, v] of Object.entries(p)) {
					const label = fieldLabels[k] ?? k;
					if (k === 'epic_id') {
						parts.push(v == null ? 'Epic cleared' : 'Epic set');
					} else if (k === 'priority' || k === 'status') {
						parts.push(
							`${label.charAt(0).toUpperCase() + label.slice(1)} set to ${String(v).replace(/_/g, ' ')}`
						);
					} else if (k === 'assignee') {
						parts.push(v ? `Assigned to ${v}` : 'Assignee cleared');
					} else if (k === 'title') {
						parts.push('Title updated');
					} else if (k === 'description') {
						parts.push('Description updated');
					} else {
						parts.push(`${label} updated`);
					}
				}
				return parts.length ? parts.join('; ') : 'Edited';
			}
			case 'commented':
				return 'Note added';
			case 'comment_edited':
				return 'Note edited';
			case 'task_added':
				return p.task_title ? `Task added: "${p.task_title}"` : 'Task added';
			case 'task_updated': {
				const title = p.task_title as string | undefined;
				const done = p.done as boolean | undefined;
				if (done !== undefined) return `Task "${title}" marked ${done ? 'done' : 'not done'}`;
				if (p.title) return `Task renamed to "${p.title}"`;
				return title ? `Task "${title}" updated` : 'Task updated';
			}
			case 'task_deleted':
				return p.task_title ? `Task deleted: "${p.task_title}"` : 'Task deleted';
			case 'relation_added': {
				const title = p.related_title as string | undefined;
				const dir = p.direction as string | undefined;
				if (dir === 'outgoing')
					return title ? `Now blocking "${title}"` : 'Blocking relation added';
				return title ? `Blocked by "${title}"` : 'Blocked-by relation added';
			}
			case 'relation_removed': {
				const title = p.related_title as string | undefined;
				const dir = p.direction as string | undefined;
				if (dir === 'outgoing')
					return title ? `No longer blocking "${title}"` : 'Blocking relation removed';
				return title ? `No longer blocked by "${title}"` : 'Blocked-by relation removed';
			}
			default:
				return ev.type;
		}
	}

	// shared input / label classes
	const inputCls =
		'w-full border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded-lg px-3 py-2 text-sm placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-shadow';
	const labelCls =
		'block text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-1.5';
	const sectionCls = 'border-t border-gray-100 dark:border-gray-800 pt-4';
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- Backdrop -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<div
	role="presentation"
	class="fixed inset-0 z-40 bg-black/40 backdrop-blur-[2px] dark:bg-black/60"
	onclick={onclose}
></div>

<!-- Slide-over panel -->
<div
	role="dialog"
	aria-modal="true"
	tabindex="-1"
	class="fixed top-0 right-0 z-50 flex h-full w-full max-w-lg flex-col overflow-hidden border-l border-gray-200 bg-white shadow-2xl dark:border-gray-800 dark:bg-gray-900"
>
	<!-- Header -->
	<div
		class="flex shrink-0 items-center justify-between border-b border-gray-200 bg-white px-5 py-3.5 dark:border-gray-800 dark:bg-gray-900"
	>
		<div class="flex items-center gap-2">
			<h2 class="text-sm font-semibold text-gray-900 dark:text-gray-100">Ticket Detail</h2>
			{#if ticket}
				<CopyId id={ticket.id} />
			{/if}
		</div>
		<button
			class="flex h-7 w-7 items-center justify-center rounded-lg text-lg leading-none text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-800 dark:hover:text-gray-300"
			onclick={onclose}
			aria-label="Close">&times;</button
		>
	</div>

	{#if loading}
		<div class="flex flex-1 items-center justify-center gap-2 text-sm text-gray-400">
			<svg class="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"
				></circle>
				<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
				></path>
			</svg>
			Loading...
		</div>
	{:else if loadError}
		<div class="flex flex-1 items-center justify-center text-sm text-red-500">{loadError}</div>
	{:else if ticket}
		<div class="flex-1 overflow-y-auto">
			<div class="space-y-4 px-5 py-4">
				{#if fieldError}
					<p
						class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-500 dark:border-red-800 dark:bg-red-950/30"
					>
						{fieldError}
					</p>
				{/if}

				<!-- Title -->
				<div>
					<label for="td-title" class={labelCls}>Title</label>
					<input id="td-title" class={inputCls} bind:value={draftTitle} />
				</div>

				<!-- Status / Priority / Epic (3-col grid) -->
				<div class="grid grid-cols-3 gap-3">
					<div>
						<label for="td-status" class={labelCls}>Status</label>
						<select id="td-status" class={inputCls} bind:value={draftStatus}>
							<option value="todo">To Do</option>
							<option value="in_progress">In Progress</option>
							<option value="done">Done</option>
						</select>
					</div>
					<div>
						<label for="td-priority" class={labelCls}>Priority</label>
						<select id="td-priority" class={inputCls} bind:value={draftPriority}>
							<option value="low">Low</option>
							<option value="medium">Medium</option>
							<option value="high">High</option>
							<option value="critical">Critical</option>
						</select>
					</div>
					<div>
						<label for="td-epic" class={labelCls}>Epic</label>
						<select id="td-epic" class={inputCls} bind:value={draftEpicId}>
							<option value="">None</option>
							{#each epics as epic (epic.id)}
								<option value={epic.id}>{epic.title}</option>
							{/each}
						</select>
					</div>
				</div>

				<!-- Assignee -->
				<div>
					<label for="td-assignee" class={labelCls}>Assignee</label>
					<input
						id="td-assignee"
						class={inputCls}
						placeholder="Unassigned"
						bind:value={draftAssignee}
					/>
				</div>

				<!-- Code References -->
				<div class={sectionCls}>
					<h3 class={labelCls}>Code References</h3>

					{#if draftReferences.length > 0}
						<ul class="mb-2 space-y-1.5">
							{#each draftReferences as ref, i}
								<li class="group flex items-center gap-2 text-sm">
									<span
										class="shrink-0 rounded-md px-1.5 py-0.5 text-[11px] font-semibold
                    {ref.kind === 'file'
											? 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400'
											: ref.kind === 'pr'
												? 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300'
												: ref.kind === 'commit'
													? 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-400'
													: 'bg-sky-100 text-sky-700 dark:bg-sky-900/40 dark:text-sky-400'}"
									>
										{ref.kind}
									</span>
									{#if ref.kind === 'file'}
										<code class="flex-1 truncate font-mono text-xs text-gray-700 dark:text-gray-300"
											>{ref.target}</code
										>
									{:else}
										<a
											href={ref.target}
											target="_blank"
											rel="noopener noreferrer"
											class="flex-1 truncate text-xs text-indigo-600 hover:underline dark:text-indigo-400"
											>{ref.label || ref.target}</a
										>
									{/if}
									{#if ref.label && ref.kind === 'file'}
										<span class="shrink-0 text-xs text-gray-400">{ref.label}</span>
									{/if}
									<button
										class="shrink-0 text-gray-300 opacity-0 transition-opacity group-hover:opacity-100 hover:text-red-500 dark:text-gray-600"
										onclick={() => {
											draftReferences = draftReferences.filter((_, idx) => idx !== i);
										}}
										aria-label="Remove reference"
									>
										<svg
											class="h-3.5 w-3.5"
											fill="none"
											viewBox="0 0 24 24"
											stroke="currentColor"
											stroke-width="2"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												d="M6 18L18 6M6 6l12 12"
											/>
										</svg>
									</button>
								</li>
							{/each}
						</ul>
					{:else}
						<p class="mb-2 text-xs text-gray-400 italic dark:text-gray-500">No references.</p>
					{/if}

					<!-- Add reference row -->
					<div class="flex flex-wrap gap-2">
						<select
							class="shrink-0 rounded-lg border border-gray-200 bg-white px-2 py-1.5 text-xs text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100"
							bind:value={newRefKind}
						>
							<option value="file">file</option>
							<option value="url">url</option>
							<option value="pr">pr</option>
							<option value="commit">commit</option>
						</select>
						<input
							class="min-w-0 flex-1 rounded-lg border border-gray-200 bg-white px-2 py-1.5 font-mono text-xs text-gray-900 placeholder-gray-400 focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder-gray-500"
							placeholder={newRefKind === 'file' ? 'src/handler.go:42' : 'https://...'}
							bind:value={newRefTarget}
							onkeydown={(e) => {
								if (e.key === 'Enter' && newRefTarget.trim()) {
									draftReferences = [
										...draftReferences,
										{
											kind: newRefKind,
											target: newRefTarget.trim(),
											label: newRefLabel.trim() || undefined
										}
									];
									newRefTarget = '';
									newRefLabel = '';
								}
							}}
						/>
						<input
							class="w-28 rounded-lg border border-gray-200 bg-white px-2 py-1.5 text-xs text-gray-900 placeholder-gray-400 focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder-gray-500"
							placeholder="label (opt)"
							bind:value={newRefLabel}
						/>
						<button
							class="rounded-lg bg-gray-100 px-3 py-1.5 text-xs font-medium text-gray-700 transition-colors hover:bg-gray-200 disabled:opacity-40 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700"
							onclick={() => {
								if (!newRefTarget.trim()) return;
								draftReferences = [
									...draftReferences,
									{
										kind: newRefKind,
										target: newRefTarget.trim(),
										label: newRefLabel.trim() || undefined
									}
								];
								newRefTarget = '';
								newRefLabel = '';
							}}
							disabled={!newRefTarget.trim()}>Add</button
						>
					</div>
				</div>

				<!-- Resolution -->
				<div class={sectionCls}>
					<div class="mb-2 flex items-center justify-between">
						<h3 class={labelCls} style="margin-bottom:0">Resolution</h3>
						{#if !draftResolution}
							<button
								class="text-xs font-medium text-indigo-500 transition-colors hover:text-indigo-700 dark:text-indigo-400 dark:hover:text-indigo-300"
								onclick={() => {
									draftResolution = { commit_sha: '', pr_url: '', notes: '', resolved_at: '' };
								}}>+ Record resolution</button
							>
						{:else}
							<button
								class="text-xs font-medium text-red-400 transition-colors hover:text-red-600"
								onclick={() => {
									draftResolution = null;
								}}>Clear</button
							>
						{/if}
					</div>

					{#if draftResolution}
						<div
							class="space-y-2 rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-gray-700 dark:bg-gray-800/50"
						>
							<div class="grid grid-cols-2 gap-2">
								<div>
									<label
										for="res-commit"
										class="mb-1 block text-[10px] font-semibold tracking-wide text-gray-400 uppercase dark:text-gray-500"
										>Commit SHA</label
									>
									<input
										id="res-commit"
										class="w-full rounded-lg border border-gray-200 bg-white px-2 py-1.5 font-mono text-xs text-gray-900 placeholder-gray-400 focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder-gray-500"
										placeholder="abc1234"
										bind:value={draftResolution.commit_sha}
									/>
								</div>
								<div>
									<label
										for="res-pr"
										class="mb-1 block text-[10px] font-semibold tracking-wide text-gray-400 uppercase dark:text-gray-500"
										>PR URL</label
									>
									<input
										id="res-pr"
										class="w-full rounded-lg border border-gray-200 bg-white px-2 py-1.5 text-xs text-gray-900 placeholder-gray-400 focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder-gray-500"
										placeholder="https://github.com/..."
										bind:value={draftResolution.pr_url}
									/>
								</div>
							</div>
							<div>
								<label
									for="res-notes"
									class="mb-1 block text-[10px] font-semibold tracking-wide text-gray-400 uppercase dark:text-gray-500"
									>Notes</label
								>
								<textarea
									id="res-notes"
									class="w-full resize-none rounded-lg border border-gray-200 bg-white px-2 py-1.5 text-xs text-gray-900 placeholder-gray-400 focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder-gray-500"
									rows="2"
									placeholder="What was done, how it was fixed…"
									bind:value={draftResolution.notes}
								></textarea>
							</div>
							<div>
								<label
									for="res-resolved-at"
									class="mb-1 block text-[10px] font-semibold tracking-wide text-gray-400 uppercase dark:text-gray-500"
									>Resolved at (RFC3339)</label
								>
								<input
									id="res-resolved-at"
									class="w-full rounded-lg border border-gray-200 bg-white px-2 py-1.5 font-mono text-xs text-gray-900 placeholder-gray-400 focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder-gray-500"
									placeholder="2026-01-01T00:00:00Z"
									bind:value={draftResolution.resolved_at}
								/>
							</div>
						</div>
					{:else}
						<p class="text-xs text-gray-400 italic dark:text-gray-500">No resolution recorded.</p>
					{/if}
				</div>

				<!-- Description -->
				<div class={sectionCls}>
					<div class="mb-1.5 flex items-center justify-between">
						<label for="td-desc" class={labelCls} style="margin-bottom:0">Description</label>
						{#if !editingDescription}
							<button
								class="text-xs font-medium text-indigo-500 transition-colors hover:text-indigo-700 dark:text-indigo-400 dark:hover:text-indigo-300"
								onclick={() => (editingDescription = true)}>Edit</button
							>
						{:else}
							<button
								class="text-xs font-medium text-gray-400 transition-colors hover:text-gray-600 dark:hover:text-gray-300"
								onclick={() => (editingDescription = false)}>Done</button
							>
						{/if}
					</div>
					{#if editingDescription}
						<textarea
							id="td-desc"
							class={inputCls + ' resize-none'}
							rows="6"
							bind:value={draftDescription}
						></textarea>
						<p class="mt-1 text-xs text-gray-400">Markdown supported</p>
					{:else}
						<!-- svelte-ignore a11y_click_events_have_key_events -->
						<div
							role="button"
							tabindex="0"
							class="md-body min-h-[3rem] cursor-text rounded-lg border border-transparent px-3 py-2 transition-colors hover:border-gray-200 dark:text-gray-300 dark:hover:border-gray-700 {!draftDescription
								? 'text-sm text-gray-400 italic dark:text-gray-500'
								: ''}"
							onclick={() => (editingDescription = true)}
						>
							{#if draftDescription}
								{@html marked.parse(draftDescription)}
							{:else}
								<span>No description — click to add</span>
							{/if}
						</div>
					{/if}
				</div>

				<!-- Tasks / Checklist -->
				<div class={sectionCls}>
					<div class="mb-2 flex items-center justify-between">
						<h3 class={labelCls} style="margin-bottom:0">
							Tasks
							{#if taskTotal > 0}
								<span class="ml-1 font-normal text-gray-400 normal-case"
									>({taskDoneCount}/{taskTotal})</span
								>
							{/if}
						</h3>
						<!-- Progress bar -->
						{#if tasks.length > 0}
							<div class="flex items-center gap-2">
								<div class="h-1.5 w-20 overflow-hidden rounded-full bg-gray-200 dark:bg-gray-700">
									<div
										class="h-full rounded-full bg-indigo-500 transition-all"
										style="width: {taskProgress}%"
									></div>
								</div>
								<span class="text-xs text-gray-400">{taskProgress}%</span>
							</div>
						{/if}
					</div>

					{#if tasks.length > 0}
						<ul class="mb-2 space-y-1.5">
							{#each tasks as task (task.id)}
								<li class="group flex items-center gap-2.5">
									<input
										type="checkbox"
										class="h-4 w-4 shrink-0 cursor-pointer rounded border-gray-300 text-indigo-600 focus:ring-indigo-500 dark:border-gray-600"
										checked={task.done}
										onchange={() => toggleTask(task)}
									/>
									<span
										class="flex-1 text-sm {task.done
											? 'text-gray-400 line-through dark:text-gray-600'
											: 'text-gray-800 dark:text-gray-200'}">{task.title}</span
									>
									<button
										class="text-gray-300 opacity-0 transition-opacity group-hover:opacity-100 hover:text-red-500 dark:text-gray-600"
										onclick={() => removeTask(task.id)}
										aria-label="Delete task"
									>
										<svg
											class="h-3.5 w-3.5"
											fill="none"
											viewBox="0 0 24 24"
											stroke="currentColor"
											stroke-width="2"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												d="M6 18L18 6M6 6l12 12"
											/>
										</svg>
									</button>
								</li>
							{/each}
						</ul>
					{/if}

					{#if pendingTasks.length > 0}
						<ul class="mb-2 space-y-1.5">
							{#each pendingTasks as title, i}
								<li class="group flex items-center gap-2.5">
									<input
										type="checkbox"
										class="h-4 w-4 shrink-0 cursor-not-allowed rounded border-gray-300 opacity-40 dark:border-gray-600"
										disabled
									/>
									<span class="flex-1 text-sm text-gray-400 italic dark:text-gray-500">{title}</span
									>
									<span
										class="rounded bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-400 dark:bg-gray-800"
										>unsaved</span
									>
									<button
										class="text-gray-300 opacity-0 transition-opacity group-hover:opacity-100 hover:text-red-500 dark:text-gray-600"
										onclick={() => removePendingTask(i)}
										aria-label="Remove pending task"
									>
										<svg
											class="h-3.5 w-3.5"
											fill="none"
											viewBox="0 0 24 24"
											stroke="currentColor"
											stroke-width="2"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												d="M6 18L18 6M6 6l12 12"
											/>
										</svg>
									</button>
								</li>
							{/each}
						</ul>
					{/if}

					<div class="mt-2 flex gap-2">
						<input
							class={inputCls + ' flex-1'}
							placeholder="Add a task…"
							bind:value={newTaskTitle}
							onkeydown={onNewTaskKeydown}
						/>
						<button
							class="rounded-lg bg-gray-100 px-3 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-200 disabled:opacity-40 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700"
							onclick={addPendingTask}
							disabled={!newTaskTitle.trim()}>Add</button
						>
					</div>
				</div>

				<!-- Relations -->
				<div class={sectionCls}>
					<h3 class={labelCls}>
						Relations
						{#if relations.length > 0}
							<span class="ml-1 font-normal text-gray-400 normal-case">({relations.length})</span>
						{/if}
					</h3>

					{#if relations.length > 0}
						<ul class="mb-3 space-y-1.5">
							{#each relations as rel}
								{@const isBlocking = rel.from_ticket_id === ticket?.id}
								<li class="group flex items-center gap-2 text-sm">
									<span
										class="shrink-0 rounded-md px-1.5 py-0.5 text-[11px] font-semibold
                    {isBlocking
											? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
											: 'bg-sky-100 text-sky-700 dark:bg-sky-900/30 dark:text-sky-400'}"
									>
										{isBlocking ? 'blocks' : 'blocked by'}
									</span>
									<span class="flex-1 truncate text-sm text-gray-700 dark:text-gray-300">
										{isBlocking ? rel.to_title : rel.from_title}
									</span>
									<button
										class="shrink-0 text-gray-300 opacity-0 transition-opacity group-hover:opacity-100 hover:text-red-500 dark:text-gray-600"
										onclick={() => removeRelation(rel)}
										aria-label="Remove relation"
									>
										<svg
											class="h-3.5 w-3.5"
											fill="none"
											viewBox="0 0 24 24"
											stroke="currentColor"
											stroke-width="2"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												d="M6 18L18 6M6 6l12 12"
											/>
										</svg>
									</button>
								</li>
							{/each}
						</ul>
					{:else}
						<p class="mb-3 text-sm text-gray-400 italic dark:text-gray-500">No relations.</p>
					{/if}

					<div class="flex gap-2">
						<select
							class="shrink-0 rounded-lg border border-gray-200 bg-white px-2 py-2 text-sm text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100"
							bind:value={newRelationDirection}
						>
							<option value="blocks">blocks</option>
							<option value="blocked_by">blocked by</option>
						</select>
						<select
							class="flex-1 rounded-lg border border-gray-200 bg-white px-2 py-2 text-sm text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:outline-none disabled:opacity-50 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100"
							bind:value={newRelationId}
							disabled={selectableTickets.length === 0}
						>
							<option value=""
								>{selectableTickets.length === 0 ? 'No other tickets' : 'Select a ticket…'}</option
							>
							{#each selectableTickets as t (t.id)}
								<option value={t.id}>{t.title}</option>
							{/each}
						</select>
						<button
							class="rounded-lg bg-gray-100 px-3 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-200 disabled:opacity-40 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700"
							onclick={submitRelation}
							disabled={addingRelation || !newRelationId}>Add</button
						>
					</div>
				</div>

				<!-- History (collapsible) -->
				<div class={sectionCls}>
					<button
						class="group mb-2 flex w-full items-center gap-1.5"
						onclick={() => (showHistory = !showHistory)}
						aria-expanded={showHistory}
					>
						<svg
							class="h-3 w-3 text-gray-400 transition-transform {showHistory ? 'rotate-90' : ''}"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							stroke-width="3"
						>
							<path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
						</svg>
						<span
							class={labelCls +
								' cursor-pointer transition-colors group-hover:text-gray-600 dark:group-hover:text-gray-300'}
							style="margin-bottom:0"
						>
							History
							{#if events.length > 0}
								<span class="ml-1 font-normal text-gray-400 normal-case">({events.length})</span>
							{/if}
						</span>
					</button>

					{#if showHistory}
						{#if events.length === 0}
							<p class="mb-2 text-sm text-gray-400 italic dark:text-gray-500">No history yet.</p>
						{:else}
							<ol
								class="relative mb-2 ml-2 space-y-3 border-l border-gray-200 dark:border-gray-700"
							>
								{#each [...events].reverse() as ev (ev.id)}
									<li class="ml-4">
										<span
											class="absolute -left-2 flex h-4 w-4 items-center justify-center rounded-full bg-gray-100 text-xs ring-2 ring-white dark:bg-gray-800 dark:ring-gray-900"
										>
											{eventIcons[ev.type] ?? '·'}
										</span>
										<div class="flex flex-col">
											<span class="text-xs font-medium text-gray-700 dark:text-gray-300">
												{eventDescription(ev)}
												{#if ev.actor}
													<span class="font-normal text-gray-500 dark:text-gray-400">
														by {ev.actor}</span
													>
												{/if}
											</span>
											<time class="text-xs text-gray-400 dark:text-gray-500"
												>{formatDate(ev.created_at)}</time
											>
										</div>
									</li>
								{/each}
							</ol>
						{/if}
					{/if}
				</div>

				<!-- Agent Notes (scratchpad) -->
				<div class={sectionCls}>
					<div class="mb-2 flex items-center justify-between">
						<h3 class={labelCls} style="margin-bottom:0">
							Agent Notes
							{#if notes.length > 0}
								<span class="ml-1 font-normal text-gray-400 normal-case">({notes.length})</span>
							{/if}
						</h3>
						<span class="text-[10px] text-gray-400 italic dark:text-gray-500"
							>scratchpad · visible to agents</span
						>
					</div>

					{#if notes.length > 0}
						<div class="mb-4 space-y-2.5">
							{#each notes as note (note.id)}
								<div
									class="rounded-lg border border-gray-100 bg-gray-50 p-3 dark:border-gray-700/60 dark:bg-gray-800/60"
								>
									{#if editingNoteId === note.id}
										<textarea
											class="mb-2 w-full resize-none rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100"
											rows="3"
											bind:value={editingNoteBody}
										></textarea>
										<div class="flex gap-2">
											<button
												class="rounded-lg bg-indigo-600 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-indigo-700 disabled:opacity-50"
												onclick={saveNoteEdit}
												disabled={savingNote}>{savingNote ? 'Saving...' : 'Save'}</button
											>
											<button
												class="rounded-lg px-3 py-1.5 text-xs font-medium text-gray-600 transition-colors hover:bg-gray-200 dark:text-gray-400 dark:hover:bg-gray-700"
												onclick={() => (editingNoteId = null)}>Cancel</button
											>
										</div>
									{:else}
										<p
											class="font-mono text-sm whitespace-pre-wrap text-gray-800 dark:text-gray-200"
										>
											{note.body}
										</p>
										<div class="mt-2 flex items-center justify-between">
											<time class="text-xs text-gray-400 dark:text-gray-500"
												>{formatDate(note.created_at)}</time
											>
											<div class="flex gap-2">
												<button
													class="text-xs text-gray-400 transition-colors hover:text-indigo-600 dark:hover:text-indigo-400"
													onclick={() => startEditNote(note)}>Edit</button
												>
												<button
													class="text-xs text-gray-400 transition-colors hover:text-red-600 dark:hover:text-red-400"
													onclick={() => removeNote(note.id)}>Delete</button
												>
											</div>
										</div>
									{/if}
								</div>
							{/each}
						</div>
					{:else}
						<p class="mb-3 text-sm text-gray-400 italic dark:text-gray-500">No notes yet.</p>
					{/if}

					<div class="space-y-2">
						<textarea
							class={inputCls + ' resize-none font-mono'}
							rows="2"
							placeholder="Add a note… (observations, reasoning, findings)"
							bind:value={newNoteBody}
						></textarea>
						<button
							class="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-indigo-700 disabled:opacity-50"
							onclick={submitNote}
							disabled={addingNote || !newNoteBody.trim()}
							>{addingNote ? 'Posting...' : 'Add note'}</button
						>
					</div>
				</div>
			</div>
		</div>

		<!-- Footer -->
		<div
			class="flex shrink-0 items-center justify-between border-t border-gray-200 bg-white px-5 py-3.5 dark:border-gray-800 dark:bg-gray-900"
		>
			<button
				class="rounded-lg px-3 py-2 text-sm font-medium text-red-600 transition-colors hover:bg-red-50 disabled:opacity-50 dark:text-red-400 dark:hover:bg-red-950/30"
				onclick={removeTicket}
				disabled={deletingTicket}>{deletingTicket ? 'Deleting...' : 'Delete ticket'}</button
			>

			<button
				class="rounded-lg px-5 py-2 text-sm font-medium shadow-sm transition-colors
          {hasDraftChanges
					? 'bg-indigo-600 text-white hover:bg-indigo-700'
					: 'cursor-not-allowed bg-gray-100 text-gray-400 dark:bg-gray-800 dark:text-gray-500'}"
				onclick={saveAll}
				disabled={saving || !hasDraftChanges}>{saving ? 'Saving...' : 'Save changes'}</button
			>
		</div>
	{/if}
</div>
