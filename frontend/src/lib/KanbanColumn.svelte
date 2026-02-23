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

  const statusDot: Record<Status, string> = {
    todo:        'bg-gray-400 dark:bg-gray-500',
    in_progress: 'bg-blue-500',
    done:        'bg-emerald-500',
  }

  const statusCount: Record<Status, string> = {
    todo:        'bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400',
    in_progress: 'bg-blue-50 dark:bg-blue-950/50 text-blue-600 dark:text-blue-400',
    done:        'bg-emerald-50 dark:bg-emerald-950/50 text-emerald-600 dark:text-emerald-400',
  }
</script>

<div class="flex flex-col rounded-xl bg-gray-100/70 dark:bg-gray-900/60 border border-gray-200 dark:border-gray-800 min-h-[400px] w-full">
  <!-- Column header -->
  <div class="flex items-center justify-between px-3.5 py-3 border-b border-gray-200 dark:border-gray-800">
    <div class="flex items-center gap-2">
      <span class="w-2 h-2 rounded-full {statusDot[status]} shrink-0"></span>
      <span class="text-xs font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide">{label}</span>
    </div>
    <span class="text-xs font-medium px-1.5 py-0.5 rounded-md {statusCount[status]}">
      {tickets.length}
    </span>
  </div>

  <!-- Drop zone -->
  <div
    class="flex flex-col gap-2 p-2.5 flex-1"
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

    {#if tickets.length === 0}
      <div class="flex-1 flex items-center justify-center min-h-[80px] rounded-lg border-2 border-dashed border-gray-200 dark:border-gray-700/60 mt-1">
        <span class="text-xs text-gray-400 dark:text-gray-600">Drop tickets here</span>
      </div>
    {/if}
  </div>
</div>
