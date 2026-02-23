<script lang="ts">
	import type { Ticket, Epic } from './types.js';
	import TicketDetail from './TicketDetail.svelte';

	interface Props {
		ticket: Ticket;
		boardId: string;
		epics: Epic[];
		onupdate?: (ticket: Ticket) => void;
		ondelete?: (ticketId: string) => void;
	}

	let { ticket, boardId, epics, onupdate, ondelete }: Props = $props();

	let showDetail = $state(false);

	const priorityConfig: Record<string, { label: string; classes: string }> = {
		low: {
			label: 'Low',
			classes: 'bg-gray-100 text-gray-500 dark:bg-gray-700/60 dark:text-gray-400'
		},
		medium: {
			label: 'Med',
			classes: 'bg-sky-100 text-sky-700 dark:bg-sky-900/40 dark:text-sky-400'
		},
		high: {
			label: 'High',
			classes: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-400'
		},
		critical: {
			label: 'Crit',
			classes: 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400'
		}
	};

	const statusAccent: Record<string, string> = {
		todo: 'border-l-gray-300 dark:border-l-gray-600',
		in_progress: 'border-l-blue-400 dark:border-l-blue-500',
		done: 'border-l-emerald-400 dark:border-l-emerald-500'
	};

	const epicName = $derived(
		ticket.epic_id ? (epics.find((e) => e.id === ticket.epic_id)?.title ?? null) : null
	);

	function getInitials(name: string): string {
		return name
			.split(/\s+/)
			.filter(Boolean)
			.slice(0, 2)
			.map((p) => p[0].toUpperCase())
			.join('');
	}
</script>

<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
<div
	role="button"
	tabindex="0"
	class="rounded-lg border border-l-4 border-gray-200 bg-white shadow-sm dark:border-gray-700/80 dark:bg-gray-800 {statusAccent[
		ticket.status
	]} cursor-pointer p-3 transition-all select-none hover:border-gray-300 hover:shadow-md dark:hover:border-gray-600"
	onclick={() => (showDetail = true)}
	onkeydown={(e) => {
		if (e.key === 'Enter' || e.key === ' ') showDetail = true;
	}}
>
	<!-- Title -->
	<p class="mb-2.5 text-sm leading-snug font-medium text-gray-900 dark:text-gray-100">
		{ticket.title}
	</p>

	<!-- Footer row: badges + assignee -->
	<div class="flex items-center justify-between gap-2">
		<div class="flex min-w-0 flex-wrap items-center gap-1.5">
			<!-- Priority -->
			<span
				class="rounded px-1.5 py-0.5 text-[11px] font-semibold {priorityConfig[ticket.priority]
					?.classes ?? ''}"
			>
				{priorityConfig[ticket.priority]?.label ?? ticket.priority}
			</span>
			<!-- Epic -->
			{#if epicName}
				<span
					class="max-w-[120px] truncate rounded bg-violet-100 px-1.5 py-0.5 text-[11px] font-medium text-violet-700 dark:bg-violet-900/40 dark:text-violet-300"
				>
					{epicName}
				</span>
			{/if}
		</div>

		<!-- Assignee avatar -->
		{#if ticket.assignee}
			<div
				class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-indigo-100 text-[10px] font-semibold text-indigo-700 dark:bg-indigo-900/50 dark:text-indigo-300"
				title={ticket.assignee}
			>
				{getInitials(ticket.assignee)}
			</div>
		{/if}
	</div>
</div>

{#if showDetail}
	<TicketDetail
		ticketId={ticket.id}
		{boardId}
		onclose={() => (showDetail = false)}
		onupdate={(updated: Ticket) => {
			onupdate?.(updated);
			showDetail = false;
		}}
		ondelete={(id: string) => {
			ondelete?.(id);
			showDetail = false;
		}}
	/>
{/if}
