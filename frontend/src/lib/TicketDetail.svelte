<script lang="ts">
  import type { Ticket, Epic } from './types.js'
  import { updateTicket, deleteTicket } from './api.js'

  interface Props {
    ticket: Ticket
    epics: Epic[]
    onclose: () => void
    onupdate: (ticket: Ticket) => void
    ondelete: (ticketId: string) => void
  }

  let { ticket, epics, onclose, onupdate, ondelete }: Props = $props()

  let title = $state('')
  let description = $state('')
  let priority = $state<Ticket['priority']>('medium')
  let epicId = $state('')
  let saving = $state(false)
  let error = $state('')

  // Initialise form fields from the ticket prop
  $effect(() => {
    title = ticket.title
    description = ticket.description ?? ''
    priority = ticket.priority
    epicId = ticket.epic_id ?? ''
  })

  async function save() {
    saving = true
    error = ''
    try {
      const updated = await updateTicket(ticket.id, {
        title,
        description,
        priority,
        epic_id: epicId || null,
      })
      onupdate(updated)
    } catch (e) {
      error = e instanceof Error ? e.message : 'Save failed'
    } finally {
      saving = false
    }
  }

  async function remove() {
    if (!confirm('Delete this ticket?')) return
    saving = true
    error = ''
    try {
      await deleteTicket(ticket.id)
      ondelete(ticket.id)
    } catch (e) {
      error = e instanceof Error ? e.message : 'Delete failed'
    } finally {
      saving = false
    }
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
<div
  role="dialog"
  aria-modal="true"
  tabindex="-1"
  class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
  onclick={(e) => { if (e.target === e.currentTarget) onclose() }}
>
  <div class="bg-white rounded-xl shadow-xl w-full max-w-lg p-6">
    <div class="flex justify-between items-start mb-4">
      <h2 class="text-lg font-semibold text-gray-900">Ticket Detail</h2>
      <button class="text-gray-400 hover:text-gray-600 text-xl leading-none" onclick={onclose}>&times;</button>
    </div>

    {#if error}
      <p class="text-red-600 text-sm mb-3">{error}</p>
    {/if}

    <div class="space-y-4">
      <div>
        <label for="td-title" class="block text-sm font-medium text-gray-700 mb-1">Title</label>
        <input
          id="td-title"
          class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          bind:value={title}
        />
      </div>

      <div>
        <label for="td-desc" class="block text-sm font-medium text-gray-700 mb-1">Description</label>
        <textarea
          id="td-desc"
          class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
          rows="4"
          bind:value={description}
        ></textarea>
      </div>

      <div class="flex gap-4">
        <div class="flex-1">
          <label for="td-priority" class="block text-sm font-medium text-gray-700 mb-1">Priority</label>
          <select
            id="td-priority"
            class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            bind:value={priority}
          >
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
            <option value="critical">Critical</option>
          </select>
        </div>

        <div class="flex-1">
          <label for="td-epic" class="block text-sm font-medium text-gray-700 mb-1">Epic</label>
          <select
            id="td-epic"
            class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            bind:value={epicId}
          >
            <option value="">None</option>
            {#each epics as epic (epic.id)}
              <option value={epic.id}>{epic.title}</option>
            {/each}
          </select>
        </div>
      </div>
    </div>

    <div class="flex justify-between items-center mt-6">
      <button
        class="px-3 py-2 text-sm text-red-600 hover:bg-red-50 rounded-lg transition-colors"
        onclick={remove}
        disabled={saving}
      >
        Delete
      </button>
      <div class="flex gap-2">
        <button
          class="px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
          onclick={onclose}
          disabled={saving}
        >
          Cancel
        </button>
        <button
          class="px-4 py-2 text-sm bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors disabled:opacity-50"
          onclick={save}
          disabled={saving}
        >
          {saving ? 'Saving...' : 'Save'}
        </button>
      </div>
    </div>
  </div>
</div>
