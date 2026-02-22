<script lang="ts">
  import type { Epic } from '../types.js'
  import { createEpic } from '../api.js'

  interface Props {
    boardId: string
    onclose: () => void
    oncreated: (epic: Epic) => void
  }

  let { boardId, onclose, oncreated }: Props = $props()

  let title = $state('')
  let description = $state('')
  let submitting = $state(false)
  let error = $state('')

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') onclose()
  }

  async function submit() {
    if (!title.trim()) { error = 'Title is required'; return }
    submitting = true
    error = ''
    try {
      const epic = await createEpic(boardId, title.trim(), description.trim() || undefined)
      oncreated(epic)
    } catch (e) {
      error = e instanceof Error ? e.message : 'Create failed'
    } finally {
      submitting = false
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
      <h2 class="text-lg font-semibold text-gray-900">New Epic</h2>
      <button class="text-gray-400 hover:text-gray-600 text-xl leading-none" onclick={onclose}>&times;</button>
    </div>

    {#if error}
      <p class="text-red-600 text-sm mb-3">{error}</p>
    {/if}

    <div class="space-y-4">
      <div>
        <label for="ce-title" class="block text-sm font-medium text-gray-700 mb-1">Title *</label>
        <input
          id="ce-title"
          class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          placeholder="Epic title"
          bind:value={title}
        />
      </div>
      <div>
        <label for="ce-desc" class="block text-sm font-medium text-gray-700 mb-1">Description</label>
        <textarea
          id="ce-desc"
          class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
          rows="2"
          placeholder="Optional description"
          bind:value={description}
        ></textarea>
      </div>
    </div>

    <div class="flex justify-end gap-2 mt-6">
      <button
        class="px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
        onclick={onclose}
        disabled={submitting}
      >Cancel</button>
      <button
        class="px-4 py-2 text-sm bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
        onclick={submit}
        disabled={submitting}
      >{submitting ? 'Creating...' : 'Create Epic'}</button>
    </div>
  </div>
</div>
