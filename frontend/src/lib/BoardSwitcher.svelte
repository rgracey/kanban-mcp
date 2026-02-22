<script lang="ts">
  import type { Board } from './types.js'
  import CreateBoard from './modals/CreateBoard.svelte'

  interface Props {
    boards: Board[]
    selectedId: string | null
    onselect: (boardId: string) => void
    onboardcreate: (board: Board) => void
  }

  let { boards, selectedId, onselect, onboardcreate }: Props = $props()

  let showCreate = $state(false)
</script>

<div class="flex items-center gap-1 border-b border-gray-200 px-4 bg-white">
  {#each boards as board (board.id)}
    <button
      class="px-4 py-3 text-sm font-medium border-b-2 transition-colors -mb-px
        {selectedId === board.id
          ? 'border-indigo-600 text-indigo-600'
          : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
      onclick={() => onselect(board.id)}
    >
      {board.name}
    </button>
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
