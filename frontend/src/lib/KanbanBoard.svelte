<script lang="ts">
  import { onDestroy } from 'svelte'
  import { toast } from 'svelte-sonner'
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
      toast.error(error)
    } finally {
      loading = false
    }
  }

  $effect(() => {
    boardId
    load()
  })

  // SSE: reload board when a change event arrives for this board
  let es: EventSource | null = null

  $effect(() => {
    const currentBoardId = boardId
    if (es) { es.close(); es = null }

    es = new EventSource('/api/v1/events')
    es.addEventListener('board_change', (e: MessageEvent) => {
      try {
        const ev = JSON.parse(e.data) as { type: string; board_id: string }
        if (ev.board_id === currentBoardId) {
          load()
        }
      } catch { /* ignore parse errors */ }
    })
    es.onerror = () => {
      // browser will auto-reconnect; nothing to do
    }
  })

  onDestroy(() => {
    es?.close()
    es = null
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
          const label = { todo: 'To Do', in_progress: 'In Progress', done: 'Done' }[status]
          toast.success(`Moved "${saved.title}" to ${label}`)
        } catch (e) {
          toast.error(e instanceof Error ? e.message : 'Failed to move ticket')
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
    toast.success(`Ticket "${ticket.title}" created`)
  }

  function handleEpicCreate(epic: Epic) {
    epics = [...epics, epic]
    showCreateEpic = false
    toast.success(`Epic "${epic.title}" created`)
  }
</script>

<div class="flex flex-col gap-4 h-full">
  <!-- Toolbar -->
  <div class="flex items-center justify-between gap-3 flex-wrap">
    <!-- Left: epic filter -->
    <div class="flex items-center gap-2 min-w-0">
      <span class="text-xs font-medium text-gray-400 dark:text-gray-500 shrink-0">Epic</span>
      <EpicFilter {epics} {selectedEpicId} onchange={(id) => (selectedEpicId = id)} />
    </div>
    <!-- Right: action buttons -->
    <div class="flex items-center gap-2 shrink-0">
      <button
        class="flex items-center gap-1.5 px-3 py-1.5 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 text-sm font-medium rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700/70 hover:border-gray-300 dark:hover:border-gray-600 transition-colors"
        onclick={() => (showCreateEpic = true)}
      >
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
        </svg>
        Epic
      </button>
      <button
        class="flex items-center gap-1.5 px-3 py-1.5 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors shadow-sm"
        onclick={() => (showCreateTicket = true)}
      >
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
        </svg>
        New Ticket
      </button>
    </div>
  </div>

  {#if loading}
    <div class="flex items-center justify-center h-64 text-gray-400 text-sm">Loading...</div>
  {:else if error}
    <div class="flex items-center justify-center h-64 text-red-500 text-sm">{error}</div>
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
