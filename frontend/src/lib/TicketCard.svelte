<script lang="ts">
  import type { Ticket, Epic } from './types.js'
  import TicketDetail from './TicketDetail.svelte'

  interface Props {
    ticket: Ticket
    boardId: string
    epics: Epic[]
    onupdate?: (ticket: Ticket) => void
    ondelete?: (ticketId: string) => void
  }

  let { ticket, boardId, epics, onupdate, ondelete }: Props = $props()

  let showDetail = $state(false)

  const priorityClasses: Record<string, string> = {
    low: 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300',
    medium: 'bg-blue-100 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300',
    high: 'bg-orange-100 text-orange-700 dark:bg-orange-900/50 dark:text-orange-300',
    critical: 'bg-red-100 text-red-700 dark:bg-red-900/50 dark:text-red-300',
  }

  const statusBorder: Record<string, string> = {
    todo: 'border-l-4 border-l-gray-300 dark:border-l-gray-600',
    in_progress: 'border-l-4 border-l-blue-400 dark:border-l-blue-500',
    done: 'border-l-4 border-l-green-400 dark:border-l-green-500',
  }

  const epicName = $derived(
    ticket.epic_id ? epics.find((e) => e.id === ticket.epic_id)?.title ?? null : null
  )
</script>

<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
<div
  role="button"
  tabindex="0"
  class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-3 cursor-pointer hover:shadow-md transition-shadow select-none {statusBorder[ticket.status]}"
  onclick={() => (showDetail = true)}
  onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') showDetail = true }}
>
  <p class="text-sm font-medium text-gray-900 dark:text-gray-100 mb-2 leading-snug">{ticket.title}</p>
  <div class="flex items-center gap-2 flex-wrap">
    <span class="px-2 py-0.5 rounded text-xs font-semibold {priorityClasses[ticket.priority]}">
      {ticket.priority}
    </span>
    {#if epicName}
      <span class="px-2 py-0.5 rounded text-xs bg-purple-100 text-purple-700 dark:bg-purple-900/50 dark:text-purple-300 font-medium">
        {epicName}
      </span>
    {/if}
  </div>
</div>

{#if showDetail}
  <TicketDetail
    ticketId={ticket.id}
    {boardId}
    onclose={() => (showDetail = false)}
    onupdate={(updated: Ticket) => {
      onupdate?.(updated)
      showDetail = false
    }}
    ondelete={(id: string) => {
      ondelete?.(id)
      showDetail = false
    }}
  />
{/if}
