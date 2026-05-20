<template>
  <div
    class="p-8 min-h-full relative"
    @dragenter.prevent="onDragEnter"
    @dragleave.prevent="onDragLeave"
    @dragover.prevent
    @drop.prevent="onDrop"
  >
    <!-- Full-page drag overlay -->
    <Transition name="fade">
      <div
        v-if="isDragging"
        class="fixed inset-0 z-40 pointer-events-none flex items-center justify-center"
      >
        <div class="absolute inset-4 rounded-2xl border-2 border-dashed border-indigo-400 bg-indigo-50/80 dark:bg-indigo-950/80" />
        <p class="relative z-10 text-xl font-semibold text-indigo-600 dark:text-indigo-300">
          Drop PDFs to add to your library
        </p>
      </div>
    </Transition>

    <!-- Hidden file input -->
    <input ref="fileInputRef" type="file" accept="application/pdf" multiple class="hidden" @change="onFileInput" />

    <!-- Header -->
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-zinc-100">Library</h1>
        <p class="text-sm text-slate-500 dark:text-zinc-400 mt-1">
          Every PDF across your classes. Drop a PDF anywhere to add it.
        </p>
      </div>
      <button
        type="button"
        class="flex items-center gap-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors"
        @click="fileInputRef?.click()"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
        </svg>
        Upload PDF
      </button>
    </div>

    <div v-if="loading" class="text-slate-400 text-sm">Loading shelf…</div>

    <!-- Empty state — double as a drop zone -->
    <div
      v-else-if="allItems.length === 0"
      class="flex flex-col items-center justify-center py-20 cursor-pointer group"
      @click="fileInputRef?.click()"
    >
      <div class="w-full max-w-sm border-2 border-dashed border-slate-300 dark:border-zinc-700 group-hover:border-indigo-400 dark:group-hover:border-indigo-500 rounded-2xl p-12 text-center transition-colors">
        <p class="text-5xl mb-4">📚</p>
        <p class="text-base font-medium text-slate-600 dark:text-zinc-300 group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">
          Drop PDFs here or click to browse
        </p>
        <p class="text-sm text-slate-400 dark:text-zinc-500 mt-1">No class or semester required</p>
      </div>
    </div>

    <!-- Grid -->
    <div v-else class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-5">
      <!-- Ghost cards for in-progress uploads -->
      <div
        v-for="p in pending"
        :key="p.uid"
        class="bg-white dark:bg-[#1e1e1e] rounded-xl overflow-hidden border-2 border-dashed"
        :class="p.error ? 'border-red-300 dark:border-red-700' : 'border-slate-300 dark:border-zinc-700 animate-pulse'"
      >
        <div class="aspect-[3/4] bg-slate-100 dark:bg-[#141414] flex items-center justify-center">
          <svg v-if="!p.error" class="w-8 h-8 text-slate-300 dark:text-zinc-600 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          <svg v-else class="w-8 h-8 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9.303 3.376c.866 1.5-.217 3.374-1.948 3.374H2.645c-1.73 0-2.813-1.874-1.948-3.374L10.052 3.378c.866-1.5 3.032-1.5 3.898 0l7.353 12.748zM12 15.75h.007v.008H12v-.008z" />
          </svg>
        </div>
        <div class="p-3">
          <p class="text-sm font-medium text-slate-500 dark:text-zinc-400 truncate">{{ p.filename }}</p>
          <p v-if="p.error" class="text-xs text-red-500 mt-1">{{ p.error }}</p>
          <p v-else class="text-xs text-slate-400 mt-1">Uploading…</p>
        </div>
      </div>

      <!-- Real PDF cards -->
      <div
        v-for="item in allItems"
        :key="item.id"
        class="group bg-white dark:bg-[#1e1e1e] rounded-xl overflow-hidden border border-slate-200 dark:border-zinc-800 hover:shadow-md hover:border-indigo-300 dark:hover:border-indigo-500 transition-all relative"
      >
        <button type="button" class="block w-full text-left" @click="open(item)">
          <DeepReadCover :url="api.attachments.fileUrl(item.id, 'inline')" />
          <div class="p-3">
            <div class="flex items-center gap-1.5 mb-1">
              <span
                class="w-2 h-2 rounded-full shrink-0"
                :style="{ backgroundColor: item.class_color ?? '#94a3b8' }"
              />
              <span class="text-[11px] text-slate-500 dark:text-zinc-400 truncate">
                {{ item.class_name ?? 'Unassigned' }}
              </span>
            </div>
            <p class="text-sm font-medium text-slate-900 dark:text-zinc-100 line-clamp-2">{{ item.filename }}</p>
            <div v-if="progressOf(item.id)" class="mt-2">
              <div class="h-1 rounded-full bg-slate-200 dark:bg-zinc-800 overflow-hidden">
                <div class="h-full bg-indigo-500" :style="{ width: progressPct(item.id) + '%' }" />
              </div>
              <p class="text-[10px] text-slate-500 dark:text-zinc-400 mt-1">
                Page {{ progressOf(item.id)?.lastPage }} of {{ progressOf(item.id)?.totalPages || '–' }}
              </p>
            </div>
          </div>
        </button>

        <!-- Hover actions -->
        <div class="absolute top-2 right-2 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <button
            type="button"
            class="text-[11px] font-medium px-2 py-1 rounded-full bg-white/90 dark:bg-[#141414]/90 text-indigo-600 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-800 hover:bg-indigo-600 hover:text-white hover:border-indigo-600 transition-colors"
            title="Start focus session"
            @click.stop="open(item)"
          >
            ▶ Focus
          </button>
          <button
            type="button"
            class="p-1.5 rounded-full bg-white/90 dark:bg-[#141414]/90 text-slate-500 dark:text-zinc-400 border border-slate-200 dark:border-zinc-800 hover:bg-slate-100 dark:hover:bg-zinc-900 transition-colors"
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

const allItems = ref<PDFLibraryItem[]>([])
const loading = ref(true)
const progressCache = ref<Map<number, DeepReadProgress>>(new Map())

// Upload
const fileInputRef = ref<HTMLInputElement | null>(null)
const isDragging = ref(false)
const dragCounter = ref(0)

interface Pending { uid: string; filename: string; error?: string }
const pending = ref<Pending[]>([])

onMounted(async () => {
  try {
    allItems.value = await api.library.pdfs()
    rebuildProgressCache()
  } finally {
    loading.value = false
  }
})

watch(() => focusStore.active, (active) => {
  if (!active) rebuildProgressCache()
})

function rebuildProgressCache() {
  const map = new Map<number, DeepReadProgress>()
  for (const it of allItems.value) {
    const p = loadProgress(it.id)
    if (p) map.set(it.id, p)
  }
  progressCache.value = map
}

function progressOf(id: number) { return progressCache.value.get(id) }

function progressPct(id: number) {
  const p = progressCache.value.get(id)
  if (!p?.totalPages) return 0
  return Math.min(100, Math.round((p.lastPage / p.totalPages) * 100))
}

function open(item: PDFLibraryItem) {
  focusStore.requestPreflight(item, '/library')
}

// --- Drag & drop ---

function onDragEnter(e: DragEvent) {
  if (!e.dataTransfer?.types.includes('Files')) return
  dragCounter.value++
  isDragging.value = true
}

function onDragLeave() {
  if (--dragCounter.value <= 0) {
    dragCounter.value = 0
    isDragging.value = false
  }
}

function onDrop(e: DragEvent) {
  dragCounter.value = 0
  isDragging.value = false
  if (e.dataTransfer?.files.length) uploadFiles(Array.from(e.dataTransfer.files))
}

function onFileInput(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files?.length) {
    uploadFiles(Array.from(input.files))
    input.value = ''
  }
}

async function uploadFiles(files: File[]) {
  const pdfs = files.filter(f => f.type === 'application/pdf')

  // Show a brief error ghost for any non-PDF
  for (const f of files.filter(f => f.type !== 'application/pdf')) {
    const uid = crypto.randomUUID()
    pending.value.push({ uid, filename: f.name, error: 'Only PDF files are supported' })
    setTimeout(() => { pending.value = pending.value.filter(p => p.uid !== uid) }, 4000)
  }

  if (!pdfs.length) return

  const batch = pdfs.map(f => ({ uid: crypto.randomUUID(), filename: f.name }))
  pending.value.push(...batch)

  await Promise.all(pdfs.map(async (file, i) => {
    const uid = batch[i].uid
    try {
      await api.attachments.uploadToLibrary(file)
      pending.value = pending.value.filter(p => p.uid !== uid)
      allItems.value = await api.library.pdfs()
      rebuildProgressCache()
    } catch (err) {
      const p = pending.value.find(p => p.uid === uid)
      if (p) p.error = err instanceof Error ? err.message : 'Upload failed'
      setTimeout(() => { pending.value = pending.value.filter(p => p.uid !== uid) }, 4000)
    }
  }))
}
</script>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.15s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
