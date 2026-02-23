<script lang="ts">
	import { untrack } from 'svelte';
	import type { Epic } from '../types.js';
	import { updateEpic, deleteEpic } from '../api.js';
	import { toast } from 'svelte-sonner';

	interface Props {
		epic: Epic;
		onclose: () => void;
		onupdated: (epic: Epic) => void;
		ondeleted: (epicId: string) => void;
	}

	let { epic, onclose, onupdated, ondeleted }: Props = $props();

	let title = $state(untrack(() => epic.title));
	let description = $state(untrack(() => epic.description ?? ''));
	let submitting = $state(false);
	let deleting = $state(false);
	let error = $state('');

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
			const updated = await updateEpic(epic.id, {
				title: title.trim(),
				description: description.trim() || undefined
			});
			toast.success(`Epic "${updated.title}" updated`);
			onupdated(updated);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Update failed';
			toast.error(error);
		} finally {
			submitting = false;
		}
	}

	async function remove() {
		if (
			!confirm(
				`Delete epic "${epic.title}"? Tickets in this epic will be kept but unassigned from it.`
			)
		)
			return;
		deleting = true;
		try {
			await deleteEpic(epic.id);
			toast.success(`Epic "${epic.title}" deleted`);
			ondeleted(epic.id);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Delete failed';
			toast.error(error);
		} finally {
			deleting = false;
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
		class="w-full max-w-md rounded-xl border border-gray-200 bg-white p-6 shadow-2xl dark:border-gray-700 dark:bg-gray-900"
	>
		<div class="mb-5 flex items-start justify-between">
			<h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">Edit Epic</h2>
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
					for="ee-title"
					class="mb-1.5 block text-xs font-medium text-gray-600 dark:text-gray-400"
					>Title <span class="text-red-500">*</span></label
				>
				<input
					id="ee-title"
					class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 placeholder-gray-400 transition-shadow focus:border-transparent focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder-gray-500"
					placeholder="Epic title"
					bind:value={title}
				/>
			</div>
			<div>
				<label
					for="ee-desc"
					class="mb-1.5 block text-xs font-medium text-gray-600 dark:text-gray-400"
					>Description</label
				>
				<textarea
					id="ee-desc"
					class="w-full resize-none rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 placeholder-gray-400 transition-shadow focus:border-transparent focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder-gray-500"
					rows="2"
					placeholder="Optional description"
					bind:value={description}
				></textarea>
			</div>
		</div>

		<div class="mt-6 flex items-center justify-between">
			<button
				class="rounded-lg px-3 py-2 text-sm font-medium text-red-600 transition-colors hover:bg-red-50 disabled:opacity-50 dark:text-red-400 dark:hover:bg-red-950/30"
				onclick={remove}
				disabled={deleting || submitting}>{deleting ? 'Deleting...' : 'Delete epic'}</button
			>

			<div class="flex gap-2">
				<button
					class="rounded-lg px-4 py-2 text-sm font-medium text-gray-600 transition-colors hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800"
					onclick={onclose}
					disabled={submitting || deleting}>Cancel</button
				>
				<button
					class="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition-colors hover:bg-indigo-700 disabled:opacity-50"
					onclick={submit}
					disabled={submitting || deleting}>{submitting ? 'Saving...' : 'Save'}</button
				>
			</div>
		</div>
	</div>
</div>
