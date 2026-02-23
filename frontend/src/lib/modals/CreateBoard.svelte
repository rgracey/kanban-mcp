<script lang="ts">
	import type { Board } from '../types.js';
	import { createBoard } from '../api.js';
	import { toast } from 'svelte-sonner';

	interface Props {
		onclose: () => void;
		oncreated: (board: Board) => void;
	}

	let { onclose, oncreated }: Props = $props();

	let name = $state('');
	let description = $state('');
	let submitting = $state(false);
	let error = $state('');

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onclose();
	}

	async function submit() {
		if (!name.trim()) {
			error = 'Name is required';
			return;
		}
		submitting = true;
		error = '';
		try {
			const board = await createBoard(name.trim(), description.trim() || undefined);
			oncreated(board);
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
		class="w-full max-w-md rounded-xl border border-gray-200 bg-white p-6 shadow-2xl dark:border-gray-700 dark:bg-gray-900"
	>
		<div class="mb-5 flex items-start justify-between">
			<h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">New Board</h2>
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
					for="cb-name"
					class="mb-1.5 block text-xs font-medium text-gray-600 dark:text-gray-400"
					>Name <span class="text-red-500">*</span></label
				>
				<input
					id="cb-name"
					class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 placeholder-gray-400 transition-shadow focus:border-transparent focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder-gray-500"
					placeholder="Board name"
					bind:value={name}
				/>
			</div>
			<div>
				<label
					for="cb-desc"
					class="mb-1.5 block text-xs font-medium text-gray-600 dark:text-gray-400"
					>Description</label
				>
				<textarea
					id="cb-desc"
					class="w-full resize-none rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 placeholder-gray-400 transition-shadow focus:border-transparent focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 dark:placeholder-gray-500"
					rows="2"
					placeholder="Optional description"
					bind:value={description}
				></textarea>
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
				disabled={submitting}>{submitting ? 'Creating...' : 'Create Board'}</button
			>
		</div>
	</div>
</div>
