<script lang="ts">
	interface Props {
		id: string;
		/** How many trailing chars of the ID to show. Default 8. */
		truncate?: number;
	}

	let { id, truncate = 8 }: Props = $props();

	let copied = $state(false);
	let timer: ReturnType<typeof setTimeout> | null = null;

	async function copy(e: MouseEvent) {
		e.stopPropagation();
		try {
			await navigator.clipboard.writeText(id);
			copied = true;
			if (timer) clearTimeout(timer);
			timer = setTimeout(() => {
				copied = false;
				timer = null;
			}, 1500);
		} catch {
			// clipboard not available — silently ignore
		}
	}
</script>

<button
	class="inline-flex items-center gap-0.5 rounded px-1 py-0.5 font-mono text-[10px] transition-colors
    {copied
		? 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-400'
		: 'bg-gray-100 text-gray-400 hover:bg-gray-200 hover:text-gray-600 dark:bg-gray-800 dark:text-gray-500 dark:hover:bg-gray-700 dark:hover:text-gray-300'}"
	onclick={copy}
	title="Click to copy ID: {id}"
	aria-label="Copy ID"
>
	{#if copied}
		Copied!
	{:else}
		…{id.slice(-truncate)}
	{/if}
</button>
