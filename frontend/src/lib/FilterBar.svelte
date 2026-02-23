<script lang="ts">
	import type { TicketFilter, Priority, SortBy, SortOrder } from './types.js';

	interface Props {
		filter: TicketFilter;
		onchange: (filter: TicketFilter) => void;
	}

	let { filter, onchange }: Props = $props();

	// qDraft tracks the raw input; we debounce before calling onchange
	let qDraft = $state('');
	let debounceTimer: ReturnType<typeof setTimeout> | null = null;

	// Sync qDraft when filter.q is cleared externally (e.g. "Clear all")
	$effect(() => {
		if (!filter.q) qDraft = '';
	});

	function update(partial: Partial<TicketFilter>) {
		onchange({ ...filter, ...partial });
	}

	function handleQInput(e: Event) {
		const val = (e.target as HTMLInputElement).value;
		qDraft = val;
		if (debounceTimer) clearTimeout(debounceTimer);
		debounceTimer = setTimeout(() => {
			update({ q: val || undefined });
		}, 300);
	}

	function handlePriorityChange(e: Event) {
		const val = (e.target as HTMLSelectElement).value as Priority | '';
		update({ priority: val || undefined });
	}

	function handleAssigneeInput(e: Event) {
		const val = (e.target as HTMLInputElement).value;
		if (debounceTimer) clearTimeout(debounceTimer);
		debounceTimer = setTimeout(() => {
			update({ assignee: val || undefined });
		}, 300);
	}

	function handleSortByChange(e: Event) {
		const val = (e.target as HTMLSelectElement).value as SortBy | '';
		update({ sort_by: val || undefined });
	}

	function handleSortOrderChange(e: Event) {
		const val = (e.target as HTMLSelectElement).value as SortOrder | '';
		update({ sort_order: val || undefined });
	}

	function clearAll() {
		qDraft = '';
		onchange({});
	}

	const hasFilters = $derived(
		!!(filter.q || filter.priority || filter.assignee || filter.sort_by)
	);
</script>

<div class="flex flex-wrap items-center gap-2">
	<!-- Keyword search -->
	<div class="relative">
		<svg
			class="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-gray-400"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
			stroke-width="2"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				d="m21 21-4.35-4.35M17 11A6 6 0 1 1 5 11a6 6 0 0 1 12 0Z"
			/>
		</svg>
		<input
			type="text"
			placeholder="Search..."
			value={qDraft}
			oninput={handleQInput}
			class="h-8 rounded-lg border border-gray-200 bg-white pl-8 pr-3 text-sm text-gray-700 placeholder-gray-400 transition-colors focus:border-indigo-400 focus:outline-none focus:ring-1 focus:ring-indigo-400 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200 dark:placeholder-gray-500 dark:focus:border-indigo-500"
		/>
	</div>

	<!-- Priority filter -->
	<select
		value={filter.priority ?? ''}
		onchange={handlePriorityChange}
		class="h-8 rounded-lg border border-gray-200 bg-white px-2 text-sm text-gray-700 transition-colors focus:border-indigo-400 focus:outline-none focus:ring-1 focus:ring-indigo-400 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200"
	>
		<option value="">All priorities</option>
		<option value="critical">Critical</option>
		<option value="high">High</option>
		<option value="medium">Medium</option>
		<option value="low">Low</option>
	</select>

	<!-- Assignee filter -->
	<div class="relative">
		<svg
			class="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-gray-400"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
			stroke-width="2"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z"
			/>
		</svg>
		<input
			type="text"
			placeholder="Assignee..."
			value={filter.assignee ?? ''}
			oninput={handleAssigneeInput}
			class="h-8 rounded-lg border border-gray-200 bg-white pl-8 pr-3 text-sm text-gray-700 placeholder-gray-400 transition-colors focus:border-indigo-400 focus:outline-none focus:ring-1 focus:ring-indigo-400 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200 dark:placeholder-gray-500 dark:focus:border-indigo-500"
		/>
	</div>

	<!-- Sort controls -->
	<select
		value={filter.sort_by ?? ''}
		onchange={handleSortByChange}
		class="h-8 rounded-lg border border-gray-200 bg-white px-2 text-sm text-gray-700 transition-colors focus:border-indigo-400 focus:outline-none focus:ring-1 focus:ring-indigo-400 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200"
	>
		<option value="">Default order</option>
		<option value="priority">Sort by priority</option>
		<option value="created_at">Sort by date</option>
	</select>

	{#if filter.sort_by}
		<select
			value={filter.sort_order ?? 'asc'}
			onchange={handleSortOrderChange}
			class="h-8 rounded-lg border border-gray-200 bg-white px-2 text-sm text-gray-700 transition-colors focus:border-indigo-400 focus:outline-none focus:ring-1 focus:ring-indigo-400 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200"
		>
			<option value="asc">Asc</option>
			<option value="desc">Desc</option>
		</select>
	{/if}

	<!-- Clear filters -->
	{#if hasFilters}
		<button
			onclick={clearAll}
			class="flex h-8 items-center gap-1 rounded-lg border border-gray-200 bg-white px-2.5 text-xs font-medium text-gray-500 transition-colors hover:border-gray-300 hover:bg-gray-50 hover:text-gray-700 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-400 dark:hover:border-gray-600 dark:hover:bg-gray-700/70 dark:hover:text-gray-200"
		>
			<svg
				class="h-3 w-3"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
				stroke-width="2.5"
			>
				<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
			</svg>
			Clear
		</button>
	{/if}
</div>
