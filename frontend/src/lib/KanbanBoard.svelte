<script lang="ts">
  import type { Ticket, Epic, Status } from './types.js'
  import { listTickets, listEpics, updateTicket } from './api.js'
  import EpicFilter from './EpicFilter.svelte'
  import KanbanColumn from './KanbanColumn.svelte'
  import CreateTicket from './modals/CreateTicket.svelte'
  import CreateEpic from './modals/CreateEpic.svelte'

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
  let showCreateEpic = $state(false)

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

  function handleConsider(status: Status, items: Ticket[]) {
    const others = tickets.filter((t) => t.status !== status)
    tickets = [...others, ...items.map((t) => ({ ...t, status }))]
  }

  async function handleFinalize(status: Status, items: Ticket[]) {
    const others = tickets.filter((t) => t.status !== status)
    tickets = [...others, ...items.map((t) => ({ ...t, status }))]

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

  function handleEpicCreate(epic: Epic) {
    epics = [...epics, epic]
    showCreateEpic = false
  }
</script>

<div class="flex flex-col gap-4 h-full">
  <div class="flex items-center justify-between flex-wrap gap-2">
    <EpicFilter {epics} {selectedEpicId} onchange={(id) => (selectedEpicId = id)} />
    <div class="flex gap-2">
      <button
        class="px-4 py-2 bg-white border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50 transition-colors"
        onclick={() => (showCreateEpic = true)}
      >
        + New Epic
      </button>
      <button
        class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors"
        onclick={() => (showCreateTicket = true)}
      >
        + New Ticket
      </button>
    </div>
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
          {boardId}
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
    onclose={() => (showCreateTicket = false)}
    oncreated={handleTicketCreate}
  />
{/if}

{#if showCreateEpic}
  <CreateEpic
    {boardId}
    onclose={() => (showCreateEpic = false)}
    oncreated={handleEpicCreate}
  />
{/if}
