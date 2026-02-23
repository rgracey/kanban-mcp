<script lang="ts">
  import { onMount } from 'svelte'
  import { Toaster, toast } from 'svelte-sonner'
  import type { Board } from './lib/types.js'
  import { listBoards } from './lib/api.js'
  import { initDarkMode, toggleDarkMode } from './lib/darkMode.js'
  import BoardSwitcher from './lib/BoardSwitcher.svelte'
  import KanbanBoard from './lib/KanbanBoard.svelte'

  let boards = $state<Board[]>([])
  let selectedId = $state<string | null>(null)
  let loading = $state(true)
  let error = $state('')
  let dark = $state(false)

  onMount(async () => {
    dark = initDarkMode()
    try {
      boards = await listBoards()
      if (boards.length > 0) selectedId = boards[0].id
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load boards'
      toast.error(error)
    } finally {
      loading = false
    }
  })

  function handleBoardCreate(board: Board) {
    boards = [...boards, board]
    selectedId = board.id
    toast.success(`Board "${board.name}" created`)
  }

  function handleBoardUpdate(board: Board) {
    boards = boards.map((b) => (b.id === board.id ? board : b))
  }

  function handleBoardDelete(boardId: string) {
    boards = boards.filter((b) => b.id !== boardId)
    if (selectedId === boardId) {
      selectedId = boards.length > 0 ? boards[0].id : null
    }
  }
</script>

<Toaster richColors position="bottom-right" />
<div class="min-h-screen bg-gray-50 dark:bg-gray-950 flex flex-col">
  {#if loading}
    <div class="flex items-center justify-center h-screen">
      <div class="flex items-center gap-3 text-gray-400">
        <svg class="animate-spin w-5 h-5" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        <span class="text-sm">Loading...</span>
      </div>
    </div>
  {:else if error}
    <div class="flex items-center justify-center h-screen text-red-500 text-sm">{error}</div>
  {:else}
    <!-- Top navigation bar -->
    <header class="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 shrink-0">
      <div class="flex items-center h-14 px-4 gap-4">
        <!-- Product wordmark -->
        <div class="flex items-center gap-2.5 shrink-0">
          <!-- Logo mark: simple grid of squares -->
          <div class="w-7 h-7 rounded-lg bg-indigo-600 flex items-center justify-center">
            <svg class="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 16 16">
              <rect x="1" y="1" width="6" height="6" rx="1"/>
              <rect x="9" y="1" width="6" height="6" rx="1"/>
              <rect x="1" y="9" width="6" height="6" rx="1"/>
              <rect x="9" y="9" width="6" height="6" rx="1"/>
            </svg>
          </div>
          <span class="text-sm font-semibold text-gray-900 dark:text-gray-100 tracking-tight">Kanban</span>
        </div>

        <!-- Divider -->
        <div class="w-px h-5 bg-gray-200 dark:bg-gray-700 shrink-0"></div>

        <!-- Board tabs — fills remaining space -->
        <div class="flex-1 min-w-0">
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
          class="shrink-0 w-8 h-8 flex items-center justify-center rounded-lg text-gray-400 dark:text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          onclick={() => (dark = toggleDarkMode(dark))}
          aria-label="Toggle dark mode"
          title="Toggle dark mode"
        >
          {#if dark}
            <svg xmlns="http://www.w3.org/2000/svg" class="w-4.5 h-4.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 3v1m0 16v1m8.66-9h-1M4.34 12h-1m15.07-6.07-.707.707M6.343 17.657l-.707.707m12.02 0-.707-.707M6.343 6.343l-.707-.707M12 7a5 5 0 100 10A5 5 0 0012 7z" />
            </svg>
          {:else}
            <svg xmlns="http://www.w3.org/2000/svg" class="w-4.5 h-4.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z" />
            </svg>
          {/if}
        </button>
      </div>
    </header>

    <main class="flex-1 p-6 overflow-hidden">
      {#if selectedId}
        <KanbanBoard boardId={selectedId} />
      {:else}
        <div class="flex flex-col items-center justify-center h-64 text-center">
          <div class="w-12 h-12 rounded-2xl bg-indigo-50 dark:bg-indigo-950/50 flex items-center justify-center mb-4">
            <svg class="w-6 h-6 text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z" />
            </svg>
          </div>
          <p class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">No boards yet</p>
          <p class="text-xs text-gray-400 dark:text-gray-500">Click <strong class="text-gray-600 dark:text-gray-400">"+ New board"</strong> above to get started.</p>
        </div>
      {/if}
    </main>
  {/if}
</div>
