<script lang="ts">
	import type { Epic } from './types.js';
	import CopyId from './CopyId.svelte';

	interface Props {
		epics: Epic[];
		selectedEpicId: string | null;
		onchange: (epicId: string | null) => void;
		onedit: (epic: Epic) => void;
	}

	let { epics, selectedEpicId, onchange, onedit }: Props = $props();
</script>

<div class="flex flex-wrap items-center gap-1.5">
	<button
		class="rounded-md px-2.5 py-1 text-xs font-medium transition-colors
      {selectedEpicId === null
			? 'bg-indigo-600 text-white shadow-sm'
			: 'bg-gray-100 text-gray-600 hover:bg-gray-200 hover:text-gray-800 dark:bg-gray-800 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-gray-200'}"
		onclick={() => onchange(null)}
	>
		All
	</button>
	{#each epics as epic (epic.id)}
		<div class="group relative flex items-center">
			<button
				class="rounded-md py-1 pl-2.5 text-xs font-medium transition-colors
          {selectedEpicId === epic.id
					? 'bg-indigo-600 pr-6 text-white shadow-sm'
					: 'bg-gray-100 pr-6 text-gray-600 hover:bg-gray-200 hover:text-gray-800 dark:bg-gray-800 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-gray-200'}"
				onclick={() => onchange(epic.id)}
			>
				{epic.title}
				<CopyId id={epic.id} />
			</button>
			<!-- Edit icon, positioned over the right side of the pill -->
			<button
				class="absolute right-1 flex h-4 w-4 items-center justify-center rounded opacity-0 transition-opacity group-hover:opacity-100
          {selectedEpicId === epic.id
					? 'text-indigo-200 hover:text-white'
					: 'text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'}"
				onclick={(e) => {
					e.stopPropagation();
					onedit(epic);
				}}
				aria-label="Edit epic"
				title="Edit epic"
			>
				<svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L6.832 19.82a4.5 4.5 0 01-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 011.13-1.897L16.863 4.487z"
					/>
				</svg>
			</button>
		</div>
	{/each}
</div>
