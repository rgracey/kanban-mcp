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
<div class="min-h-screen bg-gray-100 dark:bg-gray-900 flex flex-col">
  {#if loading}
    <div class="flex items-center justify-center h-screen text-gray-400 text-sm">Loading...</div>
  {:else if error}
    <div class="flex items-center justify-center h-screen text-red-500 text-sm">{error}</div>
  {:else}
    <div class="flex items-center bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
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
      <button
        class="shrink-0 px-3 py-3 mr-2 text-gray-500 dark:text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors"
        onclick={() => (dark = toggleDarkMode(dark))}
        aria-label="Toggle dark mode"
        title="Toggle dark mode"
      >
        {#if dark}
          <!-- Sun icon -->
          <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 3v1m0 16v1m8.66-9h-1M4.34 12h-1m15.07-6.07-.707.707M6.343 17.657l-.707.707m12.02 0-.707-.707M6.343 6.343l-.707-.707M12 7a5 5 0 100 10A5 5 0 0012 7z" />
          </svg>
        {:else}
          <!-- Moon icon -->
          <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z" />
          </svg>
        {/if}
      </button>
    </div>

    <main class="flex-1 p-6">
      {#if selectedId}
        <KanbanBoard boardId={selectedId} />
      {:else}
        <div class="flex flex-col items-center justify-center h-64 text-gray-400 dark:text-gray-500">
          <p class="text-lg mb-2">No boards yet</p>
          <p class="text-sm">Click <strong>"+ New board"</strong> above to get started.</p>
        </div>
      {/if}
    </main>
  {/if}
</div>
