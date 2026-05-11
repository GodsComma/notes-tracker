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

        <!-- Hover actions: Focus + Settings -->
        <div class="absolute top-2 right-2 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <button
            type="button"
            class="text-[11px] font-medium px-2 py-1 rounded-full bg-white/90 dark:bg-slate-900/90 text-indigo-600 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-800 hover:bg-indigo-600 hover:text-white hover:border-indigo-600 transition-colors"
            title="Start focus session"
            @click.stop="open(item)"
          >
            ▶ Focus
          </button>
          <button
            type="button"
            class="p-1.5 rounded-full bg-white/90 dark:bg-slate-900/90 text-slate-500 dark:text-slate-400 border border-slate-200 dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            title="Focus session settings"
            @click.stop="focusStore.openSettings(item, '/library')"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
          </button>
        </div>
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
