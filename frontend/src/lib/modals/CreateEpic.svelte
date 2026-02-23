<script lang="ts">
  import type { Epic } from '../types.js'
  import { createEpic } from '../api.js'
  import { toast } from 'svelte-sonner'

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
      toast.error(error)
    } finally {
      submitting = false
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- svelte-ignore a11y_click_events_have_key_events -->
<div
  role="presentation"
  class="fixed inset-0 bg-black/50 dark:bg-black/70 flex items-center justify-center z-50 p-4 backdrop-blur-sm"
  onclick={(e) => { if (e.target === e.currentTarget) onclose() }}
>
  <div
    role="dialog"
    aria-modal="true"
    tabindex="-1"
    class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-xl shadow-2xl w-full max-w-md p-6"
  >
    <div class="flex justify-between items-start mb-5">
      <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">New Epic</h2>
      <button
        class="w-7 h-7 flex items-center justify-center rounded-lg text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-lg leading-none"
        onclick={onclose}
        aria-label="Close"
      >&times;</button>
    </div>

    {#if error}
      <p class="text-red-500 text-sm mb-3 bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800 rounded-lg px-3 py-2">{error}</p>
    {/if}

    <div class="space-y-4">
      <div>
        <label for="ce-title" class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1.5">Title <span class="text-red-500">*</span></label>
        <input
          id="ce-title"
          class="w-full border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded-lg px-3 py-2 text-sm placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-shadow"
          placeholder="Epic title"
          bind:value={title}
        />
      </div>
      <div>
        <label for="ce-desc" class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1.5">Description</label>
        <textarea
          id="ce-desc"
          class="w-full border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded-lg px-3 py-2 text-sm placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none transition-shadow"
          rows="2"
          placeholder="Optional description"
          bind:value={description}
        ></textarea>
      </div>
    </div>

    <div class="flex justify-end gap-2 mt-6">
      <button
        class="px-4 py-2 text-sm font-medium text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
        onclick={onclose}
        disabled={submitting}
      >Cancel</button>
      <button
        class="px-4 py-2 text-sm font-medium bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50 shadow-sm"
        onclick={submit}
        disabled={submitting}
      >{submitting ? 'Creating...' : 'Create Epic'}</button>
    </div>
  </div>
</div>
