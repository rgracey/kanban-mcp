<script lang="ts">
  import { onMount } from 'svelte'
  import { Toaster, toast } from 'svelte-sonner'
  import type { Board } from './lib/types.js'
  import { listBoards } from './lib/api.js'
  import BoardSwitcher from './lib/BoardSwitcher.svelte'
  import KanbanBoard from './lib/KanbanBoard.svelte'

  let boards = $state<Board[]>([])
  let selectedId = $state<string | null>(null)
  let loading = $state(true)
  let error = $state('')

  onMount(async () => {
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
<div class="min-h-screen bg-gray-100 flex flex-col">
  {#if loading}
    <div class="flex items-center justify-center h-screen text-gray-400 text-sm">Loading...</div>
  {:else if error}
    <div class="flex items-center justify-center h-screen text-red-500 text-sm">{error}</div>
  {:else}
    <BoardSwitcher
      {boards}
      {selectedId}
      onselect={(id) => (selectedId = id)}
      onboardcreate={handleBoardCreate}
      onboardupdate={handleBoardUpdate}
      onboarddelete={handleBoardDelete}
    />

    <main class="flex-1 p-6">
      {#if selectedId}
        <KanbanBoard boardId={selectedId} />
      {:else}
        <div class="flex flex-col items-center justify-center h-64 text-gray-400">
          <p class="text-lg mb-2">No boards yet</p>
          <p class="text-sm">Click <strong>"+ New board"</strong> above to get started.</p>
        </div>
      {/if}
    </main>
  {/if}
</div>
