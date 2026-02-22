<script lang="ts">
  import { dndzone } from 'svelte-dnd-action'
  import type { Ticket, Epic, Status } from './types.js'
  import TicketCard from './TicketCard.svelte'

  interface Props {
    status: Status
    label: string
    tickets: Ticket[]
    epics: Epic[]
    boardId: string
    onconsider: (status: Status, tickets: Ticket[]) => void
    onfinalize: (status: Status, tickets: Ticket[]) => void
    onticketupdate: (ticket: Ticket) => void
    onticketdelete: (ticketId: string) => void
  }

  let { status, label, tickets, epics, boardId, onconsider, onfinalize, onticketupdate, onticketdelete }: Props = $props()

  const columnColors: Record<Status, string> = {
    todo: 'bg-gray-50 border-gray-200 dark:bg-gray-800/50 dark:border-gray-700',
    in_progress: 'bg-blue-50 border-blue-200 dark:bg-blue-950/40 dark:border-blue-900',
    done: 'bg-green-50 border-green-200 dark:bg-green-950/40 dark:border-green-900',
  }

  const headerColors: Record<Status, string> = {
    todo: 'text-gray-700 dark:text-gray-300',
    in_progress: 'text-blue-700 dark:text-blue-400',
    done: 'text-green-700 dark:text-green-400',
  }
</script>

<div class="flex flex-col rounded-xl border {columnColors[status]} min-h-[400px] w-full">
  <div class="flex items-center justify-between px-4 py-3 border-b border-inherit">
    <span class="font-semibold text-sm {headerColors[status]}">{label}</span>
    <span class="text-xs font-medium bg-white dark:bg-gray-700 dark:text-gray-300 rounded-full px-2 py-0.5 shadow-sm text-gray-600">
      {tickets.length}
    </span>
  </div>

  <div
    class="flex flex-col gap-2 p-3 flex-1"
    use:dndzone={{ items: tickets, flipDurationMs: 150 }}
    onconsider={(e: CustomEvent<{ items: Ticket[] }>) => onconsider(status, e.detail.items)}
    onfinalize={(e: CustomEvent<{ items: Ticket[] }>) => onfinalize(status, e.detail.items)}
  >
    {#each tickets as ticket (ticket.id)}
      <div>
        <TicketCard
          {ticket}
          {boardId}
          {epics}
          onupdate={onticketupdate}
          ondelete={onticketdelete}
        />
      </div>
    {/each}
  </div>
</div>
