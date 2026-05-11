<template>
  <div class="p-8">
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">Library</h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">
          Every PDF across your classes. Click any card to start a focus session.
        </p>
      </div>
    </div>

    <div v-if="loading" class="text-slate-400 text-sm">Loading shelf…</div>

    <div v-else-if="items.length === 0" class="text-center py-24 text-slate-400 dark:text-slate-500">
      <p class="text-5xl mb-4">📚</p>
      <p class="text-lg font-medium">No PDFs in your library yet</p>
      <p class="text-sm mt-1">Attach a PDF to any class or source to see it here.</p>
    </div>

    <div v-else class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-5">
      <div
        v-for="item in items"
        :key="item.id"
        class="group bg-white dark:bg-slate-800 rounded-xl overflow-hidden border border-slate-200 dark:border-slate-700 hover:shadow-md hover:border-indigo-300 dark:hover:border-indigo-500 transition-all relative"
      >
        <button type="button" class="block w-full text-left" @click="open(item)">
          <DeepReadCover :url="api.attachments.fileUrl(item.id, 'inline')" />
          <div class="p-3">
            <div class="flex items-center gap-1.5 mb-1">
              <span class="w-2 h-2 rounded-full shrink-0" :style="{ backgroundColor: item.class_color || '#6366f1' }" />
              <span class="text-[11px] text-slate-500 dark:text-slate-400 truncate">{{ item.class_name }}</span>
            </div>
            <p class="text-sm font-medium text-slate-900 dark:text-slate-100 line-clamp-2">{{ item.filename }}</p>

            <div v-if="progressOf(item.id)" class="mt-2">
              <div class="h-1 rounded-full bg-slate-200 dark:bg-slate-700 overflow-hidden">
                <div class="h-full bg-indigo-500" :style="{ width: progressPct(item.id) + '%' }" />
              </div>
              <p class="text-[10px] text-slate-500 dark:text-slate-400 mt-1">
                Page {{ progressOf(item.id)?.lastPage }} of {{ progressOf(item.id)?.totalPages || '–' }}
              </p>
            </div>
          </div>
        </button>

        <!-- Secondary action: Focus Read pill -->
        <button
          type="button"
          class="absolute top-2 right-2 text-[11px] font-medium px-2 py-1 rounded-full bg-white/90 dark:bg-slate-900/90 text-indigo-600 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-800 opacity-0 group-hover:opacity-100 transition-opacity hover:bg-indigo-600 hover:text-white hover:border-indigo-600"
          title="Start focus session"
          @click.stop="open(item)"
        >
          ▶ Focus
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type PDFLibraryItem } from '@/api'
import DeepReadCover from '@/components/DeepReadCover.vue'
import { useFocusStore } from '@/stores/focus'
import { loadProgress, type DeepReadProgress } from '@/composables/useDeepRead'

const focusStore = useFocusStore()

const items = ref<PDFLibraryItem[]>([])
const loading = ref(true)
const progressCache = ref<Map<number, DeepReadProgress>>(new Map())

onMounted(async () => {
  try {
    items.value = await api.library.pdfs()
    rebuildProgressCache()
  } finally {
    loading.value = false
  }
})

// Refresh progress cache when a session ends and we land back here
watch(() => focusStore.active, (active) => {
  if (!active) rebuildProgressCache()
})

function rebuildProgressCache() {
  const map = new Map<number, DeepReadProgress>()
  for (const it of items.value) {
    const p = loadProgress(it.id)
    if (p) map.set(it.id, p)
  }
  progressCache.value = map
}

function progressOf(id: number): DeepReadProgress | undefined {
  return progressCache.value.get(id)
}

function progressPct(id: number) {
  const p = progressCache.value.get(id)
  if (!p || !p.totalPages) return 0
  return Math.min(100, Math.round((p.lastPage / p.totalPages) * 100))
}

function open(item: PDFLibraryItem) {
  focusStore.requestPreflight(item, '/library')
}
</script>
