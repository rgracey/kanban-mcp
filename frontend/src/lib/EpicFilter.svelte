<script lang="ts">
  import type { Epic } from './types.js'

  interface Props {
    epics: Epic[]
    selectedEpicId: string | null
    onchange: (epicId: string | null) => void
  }

  let { epics, selectedEpicId, onchange }: Props = $props()
</script>

<div class="flex gap-1.5 flex-wrap items-center">
  <button
    class="px-2.5 py-1 rounded-md text-xs font-medium transition-colors
      {selectedEpicId === null
        ? 'bg-indigo-600 text-white shadow-sm'
        : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-800 dark:hover:text-gray-200'}"
    onclick={() => onchange(null)}
  >
    All
  </button>
  {#each epics as epic (epic.id)}
    <button
      class="px-2.5 py-1 rounded-md text-xs font-medium transition-colors
        {selectedEpicId === epic.id
          ? 'bg-indigo-600 text-white shadow-sm'
          : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 hover:text-gray-800 dark:hover:text-gray-200'}"
      onclick={() => onchange(epic.id)}
    >
      {epic.title}
    </button>
  {/each}
</div>
