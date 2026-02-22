<script lang="ts">
  import type { Epic } from './types.js'

  interface Props {
    epics: Epic[]
    selectedEpicId: string | null
    onchange: (epicId: string | null) => void
  }

  let { epics, selectedEpicId, onchange }: Props = $props()
</script>

<div class="flex gap-2 flex-wrap items-center">
  <span class="text-sm font-medium text-gray-500 mr-1">Epic:</span>
  <button
    class="px-3 py-1 rounded-full text-sm font-medium transition-colors
      {selectedEpicId === null
        ? 'bg-indigo-600 text-white'
        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'}"
    onclick={() => onchange(null)}
  >
    All
  </button>
  {#each epics as epic (epic.id)}
    <button
      class="px-3 py-1 rounded-full text-sm font-medium transition-colors
        {selectedEpicId === epic.id
          ? 'bg-indigo-600 text-white'
          : 'bg-gray-100 text-gray-700 hover:bg-gray-200'}"
      onclick={() => onchange(epic.id)}
    >
      {epic.title}
    </button>
  {/each}
</div>
