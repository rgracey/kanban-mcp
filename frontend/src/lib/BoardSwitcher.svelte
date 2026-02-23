<script lang="ts">
	import type { Board } from './types.js';
	import CreateBoard from './modals/CreateBoard.svelte';
	import EditBoard from './modals/EditBoard.svelte';
	import CopyId from './CopyId.svelte';

	interface Props {
		boards: Board[];
		selectedId: string | null;
		onselect: (boardId: string) => void;
		onboardcreate: (board: Board) => void;
		onboardupdate: (board: Board) => void;
		onboarddelete: (boardId: string) => void;
	}

	let { boards, selectedId, onselect, onboardcreate, onboardupdate, onboarddelete }: Props =
		$props();

	let showCreate = $state(false);
	let editingBoard = $state<Board | null>(null);
</script>

<nav class="flex h-full items-center gap-0.5" aria-label="Boards">
	{#each boards as board (board.id)}
		{@const active = selectedId === board.id}
		<div class="group flex items-center">
			<button
				class="relative flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm transition-colors
          {active
					? 'bg-indigo-50 font-medium text-indigo-700 dark:bg-indigo-950/60 dark:text-indigo-300'
					: 'font-normal text-gray-500 hover:bg-gray-100 hover:text-gray-800 dark:text-gray-400 dark:hover:bg-gray-800 dark:hover:text-gray-200'}"
				onclick={() => onselect(board.id)}
			>
				{board.name}
				{#if active}
					<CopyId id={board.id} />
				{/if}
			</button>
			{#if active}
				<button
					class="-ml-0.5 flex h-6 w-6 items-center justify-center rounded-md text-gray-400 opacity-0 transition-colors group-hover:opacity-100 hover:bg-indigo-50 hover:text-indigo-600 focus:opacity-100 dark:text-gray-500 dark:hover:bg-indigo-950/60 dark:hover:text-indigo-400"
					onclick={() => (editingBoard = board)}
					aria-label="Edit board"
					title="Edit board"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						class="h-3 w-3"
						viewBox="0 0 20 20"
						fill="currentColor"
					>
						<path
							d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z"
						/>
					</svg>
				</button>
			{/if}
		</div>
	{/each}

	<button
		class="ml-1 flex items-center gap-1 rounded-md px-3 py-1.5 text-sm text-gray-400 transition-colors hover:bg-indigo-50 hover:text-indigo-600 dark:text-gray-500 dark:hover:bg-indigo-950/60 dark:hover:text-indigo-400"
		onclick={() => (showCreate = true)}
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
		New board
	</button>
</nav>

{#if showCreate}
	<CreateBoard
		onclose={() => (showCreate = false)}
		oncreated={(board) => {
			onboardcreate(board);
			showCreate = false;
		}}
	/>
{/if}

{#if editingBoard}
	<EditBoard
		board={editingBoard}
		onclose={() => (editingBoard = null)}
		onupdated={(board) => {
			onboardupdate(board);
			editingBoard = null;
		}}
		ondeleted={(id) => {
			onboarddelete(id);
			editingBoard = null;
		}}
	/>
{/if}
