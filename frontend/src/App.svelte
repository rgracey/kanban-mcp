<script lang="ts">
	import { onMount } from 'svelte';
	import { Toaster, toast } from 'svelte-sonner';
	import type { Board } from './lib/types.js';
	import { listBoards } from './lib/api.js';
	import { initDarkMode, toggleDarkMode } from './lib/darkMode.js';
	import BoardSwitcher from './lib/BoardSwitcher.svelte';
	import KanbanBoard from './lib/KanbanBoard.svelte';

	let boards = $state<Board[]>([]);
	let selectedId = $state<string | null>(null);
	let loading = $state(true);
	let error = $state('');
	let dark = $state(false);

	onMount(async () => {
		dark = initDarkMode();
		try {
			boards = await listBoards();
			if (boards.length > 0) selectedId = boards[0].id;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load boards';
			toast.error(error);
		} finally {
			loading = false;
		}
	});

	function handleBoardCreate(board: Board) {
		boards = [...boards, board];
		selectedId = board.id;
		toast.success(`Board "${board.name}" created`);
	}

	function handleBoardUpdate(board: Board) {
		boards = boards.map((b) => (b.id === board.id ? board : b));
	}

	function handleBoardDelete(boardId: string) {
		boards = boards.filter((b) => b.id !== boardId);
		if (selectedId === boardId) {
			selectedId = boards.length > 0 ? boards[0].id : null;
		}
	}
</script>

<Toaster richColors position="bottom-right" />
<div class="flex min-h-screen flex-col bg-gray-50 dark:bg-gray-950">
	{#if loading}
		<div class="flex h-screen items-center justify-center">
			<div class="flex items-center gap-3 text-gray-400">
				<svg class="h-5 w-5 animate-spin" fill="none" viewBox="0 0 24 24">
					<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"
					></circle>
					<path
						class="opacity-75"
						fill="currentColor"
						d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
					></path>
				</svg>
				<span class="text-sm">Loading...</span>
			</div>
		</div>
	{:else if error}
		<div class="flex h-screen items-center justify-center text-sm text-red-500">{error}</div>
	{:else}
		<!-- Top navigation bar -->
		<header
			class="shrink-0 border-b border-gray-200 bg-white dark:border-gray-800 dark:bg-gray-900"
		>
			<div class="flex h-14 items-center gap-4 px-4">
				<!-- Product wordmark -->
				<div class="flex shrink-0 items-center gap-2.5">
					<!-- Logo mark: simple grid of squares -->
					<div class="flex h-7 w-7 items-center justify-center rounded-lg bg-indigo-600">
						<svg class="h-4 w-4 text-white" fill="currentColor" viewBox="0 0 16 16">
							<rect x="1" y="1" width="6" height="6" rx="1" />
							<rect x="9" y="1" width="6" height="6" rx="1" />
							<rect x="1" y="9" width="6" height="6" rx="1" />
							<rect x="9" y="9" width="6" height="6" rx="1" />
						</svg>
					</div>
					<span class="text-sm font-semibold tracking-tight text-gray-900 dark:text-gray-100"
						>Kanban</span
					>
				</div>

				<!-- Divider -->
				<div class="h-5 w-px shrink-0 bg-gray-200 dark:bg-gray-700"></div>

				<!-- Board tabs — fills remaining space -->
				<div class="min-w-0 flex-1">
					<BoardSwitcher
						{boards}
						{selectedId}
						onselect={(id) => (selectedId = id)}
						onboardcreate={handleBoardCreate}
						onboardupdate={handleBoardUpdate}
						onboarddelete={handleBoardDelete}
					/>
				</div>

				<!-- Dark mode toggle -->
				<button
					class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-gray-500 dark:hover:bg-gray-800 dark:hover:text-gray-300"
					onclick={() => (dark = toggleDarkMode(dark))}
					aria-label="Toggle dark mode"
					title="Toggle dark mode"
				>
					{#if dark}
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-4.5 w-4.5"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							stroke-width="2"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								d="M12 3v1m0 16v1m8.66-9h-1M4.34 12h-1m15.07-6.07-.707.707M6.343 17.657l-.707.707m12.02 0-.707-.707M6.343 6.343l-.707-.707M12 7a5 5 0 100 10A5 5 0 0012 7z"
							/>
						</svg>
					{:else}
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-4.5 w-4.5"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							stroke-width="2"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z"
							/>
						</svg>
					{/if}
				</button>
			</div>
		</header>

		<main class="flex-1 overflow-hidden p-6">
			{#if selectedId}
				<KanbanBoard boardId={selectedId} />
			{:else}
				<div class="flex h-64 flex-col items-center justify-center text-center">
					<div
						class="mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-indigo-50 dark:bg-indigo-950/50"
					>
						<svg
							class="h-6 w-6 text-indigo-400"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							stroke-width="1.5"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								d="M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z"
							/>
						</svg>
					</div>
					<p class="mb-1 text-sm font-medium text-gray-700 dark:text-gray-300">No boards yet</p>
					<p class="text-xs text-gray-400 dark:text-gray-500">
						Click <strong class="text-gray-600 dark:text-gray-400">"+ New board"</strong> above to get
						started.
					</p>
				</div>
			{/if}
		</main>
	{/if}
</div>
