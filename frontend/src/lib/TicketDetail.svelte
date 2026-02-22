<script lang="ts">
  import type { Ticket, Epic, Comment } from './types.js'
  import {
    getTicket,
    updateTicket,
    deleteTicket,
    listEpics,
    listComments,
    createComment,
    updateComment,
    deleteComment,
  } from './api.js'
  import { toast } from 'svelte-sonner'

  interface Props {
    ticketId: string
    boardId: string
    onclose: () => void
    onupdate?: (ticket: Ticket) => void
    ondelete?: (ticketId: string) => void
  }

  let { ticketId, boardId, onclose, onupdate, ondelete }: Props = $props()

  // --- data ---
  let ticket = $state<Ticket | null>(null)
  let epics = $state<Epic[]>([])
  let comments = $state<Comment[]>([])
  let loading = $state(true)
  let loadError = $state('')

  // --- inline field editing ---
  let saving = $state<Record<string, boolean>>({})
  let fieldError = $state('')

  // --- comment form ---
  let newCommentBody = $state('')
  let addingComment = $state(false)

  // --- per-comment edit state ---
  let editingCommentId = $state<string | null>(null)
  let editingCommentBody = $state('')
  let savingComment = $state(false)

  // --- delete ticket ---
  let deletingTicket = $state(false)

  async function load() {
    loading = true
    loadError = ''
    try {
      ;[ticket, epics, comments] = await Promise.all([
        getTicket(ticketId),
        listEpics(boardId),
        listComments(ticketId),
      ])
    } catch (e) {
      loadError = e instanceof Error ? e.message : 'Failed to load ticket'
      toast.error(loadError)
    } finally {
      loading = false
    }
  }

  $effect(() => {
    ticketId
    load()
  })

  // Escape key to close
  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') onclose()
  }

  async function saveField(field: keyof Pick<Ticket, 'title' | 'description' | 'status' | 'priority' | 'epic_id'>, value: string | null) {
    if (!ticket) return
    saving = { ...saving, [field]: true }
    fieldError = ''
    try {
      const updated = await updateTicket(ticket.id, { [field]: value })
      ticket = updated
      onupdate?.(updated)
      toast.success('Saved')
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Save failed'
      fieldError = msg
      toast.error(msg)
    } finally {
      saving = { ...saving, [field]: false }
    }
  }

  function onTitleBlur(e: FocusEvent) {
    const val = (e.currentTarget as HTMLInputElement).value.trim()
    if (ticket && val && val !== ticket.title) saveField('title', val)
  }

  function onTitleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') (e.currentTarget as HTMLInputElement).blur()
  }

  function onDescBlur(e: FocusEvent) {
    const val = (e.currentTarget as HTMLTextAreaElement).value
    if (ticket && val !== (ticket.description ?? '')) saveField('description', val)
  }

  function onStatusChange(e: Event) {
    const val = (e.currentTarget as HTMLSelectElement).value
    saveField('status', val)
  }

  function onPriorityChange(e: Event) {
    const val = (e.currentTarget as HTMLSelectElement).value
    saveField('priority', val)
  }

  function onEpicChange(e: Event) {
    const val = (e.currentTarget as HTMLSelectElement).value
    saveField('epic_id', val || null)
  }

  // --- comments ---
  async function submitComment() {
    if (!newCommentBody.trim() || !ticket) return
    addingComment = true
    try {
      const comment = await createComment(ticket.id, newCommentBody.trim())
      comments = [...comments, comment]
      newCommentBody = ''
      toast.success('Comment posted')
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to post comment')
    } finally {
      addingComment = false
    }
  }

  function startEditComment(comment: Comment) {
    editingCommentId = comment.id
    editingCommentBody = comment.body
  }

  async function saveCommentEdit() {
    if (!editingCommentId) return
    savingComment = true
    try {
      const updated = await updateComment(editingCommentId, editingCommentBody)
      comments = comments.map((c) => (c.id === updated.id ? updated : c))
      editingCommentId = null
      toast.success('Comment updated')
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to update comment')
    } finally {
      savingComment = false
    }
  }

  async function removeComment(id: string) {
    try {
      await deleteComment(id)
      comments = comments.filter((c) => c.id !== id)
      toast.success('Comment deleted')
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to delete comment')
    }
  }

  // --- delete ticket ---
  async function removeTicket() {
    if (!ticket || !confirm('Delete this ticket? This cannot be undone.')) return
    deletingTicket = true
    try {
      await deleteTicket(ticket.id)
      toast.success('Ticket deleted')
      ondelete?.(ticket.id)
      onclose()
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Delete failed'
      fieldError = msg
      toast.error(msg)
    } finally {
      deletingTicket = false
    }
  }

  function formatDate(iso: string) {
    return new Date(iso).toLocaleString(undefined, {
      dateStyle: 'medium',
      timeStyle: 'short',
    })
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- Backdrop -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<div
  role="presentation"
  class="fixed inset-0 bg-black/40 z-40"
  onclick={onclose}
></div>

<!-- Slide-over panel -->
<div
  role="dialog"
  aria-modal="true"
  tabindex="-1"
  class="fixed top-0 right-0 h-full w-full max-w-lg bg-white shadow-2xl z-50 flex flex-col overflow-hidden"
>
  <!-- Header -->
  <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 shrink-0">
    <h2 class="text-base font-semibold text-gray-900">Ticket Detail</h2>
    <button
      class="text-gray-400 hover:text-gray-600 text-2xl leading-none"
      onclick={onclose}
      aria-label="Close"
    >&times;</button>
  </div>

  {#if loading}
    <div class="flex-1 flex items-center justify-center text-gray-400 text-sm">Loading...</div>
  {:else if loadError}
    <div class="flex-1 flex items-center justify-center text-red-500 text-sm">{loadError}</div>
  {:else if ticket}
    <div class="flex-1 overflow-y-auto px-6 py-4 space-y-5">

      {#if fieldError}
        <p class="text-red-500 text-sm">{fieldError}</p>
      {/if}

      <!-- Title -->
      <div>
        <label for="td-title" class="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">Title</label>
        <input
          id="td-title"
          class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm text-gray-900 focus:outline-none focus:ring-2 focus:ring-indigo-500 {saving['title'] ? 'opacity-50' : ''}"
          value={ticket.title}
          onblur={onTitleBlur}
          onkeydown={onTitleKeydown}
        />
      </div>

      <!-- Description -->
      <div>
        <label for="td-desc" class="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">Description</label>
        <textarea
          id="td-desc"
          class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm text-gray-900 focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none {saving['description'] ? 'opacity-50' : ''}"
          rows="4"
          value={ticket.description ?? ''}
          onblur={onDescBlur}
        ></textarea>
      </div>

      <!-- Status / Priority / Epic -->
      <div class="grid grid-cols-3 gap-3">
        <div>
          <label for="td-status" class="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">Status</label>
          <select
            id="td-status"
            class="w-full border border-gray-200 rounded-lg px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            value={ticket.status}
            onchange={onStatusChange}
            disabled={!!saving['status']}
          >
            <option value="todo">To Do</option>
            <option value="in_progress">In Progress</option>
            <option value="done">Done</option>
          </select>
        </div>

        <div>
          <label for="td-priority" class="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">Priority</label>
          <select
            id="td-priority"
            class="w-full border border-gray-200 rounded-lg px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            value={ticket.priority}
            onchange={onPriorityChange}
            disabled={!!saving['priority']}
          >
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
            <option value="critical">Critical</option>
          </select>
        </div>

        <div>
          <label for="td-epic" class="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">Epic</label>
          <select
            id="td-epic"
            class="w-full border border-gray-200 rounded-lg px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            value={ticket.epic_id ?? ''}
            onchange={onEpicChange}
            disabled={!!saving['epic_id']}
          >
            <option value="">None</option>
            {#each epics as epic (epic.id)}
              <option value={epic.id}>{epic.title}</option>
            {/each}
          </select>
        </div>
      </div>

      <!-- Comments -->
      <div>
        <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-3">Comments</h3>

        {#if comments.length === 0}
          <p class="text-sm text-gray-400 italic mb-3">No comments yet.</p>
        {:else}
          <div class="space-y-3 mb-4">
            {#each comments as comment (comment.id)}
              <div class="bg-gray-50 rounded-lg p-3">
                {#if editingCommentId === comment.id}
                  <textarea
                    class="w-full border border-gray-300 rounded px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none mb-2"
                    rows="3"
                    bind:value={editingCommentBody}
                  ></textarea>
                  <div class="flex gap-2">
                    <button
                      class="px-3 py-1 text-xs bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
                      onclick={saveCommentEdit}
                      disabled={savingComment}
                    >{savingComment ? 'Saving...' : 'Save'}</button>
                    <button
                      class="px-3 py-1 text-xs text-gray-600 hover:bg-gray-200 rounded"
                      onclick={() => (editingCommentId = null)}
                    >Cancel</button>
                  </div>
                {:else}
                  <p class="text-sm text-gray-800 whitespace-pre-wrap">{comment.body}</p>
                  <div class="flex items-center justify-between mt-2">
                    <span class="text-xs text-gray-400">{formatDate(comment.created_at)}</span>
                    <div class="flex gap-2">
                      <button
                        class="text-xs text-gray-400 hover:text-indigo-600"
                        onclick={() => startEditComment(comment)}
                      >Edit</button>
                      <button
                        class="text-xs text-gray-400 hover:text-red-600"
                        onclick={() => removeComment(comment.id)}
                      >Delete</button>
                    </div>
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        {/if}

        <!-- Add comment -->
        <div class="space-y-2">
          <textarea
            class="w-full border border-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
            rows="2"
            placeholder="Add a comment…"
            bind:value={newCommentBody}
          ></textarea>
          <button
            class="px-4 py-1.5 text-sm bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition-colors"
            onclick={submitComment}
            disabled={addingComment || !newCommentBody.trim()}
          >{addingComment ? 'Posting...' : 'Post comment'}</button>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="shrink-0 px-6 py-4 border-t border-gray-200 flex justify-between items-center">
      <button
        class="px-3 py-2 text-sm text-red-600 hover:bg-red-50 rounded-lg transition-colors disabled:opacity-50"
        onclick={removeTicket}
        disabled={deletingTicket}
      >{deletingTicket ? 'Deleting...' : 'Delete ticket'}</button>
    </div>
  {/if}
</div>
