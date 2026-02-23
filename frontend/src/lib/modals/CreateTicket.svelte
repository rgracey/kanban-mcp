<script lang="ts">
  import type { Ticket, Epic } from '../types.js'
  import { createTicket, listEpics } from '../api.js'
  import { toast } from 'svelte-sonner'

  interface Props {
    boardId: string
    onclose: () => void
    oncreated: (ticket: Ticket) => void
  }

  let { boardId, onclose, oncreated }: Props = $props()

  let epics = $state<Epic[]>([])
  let title = $state('')
  let description = $state('')
  let priority = $state<'low' | 'medium' | 'high' | 'critical'>('medium')
  let epicId = $state('')
  let assignee = $state('')
  let submitting = $state(false)
  let error = $state('')

  $effect(() => {
    listEpics(boardId).then((e) => (epics = e)).catch(() => {})
  })

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') onclose()
  }

  async function submit() {
    if (!title.trim()) { error = 'Title is required'; return }
    submitting = true
    error = ''
    try {
      const ticket = await createTicket(boardId, {
        title: title.trim(),
        description,
        priority,
        epic_id: epicId || null,
        assignee,
        status: 'todo',
      })
      oncreated(ticket)
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
    class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-xl shadow-2xl w-full max-w-lg p-6"
  >
    <div class="flex justify-between items-start mb-5">
      <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">New Ticket</h2>
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
        <label for="ct-title" class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1.5">Title <span class="text-red-500">*</span></label>
        <input
          id="ct-title"
          class="w-full border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded-lg px-3 py-2 text-sm placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-shadow"
          placeholder="Ticket title"
          bind:value={title}
        />
      </div>

      <div>
        <label for="ct-desc" class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1.5">Description</label>
        <textarea
          id="ct-desc"
          class="w-full border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded-lg px-3 py-2 text-sm placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent resize-none transition-shadow"
          rows="3"
          placeholder="Optional description (Markdown supported)"
          bind:value={description}
        ></textarea>
      </div>

      <div class="grid grid-cols-2 gap-3">
        <div>
          <label for="ct-priority" class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1.5">Priority</label>
          <select
            id="ct-priority"
            class="w-full border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-shadow"
            bind:value={priority}
          >
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
            <option value="critical">Critical</option>
          </select>
        </div>

        <div>
          <label for="ct-epic" class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1.5">Epic</label>
          <select
            id="ct-epic"
            class="w-full border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-shadow"
            bind:value={epicId}
          >
            <option value="">None</option>
            {#each epics as epic (epic.id)}
              <option value={epic.id}>{epic.title}</option>
            {/each}
          </select>
        </div>
      </div>

      <div>
        <label for="ct-assignee" class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1.5">Assignee</label>
        <input
          id="ct-assignee"
          class="w-full border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded-lg px-3 py-2 text-sm placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-shadow"
          placeholder="Leave blank to leave unassigned"
          bind:value={assignee}
        />
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
      >{submitting ? 'Creating...' : 'Create Ticket'}</button>
    </div>
  </div>
</div>
