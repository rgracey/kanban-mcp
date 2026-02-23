<script lang="ts">
	import type { Ticket, Epic } from './types.js';
	import { createTicket } from './api.js';

	interface Props {
		boardId: string;
		epics: Epic[];
		onclose: () => void;
		oncreate: (ticket: Ticket) => void;
	}

	let { boardId, epics, onclose, oncreate }: Props = $props();

	let title = $state('');
	let description = $state('');
	let priority = $state<'low' | 'medium' | 'high' | 'critical'>('medium');
	let epicId = $state('');
	let saving = $state(false);
	let error = $state('');

	async function submit() {
		if (!title.trim()) {
			error = 'Title is required';
			return;
		}
		saving = true;
		error = '';
		try {
			const ticket = await createTicket(boardId, {
				title: title.trim(),
				description,
				priority,
				epic_id: epicId || null
			});
			oncreate(ticket);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Create failed';
		} finally {
			saving = false;
		}
	}
</script>

<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
<div
	role="dialog"
	aria-modal="true"
	tabindex="-1"
	class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
	onclick={(e) => {
		if (e.target === e.currentTarget) onclose();
	}}
>
	<div class="w-full max-w-lg rounded-xl bg-white p-6 shadow-xl">
		<div class="mb-4 flex items-start justify-between">
			<h2 class="text-lg font-semibold text-gray-900">New Ticket</h2>
			<button class="text-xl leading-none text-gray-400 hover:text-gray-600" onclick={onclose}
				>&times;</button
			>
		</div>

		{#if error}
			<p class="mb-3 text-sm text-red-600">{error}</p>
		{/if}

		<div class="space-y-4">
			<div>
				<label for="ct-title" class="mb-1 block text-sm font-medium text-gray-700">Title *</label>
				<input
					id="ct-title"
					class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
					placeholder="Ticket title"
					bind:value={title}
				/>
			</div>

			<div>
				<label for="ct-desc" class="mb-1 block text-sm font-medium text-gray-700">Description</label
				>
				<textarea
					id="ct-desc"
					class="w-full resize-none rounded-lg border border-gray-300 px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
					rows="3"
					bind:value={description}
				></textarea>
			</div>

			<div class="flex gap-4">
				<div class="flex-1">
					<label for="ct-priority" class="mb-1 block text-sm font-medium text-gray-700"
						>Priority</label
					>
					<select
						id="ct-priority"
						class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
						bind:value={priority}
					>
						<option value="low">Low</option>
						<option value="medium">Medium</option>
						<option value="high">High</option>
						<option value="critical">Critical</option>
					</select>
				</div>

				<div class="flex-1">
					<label for="ct-epic" class="mb-1 block text-sm font-medium text-gray-700">Epic</label>
					<select
						id="ct-epic"
						class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none"
						bind:value={epicId}
					>
						<option value="">None</option>
						{#each epics as epic (epic.id)}
							<option value={epic.id}>{epic.title}</option>
						{/each}
					</select>
				</div>
			</div>
		</div>

		<div class="mt-6 flex justify-end gap-2">
			<button
				class="rounded-lg px-4 py-2 text-sm text-gray-700 transition-colors hover:bg-gray-100"
				onclick={onclose}
				disabled={saving}
			>
				Cancel
			</button>
			<button
				class="rounded-lg bg-indigo-600 px-4 py-2 text-sm text-white transition-colors hover:bg-indigo-700 disabled:opacity-50"
				onclick={submit}
				disabled={saving}
			>
				{saving ? 'Creating...' : 'Create Ticket'}
			</button>
		</div>
	</div>
</div>
