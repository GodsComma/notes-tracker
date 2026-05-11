<template>
  <div class="flex flex-col h-full">
    <!-- Breadcrumb + source header -->
    <div class="border-b border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 shrink-0">
      <!-- Collapsed strip -->
      <div v-if="headerCollapsed" class="px-8 py-2 flex items-center justify-between gap-4">
        <div class="flex items-center gap-2 min-w-0">
          <span v-if="source" class="shrink-0 text-xs font-medium px-2 py-0.5 rounded-full" :class="typeBadgeClass(source.type)">{{ source.type }}</span>
          <span class="text-sm font-semibold text-slate-800 dark:text-slate-100 truncate">{{ source?.title }}</span>
        </div>
        <button
          class="shrink-0 p-1 rounded text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors"
          title="Expand header"
          @click="headerCollapsed = false"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
          </svg>
        </button>
      </div>

      <!-- Full header -->
      <template v-else>
        <div class="px-8 pt-6 pb-4">
          <nav class="text-xs text-slate-400 mb-3 flex items-center gap-1.5">
            <RouterLink to="/" class="hover:text-slate-700">Home</RouterLink>
            <span>/</span>
            <RouterLink v-if="cls" :to="`/classes/${cls.id}`" class="hover:text-slate-700">{{ cls.name }}</RouterLink>
            <span>/</span>
            <span class="text-slate-600 dark:text-slate-400 truncate max-w-48">{{ source?.title }}</span>
          </nav>

          <div v-if="source" class="flex items-start justify-between gap-4">
            <div class="flex items-center gap-3 flex-1 min-w-0">
              <span class="shrink-0 text-xs font-medium px-2 py-0.5 rounded-full" :class="typeBadgeClass(source.type)">
                {{ source.type }}
              </span>
              <div class="flex-1 min-w-0">
                <h1 class="text-xl font-bold text-slate-900 dark:text-slate-100 truncate">{{ source.title }}</h1>
                <p v-if="source.reference || source.date" class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                  <span v-if="source.date">{{ formatDate(source.date) }}</span>
                  <span v-if="source.date && source.reference"> · </span>
                  <span v-if="source.reference">{{ source.reference }}</span>
                </p>
              </div>
            </div>
            <div class="flex items-center gap-3 shrink-0">
              <span class="text-xs" :class="statusClass">{{ statusText }}</span>

              <!-- Lock toggle -->
              <button
                class="flex items-center gap-1.5 px-2 py-1 text-xs font-medium rounded transition-colors"
                :class="editorStore.locked
                  ? 'bg-indigo-100 text-indigo-700 hover:bg-indigo-200 dark:bg-indigo-900/40 dark:text-indigo-300'
                  : 'text-slate-400 hover:text-slate-700 dark:hover:text-slate-200'"
                :title="editorStore.locked ? 'Unlock note (⌘L)' : 'Lock note (⌘L)'"
                @click="toggleLock"
              >
                <svg v-if="editorStore.locked" class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
                <svg v-else class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
                </svg>
                <span v-if="editorStore.locked">Locked</span>
              </button>

              <!-- PDF picker (only when 2+ PDFs) -->
              <select
                v-if="pdfAttachments.length > 1"
                v-model.number="activePdfId"
                class="text-xs border border-slate-200 dark:border-slate-600 rounded px-2 py-1 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-300"
              >
                <option v-for="p in pdfAttachments" :key="p.id" :value="p.id">{{ p.filename }}</option>
              </select>

              <!-- Focus Mode (PDF only, has source) -->
              <template v-if="activePdf">
                <button
                  class="flex items-center gap-1.5 px-2.5 py-1 text-xs font-medium rounded-full bg-indigo-600 text-white hover:bg-indigo-700 transition-colors"
                  title="Start focus session"
                  @click="startFocusOnActivePdf"
                >
                  <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M4 8V4m0 0h4M4 4l5 5m11-5h-4m4 0v4m0-4l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
                  </svg>
                  Focus Mode
                </button>
                <button
                  class="p-1 rounded text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 transition-colors"
                  title="Focus session settings"
                  @click="openFocusSettings"
                >
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                    <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                </button>
              </template>

              <!-- Attachments toggle -->
              <button
                class="p-1 rounded transition-colors"
                :class="showAttachments ? 'text-indigo-600 bg-indigo-50' : 'text-slate-400 hover:text-slate-700'"
                title="Attachments"
                @click="showAttachments = !showAttachments"
              >
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
                </svg>
              </button>
              <button class="text-slate-400 hover:text-slate-700" title="Edit source" @click="openEdit">
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15.232 5.232l3.536 3.536M9 11l6.536-6.536a2 2 0 112.828 2.828L11.828 13.828a2 2 0 01-.828.497l-3 .75.75-3a2 2 0 01.497-.828z" />
                </svg>
              </button>

              <!-- Collapse header -->
              <button
                class="p-1 rounded text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors"
                title="Collapse header"
                @click="headerCollapsed = true"
              >
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- Body: PDF split-pane (when PDF) + editor + optional attachments side panel -->
    <div class="flex-1 flex overflow-hidden">
      <!-- PDF pane -->
      <div
        v-if="activePdf"
        class="shrink-0 overflow-hidden border-r border-slate-200 dark:border-slate-700"
        :style="{ width: pdfPaneWidth + 'px' }"
      >
        <PdfViewer
          ref="pdfViewer"
          :url="api.attachments.fileUrl(activePdf.id, 'inline')"
          :anchors="anchorsForActivePdf"
          @select="onPdfSelect"
          @anchor-click="onPdfAnchorClick"
        />
      </div>

      <!-- Drag divider -->
      <div
        v-if="activePdf"
        class="w-1.5 cursor-col-resize bg-slate-200 dark:bg-slate-700 hover:bg-indigo-400 dark:hover:bg-indigo-500 transition-colors shrink-0"
        @mousedown="startDrag"
      />

      <!-- Editor pane -->
      <div class="flex-1 overflow-y-auto px-8 py-6 min-w-0">
        <div :class="activePdf ? 'max-w-3xl mx-auto' : 'max-w-4xl mx-auto'">
          <TipTapEditor
            v-if="noteLoaded"
            ref="editorRef"
            :initial-content="initialBody"
            :locked="editorStore.locked"
            @update="onNoteUpdate"
            @anchor-click="onEditorAnchorClick"
          />
        </div>
      </div>

      <!-- Attachments side panel -->
      <aside
        v-if="showAttachments"
        class="w-80 shrink-0 border-l border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 overflow-y-auto p-4"
      >
        <p class="text-xs font-semibold uppercase tracking-wide text-slate-400 dark:text-slate-500 mb-3">Attachments</p>
        <AttachmentPanel
          v-if="noteLoaded"
          :scope="{ type: 'source', id: sourceID }"
          :focus-context="sourceFocusContext"
          :return-to="$route.path"
          @list-change="onAttachmentsChange"
        />
      </aside>
    </div>

    <!-- Edit Source Modal -->
    <BaseModal :open="showEdit" title="Edit Source" @close="showEdit = false">
      <form class="space-y-4" @submit.prevent="saveSource">
        <div>
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Type</label>
          <select v-model="editForm.type" class="input">
            <option v-for="t in sourceTypes" :key="t" :value="t">{{ t }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Title *</label>
          <input v-model="editForm.title" required class="input" />
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Date</label>
            <input v-model="editForm.date" type="date" class="input" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Reference</label>
            <input v-model="editForm.reference" class="input" />
          </div>
        </div>
        <div class="flex justify-end gap-2 pt-2">
          <button type="button" class="btn-ghost" @click="showEdit = false">Cancel</button>
          <button type="submit" class="btn-primary">Save</button>
        </div>
      </form>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { api, type Source, type Class, type Attachment } from '@/api'
import type { PdfAnchor, PdfRect } from '@/components/PdfViewer.vue'
import type { PdfQuoteColor } from '@/extensions/PdfQuote'
import { preloadPdf } from '@/composables/usePdfCache'
import { useEditorStore } from '@/stores/editor'
import { useClassStore } from '@/stores/class'
import { useFocusStore } from '@/stores/focus'
import TipTapEditor from '@/components/TipTapEditor.vue'
import PdfViewer from '@/components/PdfViewer.vue'
import BaseModal from '@/components/BaseModal.vue'
import AttachmentPanel from '@/components/AttachmentPanel.vue'

const route = useRoute()
const editorStore = useEditorStore()
const classStore = useClassStore()
const focusStore = useFocusStore()
const sourceID = Number(route.params.id)

const source = ref<Source | null>(null)
const cls = ref<Class | null>(null)
const initialBody = ref('')
const noteLoaded = ref(false)
const showEdit = ref(false)
const showAttachments = ref(false)
const headerCollapsed = ref(false)
const sourceTypes = ['Lecture', 'Reading', 'Problem Set', 'Discussion', 'Lab', 'Other']
const editForm = ref({ type: 'Lecture', title: '', date: '', reference: '' })

const attachments = ref<Attachment[]>([])
const activePdfId = ref<number | null>(null)

const editorRef = ref<InstanceType<typeof TipTapEditor> | null>(null)
const pdfViewer = ref<InstanceType<typeof PdfViewer> | null>(null)

const currentNoteJSON = ref<unknown>(null)

const pdfAttachments = computed(() =>
  attachments.value.filter(a => a.mime_type === 'application/pdf')
)

const activePdf = computed(() =>
  pdfAttachments.value.find(p => p.id === activePdfId.value) ?? null
)

const allAnchors = computed<PdfAnchor[]>(() => {
  return collectAnchors(currentNoteJSON.value)
})

const anchorsForActivePdf = computed<PdfAnchor[]>(() => {
  const id = activePdfId.value
  if (id === null) return []
  return allAnchors.value.filter(a => (a as PdfAnchor & { attachmentID: number }).attachmentID === id)
})

const sourceFocusContext = computed(() => {
  if (!cls.value || !source.value) return undefined
  return {
    class_id: cls.value.id,
    class_name: cls.value.name,
    class_color: cls.value.color,
    source_id: source.value.id,
    source_title: source.value.title
  }
})

// Fix 7: warm the PDF cache as soon as the active PDF is known so that
// PdfViewer can skip the getDocument() download step on entry to focus mode.
watch(activePdf, (pdf) => {
  if (pdf) preloadPdf(api.attachments.fileUrl(pdf.id, 'inline'))
})

function focusItem() {
  if (!activePdf.value || !sourceFocusContext.value) return null
  return {
    id: activePdf.value.id,
    filename: activePdf.value.filename,
    size_bytes: activePdf.value.size_bytes,
    created_at: activePdf.value.created_at,
    ...sourceFocusContext.value
  }
}

async function startFocusOnActivePdf() {
  const item = focusItem()
  if (!item) return
  editorStore.flushSave()
  focusStore.requestPreflight(item, route.path)
}

async function openFocusSettings() {
  const item = focusItem()
  if (!item) return
  focusStore.openSettings(item, route.path)
}

// Walk TipTap JSON document to extract pdfQuote nodes
function collectAnchors(doc: unknown): (PdfAnchor & { attachmentID: number })[] {
  const out: (PdfAnchor & { attachmentID: number })[] = []
  if (!doc || typeof doc !== 'object') return out
  const stack: unknown[] = [doc]
  while (stack.length) {
    const n = stack.pop()
    if (!n || typeof n !== 'object') continue
    const node = n as { type?: string; attrs?: Record<string, unknown>; content?: unknown[] }
    if (node.type === 'pdfQuote' && node.attrs) {
      out.push({
        anchorId: String(node.attrs.anchorId ?? ''),
        page: Number(node.attrs.page ?? 1),
        text: String(node.attrs.text ?? ''),
        rects: (node.attrs.rects as PdfRect[] | undefined) ?? [],
        color: (node.attrs.color as PdfQuoteColor | undefined) ?? 'yellow',
        attachmentID: Number(node.attrs.attachmentID ?? 0)
      })
    }
    if (Array.isArray(node.content)) {
      for (const child of node.content) stack.push(child)
    }
  }
  return out
}

function generateAnchorId() {
  return 'a_' + Math.random().toString(36).slice(2, 10) + Date.now().toString(36).slice(-4)
}

onMounted(async () => {
  editorStore.setSourceID(sourceID)
  const [src, note] = await Promise.all([
    api.sources.get(sourceID),
    api.notes.get(sourceID)
  ])
  source.value = src
  cls.value = classStore.classes.find(c => c.id === src.class_id) ?? null
  initialBody.value = note.body ?? ''
  editorStore.body = initialBody.value
  editorStore.setLockedFromServer(!!note.locked)
  currentNoteJSON.value = initialBody.value ? JSON.parse(initialBody.value) : null
  noteLoaded.value = true
  window.addEventListener('keydown', onKey)
  window.addEventListener('app:focus-session-ended', onFocusSessionEnded)

  attachments.value = await api.attachments.listForSource(sourceID)
  showAttachments.value = attachments.value.length > 0 && pdfAttachments.value.length === 0
  if (pdfAttachments.value.length > 0) activePdfId.value = pdfAttachments.value[0].id
})

onBeforeUnmount(() => {
  editorStore.flushSave()
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
  window.removeEventListener('keydown', onKey)
  window.removeEventListener('app:focus-session-ended', onFocusSessionEnded)
})

async function onFocusSessionEnded(e: Event) {
  if ((e as CustomEvent).detail?.sourceId !== sourceID) return
  const note = await api.notes.get(sourceID)
  const newBody = note.body ?? ''
  editorStore.setLockedFromServer(!!note.locked)
  // TipTapEditor watches initialContent and calls setContent when it changes
  initialBody.value = newBody
  await nextTick()
  editorStore.body = newBody
  currentNoteJSON.value = newBody ? JSON.parse(newBody) : null
}

async function toggleLock() {
  await editorStore.setLock(!editorStore.locked)
}

function onKey(e: KeyboardEvent) {
  // Suppress when typing in any input/textarea, including the editor itself.
  const t = e.target as HTMLElement | null
  if (t && (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.isContentEditable)) return
  if ((e.metaKey || e.ctrlKey) && (e.key === 'l' || e.key === 'L')) {
    e.preventDefault()
    toggleLock()
  }
}

function onNoteUpdate(content: string) {
  editorStore.onChange(content)
  try { currentNoteJSON.value = content ? JSON.parse(content) : null } catch { /* ignore */ }
}

async function onAttachmentsChange() {
  attachments.value = await api.attachments.listForSource(sourceID)
  if (activePdfId.value && !pdfAttachments.value.find(p => p.id === activePdfId.value)) {
    activePdfId.value = pdfAttachments.value[0]?.id ?? null
  }
  if (!activePdfId.value && pdfAttachments.value.length > 0) {
    activePdfId.value = pdfAttachments.value[0].id
  }
}

function onPdfSelect(payload: { page: number; text: string; rects: PdfRect[]; color: PdfQuoteColor }) {
  if (!activePdfId.value || !editorRef.value) return
  if (editorStore.locked) return // lock gate — no new captures
  editorRef.value.insertPdfQuote({
    attachmentID: activePdfId.value,
    page: payload.page,
    text: payload.text,
    anchorId: generateAnchorId(),
    rects: payload.rects,
    color: payload.color
  })
}

function onPdfAnchorClick(anchorId: string) {
  editorRef.value?.scrollToAnchor(anchorId)
}

function onEditorAnchorClick(anchorId: string) {
  // Switch to the right PDF if needed
  const anchor = allAnchors.value.find(a => a.anchorId === anchorId) as
    (PdfAnchor & { attachmentID: number }) | undefined
  if (!anchor) return
  if (activePdfId.value !== anchor.attachmentID) {
    activePdfId.value = anchor.attachmentID
    // Wait a tick for the PdfViewer to load the new doc
    window.setTimeout(() => pdfViewer.value?.scrollToAnchor(anchorId), 400)
  } else {
    pdfViewer.value?.scrollToAnchor(anchorId)
  }
}

// --- Resizable divider ---

const pdfPaneWidth = ref(640)
let dragStartX = 0
let dragStartW = 0

function startDrag(e: MouseEvent) {
  dragStartX = e.clientX
  dragStartW = pdfPaneWidth.value
  document.addEventListener('mousemove', onDrag)
  document.addEventListener('mouseup', stopDrag)
  e.preventDefault()
}

function onDrag(e: MouseEvent) {
  const delta = e.clientX - dragStartX
  pdfPaneWidth.value = Math.max(320, Math.min(window.innerWidth - 400, dragStartW + delta))
}

function stopDrag() {
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
}

const statusText = computed(() => {
  switch (editorStore.saveStatus) {
    case 'saving': return 'Saving…'
    case 'saved': return 'Saved'
    case 'error': return 'Save failed – retrying'
    case 'locked': return 'Locked'
    default: return ''
  }
})

const statusClass = computed(() => {
  switch (editorStore.saveStatus) {
    case 'saving': return 'text-slate-400'
    case 'saved': return 'text-green-500'
    case 'error': return 'text-red-400'
    case 'locked': return 'text-indigo-500'
    default: return 'text-transparent'
  }
})

function typeBadgeClass(type: string) {
  const map: Record<string, string> = {
    Lecture: 'bg-blue-100 text-blue-700',
    Reading: 'bg-purple-100 text-purple-700',
    'Problem Set': 'bg-orange-100 text-orange-700',
    Discussion: 'bg-green-100 text-green-700',
    Lab: 'bg-teal-100 text-teal-700',
    Other: 'bg-slate-100 text-slate-600'
  }
  return map[type] ?? 'bg-slate-100 text-slate-600'
}

function formatDate(d: string) {
  if (!d) return ''
  return new Date(d + 'T00:00:00').toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

function openEdit() {
  if (!source.value) return
  editForm.value = {
    type: source.value.type,
    title: source.value.title,
    date: source.value.date,
    reference: source.value.reference
  }
  showEdit.value = true
}

async function saveSource() {
  if (!source.value) return
  const updated = await api.sources.update(source.value.id, editForm.value)
  source.value = updated
  showEdit.value = false
}
</script>
