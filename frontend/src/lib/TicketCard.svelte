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
    low: 'bg-gray-100 text-gray-600',
    medium: 'bg-blue-100 text-blue-700',
    high: 'bg-orange-100 text-orange-700',
    critical: 'bg-red-100 text-red-700',
  }

  const epicName = $derived(
    ticket.epic_id ? epics.find((e) => e.id === ticket.epic_id)?.title ?? null : null
  )
</script>

<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
<div
  role="button"
  tabindex="0"
  class="bg-white rounded-lg shadow-sm border border-gray-200 p-3 cursor-pointer hover:shadow-md transition-shadow select-none"
  onclick={() => (showDetail = true)}
  onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') showDetail = true }}
>
  <p class="text-sm font-medium text-gray-900 mb-2 leading-snug">{ticket.title}</p>
  <div class="flex items-center gap-2 flex-wrap">
    <span class="px-2 py-0.5 rounded text-xs font-semibold {priorityClasses[ticket.priority]}">
      {ticket.priority}
    </span>
    {#if epicName}
      <span class="px-2 py-0.5 rounded text-xs bg-purple-100 text-purple-700 font-medium">
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
