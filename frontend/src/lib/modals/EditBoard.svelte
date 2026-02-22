<script lang="ts">
  import { untrack } from 'svelte'
  import type { Board } from '../types.js'
  import { updateBoard, deleteBoard } from '../api.js'
  import { toast } from 'svelte-sonner'

  interface Props {
    board: Board
    onclose: () => void
    onupdated: (board: Board) => void
    ondeleted: (boardId: string) => void
  }

  let { board, onclose, onupdated, ondeleted }: Props = $props()

  // untrack prevents Svelte from treating these as reactive dependencies —
  // the modal seeds form fields from the board prop once on open.
  let name = $state(untrack(() => board.name))
  let description = $state(untrack(() => board.description ?? ''))
  let submitting = $state(false)
  let deleting = $state(false)
  let error = $state('')

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') onclose()
  }

  async function submit() {
    if (!name.trim()) { error = 'Name is required'; return }
    submitting = true
    error = ''
    try {
      const updated = await updateBoard(board.id, {
        name: name.trim(),
        description: description.trim() || undefined,
      })
      toast.success(`Board "${updated.name}" updated`)
      onupdated(updated)
    } catch (e) {
      error = e instanceof Error ? e.message : 'Update failed'
      toast.error(error)
    } finally {
      submitting = false
    }
  }

  async function remove() {
    if (!confirm(`Delete board "${board.name}"? This will delete all its epics and tickets and cannot be undone.`)) return
    deleting = true
    try {
      await deleteBoard(board.id)
      toast.success(`Board "${board.name}" deleted`)
      ondeleted(board.id)
    } catch (e) {
      error = e instanceof Error ? e.message : 'Delete failed'
      toast.error(error)
    } finally {
      deleting = false
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- svelte-ignore a11y_click_events_have_key_events -->
<div
  role="presentation"
  class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
  onclick={(e) => { if (e.target === e.currentTarget) onclose() }}
>
  <div
    role="dialog"
    aria-modal="true"
    tabindex="-1"
    class="bg-white rounded-xl shadow-xl w-full max-w-md p-6"
  >
    <div class="flex justify-between items-start mb-5">
      <h2 class="text-lg font-semibold text-gray-900">Edit Board</h2>
      <button class="text-gray-400 hover:text-gray-600 text-xl leading-none" onclick={onclose}>&times;</button>
    </div>

    {#if error}
      <p class="text-red-600 text-sm mb-3">{error}</p>
    {/if}

    <div class="space-y-4">
      <div>
        <label for="eb-name" class="block text-sm font-medium text-gray-700 mb-1">Name *</label>
        <input
          id="eb-name"
          class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          placeholder="Board name"
          bind:value={name}
        />
      </div>
      <div>
        <label for="eb-desc" class="block text-sm font-medium text-gray-700 mb-1">Description</label>
        <textarea
          id="eb-desc"
          class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
          rows="2"
          placeholder="Optional description"
          bind:value={description}
        ></textarea>
      </div>
    </div>

    <div class="flex justify-between items-center mt-6">
      <button
        class="px-3 py-2 text-sm text-red-600 hover:bg-red-50 rounded-lg transition-colors disabled:opacity-50"
        onclick={remove}
        disabled={deleting || submitting}
      >{deleting ? 'Deleting...' : 'Delete board'}</button>

      <div class="flex gap-2">
        <button
          class="px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
          onclick={onclose}
          disabled={submitting || deleting}
        >Cancel</button>
        <button
          class="px-4 py-2 text-sm bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
          onclick={submit}
          disabled={submitting || deleting}
        >{submitting ? 'Saving...' : 'Save'}</button>
      </div>
    </div>
  </div>
</div>
