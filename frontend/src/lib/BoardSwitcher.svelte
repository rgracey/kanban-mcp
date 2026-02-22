<script lang="ts">
  import type { Board } from './types.js'
  import { createBoard } from './api.js'

  interface Props {
    boards: Board[]
    selectedId: string | null
    onselect: (boardId: string) => void
    onboardcreate: (board: Board) => void
  }

  let { boards, selectedId, onselect, onboardcreate }: Props = $props()

  let showCreate = $state(false)
  let newName = $state('')
  let newDesc = $state('')
  let creating = $state(false)
  let error = $state('')

  async function submitCreate() {
    if (!newName.trim()) { error = 'Name is required'; return }
    creating = true
    error = ''
    try {
      const board = await createBoard(newName.trim(), newDesc.trim() || undefined)
      onboardcreate(board)
      showCreate = false
      newName = ''
      newDesc = ''
    } catch (e) {
      error = e instanceof Error ? e.message : 'Create failed'
    } finally {
      creating = false
    }
  }
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
  <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
  <div
    role="dialog"
    aria-modal="true"
    tabindex="-1"
    class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
    onclick={(e) => { if (e.target === e.currentTarget) showCreate = false }}
  >
    <div class="bg-white rounded-xl shadow-xl w-full max-w-md p-6">
      <div class="flex justify-between items-start mb-4">
        <h2 class="text-lg font-semibold text-gray-900">New Board</h2>
        <button class="text-gray-400 hover:text-gray-600 text-xl leading-none" onclick={() => (showCreate = false)}>&times;</button>
      </div>

      {#if error}
        <p class="text-red-600 text-sm mb-3">{error}</p>
      {/if}

      <div class="space-y-4">
        <div>
          <label for="cb-name" class="block text-sm font-medium text-gray-700 mb-1">Name *</label>
          <input
            id="cb-name"
            class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            placeholder="Board name"
            bind:value={newName}
          />
        </div>
        <div>
          <label for="cb-desc" class="block text-sm font-medium text-gray-700 mb-1">Description</label>
          <input
            id="cb-desc"
            class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            placeholder="Optional description"
            bind:value={newDesc}
          />
        </div>
      </div>

      <div class="flex justify-end gap-2 mt-6">
        <button
          class="px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
          onclick={() => (showCreate = false)}
          disabled={creating}
        >
          Cancel
        </button>
        <button
          class="px-4 py-2 text-sm bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
          onclick={submitCreate}
          disabled={creating}
        >
          {creating ? 'Creating...' : 'Create Board'}
        </button>
      </div>
    </div>
  </div>
{/if}
