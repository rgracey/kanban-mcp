<script lang="ts">
  import type { Board } from './types.js'
  import CreateBoard from './modals/CreateBoard.svelte'
  import EditBoard from './modals/EditBoard.svelte'

  interface Props {
    boards: Board[]
    selectedId: string | null
    onselect: (boardId: string) => void
    onboardcreate: (board: Board) => void
    onboardupdate: (board: Board) => void
    onboarddelete: (boardId: string) => void
  }

  let { boards, selectedId, onselect, onboardcreate, onboardupdate, onboarddelete }: Props = $props()

  let showCreate = $state(false)
  let editingBoard = $state<Board | null>(null)
</script>

<div class="flex items-center gap-1 border-b border-gray-200 px-4 bg-white">
  {#each boards as board (board.id)}
    <div class="flex items-center -mb-px">
      <button
        class="px-4 py-3 text-sm font-medium border-b-2 transition-colors
          {selectedId === board.id
            ? 'border-indigo-600 text-indigo-600'
            : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
        onclick={() => onselect(board.id)}
      >
        {board.name}
      </button>
      {#if selectedId === board.id}
        <button
          class="p-1 text-gray-400 hover:text-indigo-600 transition-colors"
          onclick={() => (editingBoard = board)}
          aria-label="Edit board"
          title="Edit board"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5" viewBox="0 0 20 20" fill="currentColor">
            <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
          </svg>
        </button>
      {/if}
    </div>
  {/each}

  <button
    class="px-4 py-3 text-sm font-medium border-b-2 border-transparent text-gray-400 hover:text-indigo-600 transition-colors -mb-px ml-2"
    onclick={() => (showCreate = true)}
  >
    + New board
  </button>
</div>

{#if showCreate}
  <CreateBoard
    onclose={() => (showCreate = false)}
    oncreated={(board) => {
      onboardcreate(board)
      showCreate = false
    }}
  />
{/if}

{#if editingBoard}
  <EditBoard
    board={editingBoard}
    onclose={() => (editingBoard = null)}
    onupdated={(board) => {
      onboardupdate(board)
      editingBoard = null
    }}
    ondeleted={(id) => {
      onboarddelete(id)
      editingBoard = null
    }}
  />
{/if}
