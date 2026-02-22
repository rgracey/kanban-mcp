<script lang="ts">
  import type { Ticket, Epic, Status } from './types.js'
  import { listTickets, listEpics, updateTicket } from './api.js'
  import EpicFilter from './EpicFilter.svelte'
  import KanbanColumn from './KanbanColumn.svelte'
  import CreateTicket from './CreateTicket.svelte'

  interface Props {
    boardId: string
  }

  let { boardId }: Props = $props()

  let tickets = $state<Ticket[]>([])
  let epics = $state<Epic[]>([])
  let selectedEpicId = $state<string | null>(null)
  let loading = $state(true)
  let error = $state('')
  let showCreateTicket = $state(false)

  const columns: { status: Status; label: string }[] = [
    { status: 'todo', label: 'To Do' },
    { status: 'in_progress', label: 'In Progress' },
    { status: 'done', label: 'Done' },
  ]

  async function load() {
    loading = true
    error = ''
    try {
      ;[tickets, epics] = await Promise.all([listTickets(boardId), listEpics(boardId)])
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load board'
    } finally {
      loading = false
    }
  }

  $effect(() => {
    boardId
    load()
  })

  const filteredTickets = $derived(
    selectedEpicId ? tickets.filter((t) => t.epic_id === selectedEpicId) : tickets
  )

  function ticketsForStatus(status: Status): Ticket[] {
    return filteredTickets.filter((t) => t.status === status)
  }

  /**
   * During drag (consider), update local state optimistically so the UI responds.
   */
  function handleConsider(status: Status, items: Ticket[]) {
    const others = tickets.filter((t) => t.status !== status)
    tickets = [...others, ...items.map((t) => ({ ...t, status }))]
  }

  /**
   * On drop (finalize), persist any status changes for tickets in this column.
   */
  async function handleFinalize(status: Status, items: Ticket[]) {
    // Optimistically commit
    const others = tickets.filter((t) => t.status !== status)
    tickets = [...others, ...items.map((t) => ({ ...t, status }))]

    // Persist tickets that actually changed status
    for (const t of items) {
      if (t.status !== status) {
        try {
          const saved = await updateTicket(t.id, { status })
          tickets = tickets.map((existing) => (existing.id === saved.id ? saved : existing))
        } catch {
          await load()
          return
        }
      }
    }
  }

  function handleTicketUpdate(updated: Ticket) {
    tickets = tickets.map((t) => (t.id === updated.id ? updated : t))
  }

  function handleTicketDelete(ticketId: string) {
    tickets = tickets.filter((t) => t.id !== ticketId)
  }

  function handleTicketCreate(ticket: Ticket) {
    tickets = [...tickets, ticket]
    showCreateTicket = false
  }
</script>

<div class="flex flex-col gap-4 h-full">
  <div class="flex items-center justify-between flex-wrap gap-2">
    <EpicFilter {epics} {selectedEpicId} onchange={(id) => (selectedEpicId = id)} />
    <button
      class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors"
      onclick={() => (showCreateTicket = true)}
    >
      + New Ticket
    </button>
  </div>

  {#if loading}
    <div class="flex items-center justify-center h-64 text-gray-400">Loading...</div>
  {:else if error}
    <div class="flex items-center justify-center h-64 text-red-500">{error}</div>
  {:else}
    <div class="grid grid-cols-3 gap-4 flex-1">
      {#each columns as col (col.status)}
        <KanbanColumn
          status={col.status}
          label={col.label}
          tickets={ticketsForStatus(col.status)}
          {epics}
          onconsider={handleConsider}
          onfinalize={handleFinalize}
          onticketupdate={handleTicketUpdate}
          onticketdelete={handleTicketDelete}
        />
      {/each}
    </div>
  {/if}
</div>

{#if showCreateTicket}
  <CreateTicket
    {boardId}
    {epics}
    onclose={() => (showCreateTicket = false)}
    oncreate={handleTicketCreate}
  />
{/if}
