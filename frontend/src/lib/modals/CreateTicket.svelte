<script lang="ts">
	import type { Ticket, Epic, Status } from '../types.js';
	import { createTicket, listEpics } from '../api.js';
	import { toast } from 'svelte-sonner';

	interface Props {
		boardId: string;
		onclose: () => void;
		oncreated: (ticket: Ticket) => void;
	}

	let { boardId, onclose, oncreated }: Props = $props();

	let epics = $state<Epic[]>([]);
	let title = $state('');
	let description = $state('');
	let priority = $state<'low' | 'medium' | 'high' | 'critical'>('medium');
	let status = $state<Status>('todo');
	let epicId = $state('');
	let assignee = $state('');
	let submitting = $state(false);
	let error = $state('');

	$effect(() => {
		listEpics(boardId)
			.then((e) => (epics = e))
			.catch(() => {});
	});

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onclose();
	}

	async function submit() {
		if (!title.trim()) {
			error = 'Title is required';
			return;
		}
		submitting = true;
		error = '';
		try {
			const ticket = await createTicket(boardId, {
				title: title.trim(),
				description,
				priority,
				status,
				epic_id: epicId || null,
				assignee
			});
			oncreated(ticket);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Create failed';
			toast.error(error);
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- svelte-ignore a11y_click_events_have_key_events -->
<div
	role="presentation"
	class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm dark:bg-black/70"
	onclick={(e) => {
		if (e.target === e.currentTarget) onclose();
	}}
>
	<div
		role="dialog"
		aria-modal="true"
		tabindex="-1"
		class="w-full max-w-lg rounded-xl border border-gray-200 bg-white p-6 shadow-2xl dark:border-gray-700 dark:bg-gray-900"
	>
		<div class="mb-5 flex items-start justify-between">
			<h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">New Ticket</h2>
			<button
				class="flex h-7 w-7 items-center justify-center rounded-lg text-lg leading-none text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-800 dark:hover:text-gray-300"
				onclick={onclose}
				aria-label="Close">&times;</button
			>
		</div>

		{#if error}
			<p
				class="mb-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-500 dark:border-red-800 dark:bg-red-950/30"
			>
				{error}
			</p>
		{/if}

		<div class="space-y-4">
			<div>
				<label
					for="ct-title"
					class="mb-1.5 block text-xs font-medium text-gray-600 dark:text-gray-400"
					>Title <span class="text-red-500">*</span></label
				>
				<input
					id="ct-title"
					class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 placeholder-gray-400 transition-shadow focus:border-transparent focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder-gray-500"
					placeholder="Ticket title"
					bind:value={title}
				/>
			</div>

			<div>
				<label
					for="ct-desc"
					class="mb-1.5 block text-xs font-medium text-gray-600 dark:text-gray-400"
					>Description</label
				>
				<textarea
					id="ct-desc"
					class="w-full resize-none rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 placeholder-gray-400 transition-shadow focus:border-transparent focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder-gray-500"
					rows="3"
					placeholder="Optional description (Markdown supported)"
					bind:value={description}
				></textarea>
			</div>

			<div class="grid grid-cols-2 gap-3">
				<div>
					<label
						for="ct-priority"
						class="mb-1.5 block text-xs font-medium text-gray-600 dark:text-gray-400"
						>Priority</label
					>
					<select
						id="ct-priority"
						class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 transition-shadow focus:border-transparent focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100"
						bind:value={priority}
					>
						<option value="low">Low</option>
						<option value="medium">Medium</option>
						<option value="high">High</option>
						<option value="critical">Critical</option>
					</select>
				</div>

				<div>
					<label
						for="ct-status"
						class="mb-1.5 block text-xs font-medium text-gray-600 dark:text-gray-400">Status</label
					>
					<select
						id="ct-status"
						class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 transition-shadow focus:border-transparent focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100"
						bind:value={status}
					>
						<option value="todo">To Do</option>
						<option value="in_progress">In Progress</option>
						<option value="blocked">Blocked</option>
						<option value="done">Done</option>
					</select>
				</div>
			</div>

			<div>
				<label
					for="ct-epic"
					class="mb-1.5 block text-xs font-medium text-gray-600 dark:text-gray-400">Epic</label
				>
				<select
					id="ct-epic"
					class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 transition-shadow focus:border-transparent focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100"
					bind:value={epicId}
				>
					<option value="">None</option>
					{#each epics as epic (epic.id)}
						<option value={epic.id}>{epic.title}</option>
					{/each}
				</select>
			</div>

			<div>
				<label
					for="ct-assignee"
					class="mb-1.5 block text-xs font-medium text-gray-600 dark:text-gray-400">Assignee</label
				>
				<input
					id="ct-assignee"
					class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 placeholder-gray-400 transition-shadow focus:border-transparent focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder-gray-500"
					placeholder="Leave blank to leave unassigned"
					bind:value={assignee}
				/>
			</div>
		</div>

		<div class="mt-6 flex justify-end gap-2">
			<button
				class="rounded-lg px-4 py-2 text-sm font-medium text-gray-600 transition-colors hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800"
				onclick={onclose}
				disabled={submitting}>Cancel</button
			>
			<button
				class="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-indigo-700 disabled:opacity-50"
				onclick={submit}
				disabled={submitting}>{submitting ? 'Creating...' : 'Create Ticket'}</button
			>
		</div>
	</div>
</div>
