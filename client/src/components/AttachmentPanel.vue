<template>
  <div
    class="rounded-xl border-2 transition-colors"
    :class="isDragging ? 'border-indigo-400 bg-indigo-50 dark:bg-indigo-950' : 'border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800'"
    @dragover.prevent="isDragging = true"
    @dragleave.prevent="isDragging = false"
    @drop.prevent="onDrop"
  >
    <!-- Header -->
    <div class="flex items-center justify-between px-4 py-3 border-b border-slate-100 dark:border-slate-700">
      <span class="text-sm font-semibold text-slate-700 dark:text-slate-300">
        Files
        <span v-if="attachments.length" class="ml-1 text-xs font-normal text-slate-400">({{ attachments.length }})</span>
      </span>
      <button
        class="flex items-center gap-1.5 text-xs text-indigo-600 hover:text-indigo-800 font-medium"
        @click="fileInput?.click()"
      >
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
        </svg>
        Upload
      </button>
      <input ref="fileInput" type="file" multiple class="hidden" @change="onFileInput" />
    </div>

    <!-- File list -->
    <div v-if="attachments.length || uploading.length" class="divide-y divide-slate-50">
      <!-- Uploading items -->
      <div
        v-for="u in uploading"
        :key="u.id"
        class="flex items-center gap-3 px-4 py-2.5 text-sm text-slate-500 dark:text-slate-400"
      >
        <span class="animate-spin text-base">⏳</span>
        <span class="flex-1 truncate">{{ u.name }}</span>
        <span class="text-xs text-indigo-500">Uploading…</span>
      </div>

      <!-- Existing attachments -->
      <div
        v-for="a in attachments"
        :key="a.id"
        class="flex items-center gap-3 px-4 py-2.5 group divide-y divide-slate-50 dark:divide-slate-700"
      >
        <span
          class="text-base shrink-0"
          :class="isPreviewable(a.mime_type) ? 'cursor-pointer' : ''"
          @click="isPreviewable(a.mime_type) ? (previewing = a) : null"
        >{{ fileIcon(a.mime_type) }}</span>

        <span
          class="flex-1 text-sm text-slate-700 dark:text-slate-300 truncate min-w-0"
          :class="isPreviewable(a.mime_type) ? 'cursor-pointer hover:text-indigo-600' : ''"
          @click="isPreviewable(a.mime_type) ? (previewing = a) : null"
        >{{ a.filename }}</span>

        <span class="text-xs text-slate-400 shrink-0">{{ formatBytes(a.size_bytes) }}</span>

        <!-- Focus Read (PDFs only, when focusContext provided) -->
        <button
          v-if="focusContext && a.mime_type === 'application/pdf'"
          class="shrink-0 text-[11px] font-medium px-2 py-0.5 rounded-full text-indigo-600 dark:text-indigo-300 bg-indigo-50 dark:bg-indigo-950/40 hover:bg-indigo-600 hover:text-white opacity-0 group-hover:opacity-100 transition-all"
          title="Start focus session"
          @click.stop="startFocus(a)"
        >
          ▶ Focus
        </button>

        <!-- Download -->
        <a
          :href="api.attachments.fileUrl(a.id, 'attachment')"
          download
          class="shrink-0 text-slate-300 hover:text-indigo-500 opacity-0 group-hover:opacity-100 transition-opacity"
          title="Download"
          @click.stop
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
          </svg>
        </a>

        <!-- Delete -->
        <button
          class="shrink-0 text-slate-300 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity"
          title="Delete"
          @click="remove(a)"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Empty drop zone -->
    <div
      v-else
      class="px-4 py-8 text-center text-slate-400 dark:text-slate-500 text-sm"
      @click="fileInput?.click()"
    >
      <p class="text-2xl mb-2">📎</p>
      <p>Drop files here or <span class="text-indigo-500 cursor-pointer hover:underline">click to upload</span></p>
    </div>
  </div>

  <!-- Preview modal (self-contained) -->
  <FilePreviewModal :attachment="previewing" @close="previewing = null" />
  <ConfirmModal
    :open="!!pendingDelete"
    title="Delete file"
    :message="pendingDelete ? `Delete &quot;${pendingDelete.filename}&quot;? This cannot be undone.` : ''"
    @confirm="confirmRemove"
    @cancel="pendingDelete = null"
  />
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type Attachment } from '@/api'
import { fileIcon, isPreviewable, formatBytes } from '@/utils/mime'
import { useFocusStore } from '@/stores/focus'
import FilePreviewModal from './FilePreviewModal.vue'
import ConfirmModal from './ConfirmModal.vue'

const props = defineProps<{
  scope: { type: 'class'; id: number } | { type: 'source'; id: number }
  focusContext?: {
    class_id: number
    class_name: string
    class_color: string
    source_id: number | null
    source_title: string
  }
  returnTo?: string
}>()

const emit = defineEmits<{ 'list-change': [] }>()

const focusStore = useFocusStore()

function startFocus(att: Attachment) {
  if (!props.focusContext) return
  focusStore.requestPreflight(
    {
      id: att.id,
      filename: att.filename,
      size_bytes: att.size_bytes,
      created_at: att.created_at,
      class_id: props.focusContext.class_id,
      class_name: props.focusContext.class_name,
      class_color: props.focusContext.class_color,
      source_id: props.focusContext.source_id,
      source_title: props.focusContext.source_title
    },
    props.returnTo
  )
}

const attachments = ref<Attachment[]>([])
const uploading = ref<{ id: number; name: string }[]>([])
const isDragging = ref(false)
let nextUploadId = 0
const previewing = ref<Attachment | null>(null)
const pendingDelete = ref<Attachment | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)

onMounted(async () => {
  attachments.value = await fetchList()
})

function fetchList(): Promise<Attachment[]> {
  return props.scope.type === 'class'
    ? api.attachments.listForClass(props.scope.id)
    : api.attachments.listForSource(props.scope.id)
}

function uploadFile(file: File): Promise<Attachment> {
  return props.scope.type === 'class'
    ? api.attachments.uploadToClass(props.scope.id, file)
    : api.attachments.uploadToSource(props.scope.id, file)
}

function onFileInput(e: Event) {
  const files = (e.target as HTMLInputElement).files
  if (files) processFiles(Array.from(files))
  // Reset input so the same file can be re-uploaded if needed
  ;(e.target as HTMLInputElement).value = ''
}

function onDrop(e: DragEvent) {
  isDragging.value = false
  const files = e.dataTransfer?.files
  if (files?.length) processFiles(Array.from(files))
}

async function processFiles(files: File[]) {
  for (const file of files) {
    const id = nextUploadId++
    uploading.value.push({ id, name: file.name })
    try {
      const att = await uploadFile(file)
      attachments.value.push(att)
      emit('list-change')
    } catch {
      alert(`Failed to upload "${file.name}". Please try again.`)
    } finally {
      uploading.value = uploading.value.filter(u => u.id !== id)
    }
  }
}

function remove(a: Attachment) {
  pendingDelete.value = a
}

async function confirmRemove() {
  const a = pendingDelete.value
  pendingDelete.value = null
  if (!a) return
  await api.attachments.delete(a.id)
  attachments.value = attachments.value.filter(x => x.id !== a.id)
  emit('list-change')
}
</script>
