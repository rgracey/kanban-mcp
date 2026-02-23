<script lang="ts">
  import type { Ticket, Epic, Comment, Task, TicketEvent } from "./types.js";
  import {
    getTicket,
    updateTicket,
    deleteTicket,
    listEpics,
    listComments,
    createComment,
    updateComment,
    deleteComment,
    listTasks,
    createTask,
    updateTask,
    deleteTask,
    listTicketEvents,
  } from "./api.js";
  import { toast } from "svelte-sonner";
  import { marked } from "marked";
  import { untrack } from "svelte";

  interface Props {
    ticketId: string;
    boardId: string;
    onclose: () => void;
    onupdate?: (ticket: Ticket) => void;
    ondelete?: (ticketId: string) => void;
  }

  let { ticketId, boardId, onclose, onupdate, ondelete }: Props = $props();

  // --- server data ---
  let ticket = $state<Ticket | null>(null);
  let epics = $state<Epic[]>([]);
  let comments = $state<Comment[]>([]);
  let tasks = $state<Task[]>([]);
  let events = $state<TicketEvent[]>([]);
  let loading = $state(true);
  let loadError = $state("");

  // --- history panel ---
  let showHistory = $state(false);

  // --- draft state (all fields go through Save) ---
  let draftTitle = $state("");
  let draftDescription = $state("");
  let draftAssignee = $state("");
  let draftStatus = $state("");
  let draftPriority = $state("");
  let draftEpicId = $state("");
  let editingDescription = $state(false);

  // --- pending new tasks (saved on Save button) ---
  let pendingTasks = $state<string[]>([]); // titles not yet persisted
  let newTaskTitle = $state("");

  // --- save / delete state ---
  let saving = $state(false);
  let fieldError = $state("");
  let deletingTicket = $state(false);

  // --- comment state ---
  let newCommentBody = $state("");
  let addingComment = $state(false);
  let editingCommentId = $state<string | null>(null);
  let editingCommentBody = $state("");
  let savingComment = $state(false);

  async function load() {
    loading = true;
    loadError = "";
    try {
      const [t, e, c, tk, ev] = await Promise.all([
        getTicket(ticketId),
        listEpics(boardId),
        listComments(ticketId),
        listTasks(ticketId),
        listTicketEvents(ticketId),
      ]);
      ticket = t;
      epics = e;
      comments = c;
      tasks = tk;
      events = ev;
      pendingTasks = [];
      // seed drafts — untrack so assignment doesn't loop
      draftTitle = untrack(() => t.title);
      draftDescription = untrack(() => t.description ?? "");
      draftAssignee = untrack(() => t.assignee ?? "");
      draftStatus = untrack(() => t.status);
      draftPriority = untrack(() => t.priority);
      draftEpicId = untrack(() => t.epic_id ?? "");
    } catch (e) {
      loadError = e instanceof Error ? e.message : "Failed to load ticket";
      toast.error(loadError);
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    ticketId;
    load();
  });

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Escape" && !editingDescription) onclose();
  }

  // dirty check: any draft differs from server value, or there are pending tasks
  const hasDraftChanges = $derived(
    ticket !== null &&
      (draftTitle !== ticket.title ||
        draftDescription !== (ticket.description ?? "") ||
        draftAssignee !== (ticket.assignee ?? "") ||
        draftStatus !== ticket.status ||
        draftPriority !== ticket.priority ||
        draftEpicId !== (ticket.epic_id ?? "") ||
        pendingTasks.length > 0),
  );

  async function saveAll() {
    if (!ticket) return;
    saving = true;
    fieldError = "";
    try {
      // 1. Build a patch containing only fields that actually changed
      const patch: Parameters<typeof updateTicket>[1] = {};
      const trimmedTitle = draftTitle.trim();
      if (trimmedTitle && trimmedTitle !== ticket.title)
        patch.title = trimmedTitle;
      if (draftDescription !== (ticket.description ?? ""))
        patch.description = draftDescription;
      if (draftAssignee !== (ticket.assignee ?? ""))
        patch.assignee = draftAssignee;
      if (draftStatus !== ticket.status)
        patch.status = draftStatus as Ticket["status"];
      if (draftPriority !== ticket.priority)
        patch.priority = draftPriority as Ticket["priority"];
      if ((draftEpicId || null) !== ticket.epic_id)
        patch.epic_id = draftEpicId || null;

      const updated =
        Object.keys(patch).length > 0
          ? await updateTicket(ticket.id, patch)
          : ticket;
      ticket = updated;
      draftTitle = updated.title;
      draftDescription = updated.description ?? "";
      draftAssignee = updated.assignee ?? "";
      draftStatus = updated.status;
      draftPriority = updated.priority;
      draftEpicId = updated.epic_id ?? "";
      editingDescription = false;
      onupdate?.(updated);

      // 2. Create pending tasks in order
      if (pendingTasks.length > 0) {
        const created = await Promise.all(
          pendingTasks.map((title) => createTask(ticket!.id, title)),
        );
        tasks = [...tasks, ...created];
        pendingTasks = [];
      }

      toast.success("Saved");
    } catch (e) {
      const msg = e instanceof Error ? e.message : "Save failed";
      fieldError = msg;
      toast.error(msg);
    } finally {
      saving = false;
    }
  }

  // --- existing tasks (toggle/delete are immediate — these are atomic actions) ---
  async function toggleTask(task: Task) {
    try {
      const updated = await updateTask(task.id, { done: !task.done });
      tasks = tasks.map((t) => (t.id === updated.id ? updated : t));
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to update task");
    }
  }

  async function removeTask(id: string) {
    try {
      await deleteTask(id);
      tasks = tasks.filter((t) => t.id !== id);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to delete task");
    }
  }

  // --- pending task list ---
  function addPendingTask() {
    const title = newTaskTitle.trim();
    if (!title) return;
    pendingTasks = [...pendingTasks, title];
    newTaskTitle = "";
  }

  function removePendingTask(i: number) {
    pendingTasks = pendingTasks.filter((_, idx) => idx !== i);
  }

  function onNewTaskKeydown(e: KeyboardEvent) {
    if (e.key === "Enter") addPendingTask();
  }

  // --- comments ---
  async function submitComment() {
    if (!newCommentBody.trim() || !ticket) return;
    addingComment = true;
    try {
      const comment = await createComment(ticket.id, newCommentBody.trim());
      comments = [...comments, comment];
      newCommentBody = "";
      toast.success("Comment posted");
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to post comment");
    } finally {
      addingComment = false;
    }
  }

  function startEditComment(comment: Comment) {
    editingCommentId = comment.id;
    editingCommentBody = comment.body;
  }

  async function saveCommentEdit() {
    if (!editingCommentId) return;
    savingComment = true;
    try {
      const updated = await updateComment(editingCommentId, editingCommentBody);
      comments = comments.map((c) => (c.id === updated.id ? updated : c));
      editingCommentId = null;
      toast.success("Comment updated");
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to update comment");
    } finally {
      savingComment = false;
    }
  }

  async function removeComment(id: string) {
    try {
      await deleteComment(id);
      comments = comments.filter((c) => c.id !== id);
      toast.success("Comment deleted");
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to delete comment");
    }
  }

  // --- delete ticket ---
  async function removeTicket() {
    if (!ticket || !confirm("Delete this ticket? This cannot be undone."))
      return;
    deletingTicket = true;
    try {
      await deleteTicket(ticket.id);
      toast.success("Ticket deleted");
      ondelete?.(ticket.id);
      onclose();
    } catch (e) {
      const msg = e instanceof Error ? e.message : "Delete failed";
      fieldError = msg;
      toast.error(msg);
    } finally {
      deletingTicket = false;
    }
  }

  function formatDate(iso: string) {
    return new Date(iso).toLocaleString(undefined, {
      dateStyle: "medium",
      timeStyle: "short",
    });
  }

  const eventIcons: Record<string, string> = {
    created:        '✦',
    moved:          '→',
    edited:         '✎',
    commented:      '💬',
    comment_edited: '💬',
    task_added:     '☐',
    task_updated:   '☑',
    task_deleted:   '✕',
  }

  function eventDescription(ev: TicketEvent): string {
    const p = ev.payload ?? {}
    switch (ev.type) {
      case 'created':
        return `Ticket created`
      case 'moved': {
        const from = (p.from as string | undefined)?.replace(/_/g, ' ')
        const to   = (p.to   as string | undefined)?.replace(/_/g, ' ')
        return from && to ? `Moved from ${from} → ${to}` : 'Status changed'
      }
      case 'edited': {
        const fields = Object.keys(p)
        return fields.length ? `Updated ${fields.join(', ')}` : 'Edited'
      }
      case 'commented':
        return 'Comment added'
      case 'comment_edited':
        return 'Comment edited'
      case 'task_added':
        return p.task_title ? `Task added: "${p.task_title}"` : 'Task added'
      case 'task_updated': {
        const title = p.task_title as string | undefined
        const done  = p.done as boolean | undefined
        if (done !== undefined)
          return `Task "${title}" marked ${done ? 'done' : 'not done'}`
        if (p.title)
          return `Task renamed to "${p.title}"`
        return title ? `Task "${title}" updated` : 'Task updated'
      }
      case 'task_deleted':
        return p.task_title ? `Task deleted: "${p.task_title}"` : 'Task deleted'
      default:
        return ev.type
    }
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
  class="fixed top-0 right-0 h-full w-full max-w-lg bg-white dark:bg-gray-900 shadow-2xl z-50 flex flex-col overflow-hidden"
>
  <!-- Header -->
  <div
    class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700 shrink-0"
  >
    <h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">
      Ticket Detail
    </h2>
    <button
      class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-2xl leading-none"
      onclick={onclose}
      aria-label="Close">&times;</button
    >
  </div>

  {#if loading}
    <div class="flex-1 flex items-center justify-center text-gray-400 text-sm">
      Loading...
    </div>
  {:else if loadError}
    <div class="flex-1 flex items-center justify-center text-red-500 text-sm">
      {loadError}
    </div>
  {:else if ticket}
    <div class="flex-1 overflow-y-auto px-6 py-4 space-y-5">
      {#if fieldError}
        <p class="text-red-500 text-sm">{fieldError}</p>
      {/if}

      <!-- Title -->
      <div>
        <label
          for="td-title"
          class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1"
          >Title</label
        >
        <input
          id="td-title"
          class="w-full border border-gray-200 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 rounded-lg px-3 py-2 text-sm text-gray-900 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          bind:value={draftTitle}
        />
      </div>

      <!-- Description -->
      <div>
        <div class="flex items-center justify-between mb-1">
          <label
            for="td-desc"
            class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide"
            >Description</label
          >
          {#if !editingDescription}
            <button
              class="text-xs text-indigo-500 hover:text-indigo-700 dark:text-indigo-400 dark:hover:text-indigo-300"
              onclick={() => (editingDescription = true)}>Edit</button
            >
          {:else}
            <button
              class="text-xs text-gray-400 hover:text-gray-600"
              onclick={() => (editingDescription = false)}>Done</button
            >
          {/if}
        </div>
        {#if editingDescription}
          <textarea
            id="td-desc"
            class="w-full border border-gray-200 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 rounded-lg px-3 py-2 text-sm text-gray-900 focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
            rows="6"
            bind:value={draftDescription}
          ></textarea>
          <p class="text-xs text-gray-400 mt-1">Markdown supported</p>
        {:else}
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <div
            role="button"
            tabindex="0"
            class="md-body dark:text-gray-300 min-h-[3rem] rounded-lg border border-transparent hover:border-gray-200 dark:hover:border-gray-700 px-3 py-2 cursor-text {!draftDescription
              ? 'text-gray-400 italic'
              : ''}"
            onclick={() => (editingDescription = true)}
          >
            {#if draftDescription}
              {@html marked.parse(draftDescription)}
            {:else}
              <span>No description — click to add</span>
            {/if}
          </div>
        {/if}
      </div>

      <!-- Status / Priority / Epic -->
      <div class="grid grid-cols-3 gap-3">
        <div>
          <label
            for="td-status"
            class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1"
            >Status</label
          >
          <select
            id="td-status"
            class="w-full border border-gray-200 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 rounded-lg px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            bind:value={draftStatus}
          >
            <option value="todo">To Do</option>
            <option value="in_progress">In Progress</option>
            <option value="done">Done</option>
          </select>
        </div>

        <div>
          <label
            for="td-priority"
            class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1"
            >Priority</label
          >
          <select
            id="td-priority"
            class="w-full border border-gray-200 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 rounded-lg px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            bind:value={draftPriority}
          >
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
            <option value="critical">Critical</option>
          </select>
        </div>

        <div>
          <label
            for="td-epic"
            class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1"
            >Epic</label
          >
          <select
            id="td-epic"
            class="w-full border border-gray-200 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 rounded-lg px-2 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            bind:value={draftEpicId}
          >
            <option value="">None</option>
            {#each epics as epic (epic.id)}
              <option value={epic.id}>{epic.title}</option>
            {/each}
          </select>
        </div>
      </div>

      <!-- Assignee -->
      <div>
        <label
          for="td-assignee"
          class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1"
          >Assignee</label
        >
        <input
          id="td-assignee"
          class="w-full border border-gray-200 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 rounded-lg px-3 py-2 text-sm text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          placeholder="Unassigned"
          bind:value={draftAssignee}
        />
      </div>

      <!-- Tasks / Checklist -->
      <div>
        <h3
          class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3"
        >
          Tasks
          {#if tasks.length + pendingTasks.length > 0}
            <span class="ml-1 font-normal normal-case">
              ({tasks.filter((t) => t.done).length}/{tasks.length +
                pendingTasks.length})
            </span>
          {/if}
        </h3>

        <!-- Persisted tasks -->
        {#if tasks.length > 0}
          <ul class="space-y-1 mb-2">
            {#each tasks as task (task.id)}
              <li class="flex items-center gap-2 group">
                <input
                  type="checkbox"
                  class="h-4 w-4 rounded border-gray-300 dark:border-gray-600 text-indigo-600 focus:ring-indigo-500 cursor-pointer"
                  checked={task.done}
                  onchange={() => toggleTask(task)}
                />
                <span
                  class="flex-1 text-sm {task.done
                    ? 'line-through text-gray-400 dark:text-gray-600'
                    : 'text-gray-800 dark:text-gray-200'}">{task.title}</span
                >
                <button
                  class="text-gray-300 dark:text-gray-600 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity text-xs"
                  onclick={() => removeTask(task.id)}
                  aria-label="Delete task">&times;</button
                >
              </li>
            {/each}
          </ul>
        {/if}

        <!-- Pending (unsaved) tasks -->
        {#if pendingTasks.length > 0}
          <ul class="space-y-1 mb-2">
            {#each pendingTasks as title, i}
              <li class="flex items-center gap-2 group">
                <input
                  type="checkbox"
                  class="h-4 w-4 rounded border-gray-300 dark:border-gray-600 opacity-40 cursor-not-allowed"
                  disabled
                />
                <span
                  class="flex-1 text-sm text-gray-500 dark:text-gray-400 italic"
                  >{title}</span
                >
                <span class="text-xs text-gray-400 dark:text-gray-500 mr-1"
                  >unsaved</span
                >
                <button
                  class="text-gray-300 dark:text-gray-600 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity text-xs"
                  onclick={() => removePendingTask(i)}
                  aria-label="Remove pending task">&times;</button
                >
              </li>
            {/each}
          </ul>
        {/if}

        <!-- Add task input -->
        <div class="flex gap-2">
          <input
            class="flex-1 border border-gray-200 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
            placeholder="Add a task…"
            bind:value={newTaskTitle}
            onkeydown={onNewTaskKeydown}
          />
          <button
            class="px-3 py-1.5 text-sm bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 disabled:opacity-50 transition-colors"
            onclick={addPendingTask}
            disabled={!newTaskTitle.trim()}>Add</button
          >
        </div>
      </div>

      <!-- History -->
      <div>
        <button
          class="flex items-center gap-1.5 w-full text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3 hover:text-gray-700 dark:hover:text-gray-300 transition-colors"
          onclick={() => (showHistory = !showHistory)}
          aria-expanded={showHistory}
        >
          <svg
            class="w-3 h-3 transition-transform {showHistory
              ? 'rotate-90'
              : ''}"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="3"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M9 5l7 7-7 7"
            />
          </svg>
          History
          {#if events.length > 0}
            <span class="ml-1 font-normal normal-case">({events.length})</span>
          {/if}
        </button>

        {#if showHistory}
          {#if events.length === 0}
            <p class="text-sm text-gray-400 italic mb-2">No history yet.</p>
          {:else}
            <ol
              class="relative border-l border-gray-200 dark:border-gray-700 ml-2 space-y-3 mb-2"
            >
              {#each [...events].reverse() as ev (ev.id)}
                <li class="ml-4">
                  <span
                    class="absolute -left-2 flex h-4 w-4 items-center justify-center rounded-full bg-gray-100 dark:bg-gray-800 ring-2 ring-white dark:ring-gray-900 text-xs"
                  >
                    {eventIcons[ev.type] ?? "·"}
                  </span>
                  <div class="flex flex-col">
                    <span class="text-xs font-medium text-gray-700 dark:text-gray-300">
                      {eventDescription(ev)}
                      {#if ev.actor}
                        <span class="font-normal text-gray-500 dark:text-gray-400"> by {ev.actor}</span>
                      {/if}
                    </span>
                    <time class="text-xs text-gray-400">{formatDate(ev.created_at)}</time>
                  </div>
                </li>
              {/each}
            </ol>
          {/if}
        {/if}
      </div>

      <!-- Comments -->
      <div>
        <h3
          class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3"
        >
          Comments
        </h3>

        {#if comments.length === 0}
          <p class="text-sm text-gray-400 italic mb-3">No comments yet.</p>
        {:else}
          <div class="space-y-3 mb-4">
            {#each comments as comment (comment.id)}
              <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
                {#if editingCommentId === comment.id}
                  <textarea
                    class="w-full border border-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 rounded px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none mb-2"
                    rows="3"
                    bind:value={editingCommentBody}
                  ></textarea>
                  <div class="flex gap-2">
                    <button
                      class="px-3 py-1 text-xs bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
                      onclick={saveCommentEdit}
                      disabled={savingComment}
                      >{savingComment ? "Saving..." : "Save"}</button
                    >
                    <button
                      class="px-3 py-1 text-xs text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 rounded"
                      onclick={() => (editingCommentId = null)}>Cancel</button
                    >
                  </div>
                {:else}
                  <p
                    class="text-sm text-gray-800 dark:text-gray-200 whitespace-pre-wrap"
                  >
                    {comment.body}
                  </p>
                  <div class="flex items-center justify-between mt-2">
                    <span class="text-xs text-gray-400"
                      >{formatDate(comment.created_at)}</span
                    >
                    <div class="flex gap-2">
                      <button
                        class="text-xs text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400"
                        onclick={() => startEditComment(comment)}>Edit</button
                      >
                      <button
                        class="text-xs text-gray-400 hover:text-red-600"
                        onclick={() => removeComment(comment.id)}>Delete</button
                      >
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
            class="w-full border border-gray-200 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-100 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
            rows="2"
            placeholder="Add a comment…"
            bind:value={newCommentBody}
          ></textarea>
          <button
            class="px-4 py-1.5 text-sm bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition-colors"
            onclick={submitComment}
            disabled={addingComment || !newCommentBody.trim()}
            >{addingComment ? "Posting..." : "Post comment"}</button
          >
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div
      class="shrink-0 px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-between items-center"
    >
      <button
        class="px-3 py-2 text-sm text-red-600 hover:bg-red-50 dark:hover:bg-red-950/30 rounded-lg transition-colors disabled:opacity-50"
        onclick={removeTicket}
        disabled={deletingTicket}
        >{deletingTicket ? "Deleting..." : "Delete ticket"}</button
      >

      <button
        class="px-5 py-2 text-sm font-medium bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 transition-colors {hasDraftChanges
          ? ''
          : 'opacity-60'}"
        onclick={saveAll}
        disabled={saving || !hasDraftChanges}
        >{saving ? "Saving..." : "Save"}</button
      >
    </div>
  {/if}
</div>
