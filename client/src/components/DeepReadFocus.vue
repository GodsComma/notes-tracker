<template>
  <div
    ref="root"
    class="dr-focus"
    :data-theme="settings.theme"
    :data-ruler="settings.ruler"
    @mousemove="onMouseMove"
  >
    <!-- Drain timer bar -->
    <div class="dr-drain">
      <div class="dr-drain-fill" :style="{ width: drainPct + '%' }" />
    </div>

    <!-- Time readout -->
    <div class="dr-time">{{ remainingLabel }}</div>

    <!-- PDF (delegated to PdfViewer for selection + highlights) -->
    <PdfViewer
      ref="pdfRef"
      class="dr-pdf"
      :url="url"
      :anchors="anchorsForThisPdf"
      :scale="zoom.scale"
      :enable-selection="annotationsLive"
      @select="onSelect"
      @anchor-migrated="onAnchorMigrated"
      @anchor-click="onAnchorClick"
      @page-change="onPageChange"
      @loaded="onPdfLoaded"
    />

    <!-- Reading ruler -->
    <div
      v-if="settings.ruler === 'underline'"
      class="dr-ruler-underline"
      :style="{ top: cursorY + 'px' }"
    />
    <template v-else-if="settings.ruler === 'masking'">
      <div class="dr-ruler-mask dr-ruler-mask-top" :style="{ height: Math.max(0, cursorY - 60) + 'px' }" />
      <div class="dr-ruler-mask dr-ruler-mask-bottom" :style="{ top: (cursorY + 60) + 'px' }" />
    </template>

    <!-- Bottom control strip -->
    <div class="dr-controls">
      <!-- Paging -->
      <button class="dr-ctl-sm" title="Previous page (PgUp / ←)" :disabled="currentPage <= 1" @click="goPrev">‹</button>
      <input
        ref="pageInputEl"
        class="dr-page-input"
        type="text"
        inputmode="numeric"
        :value="pageInputValue"
        @focus="pageInputFocused = true"
        @blur="commitPageInput"
        @keydown.enter.prevent="commitPageInput($event)"
        @keydown.esc.prevent="cancelPageInput"
        @input="onPageInputType(($event.target as HTMLInputElement).value)"
      />
      <span class="dr-page-total">/ {{ totalPages || '–' }}</span>
      <button class="dr-ctl-sm" title="Next page (PgDn / →)" :disabled="totalPages > 0 && currentPage >= totalPages" @click="goNext">›</button>

      <span class="dr-spacer" />

      <!-- Zoom -->
      <button class="dr-ctl-sm" title="Zoom out" @click="zoomOut">−</button>
      <span class="dr-zoom-pct">{{ zoomPct }}%</span>
      <button class="dr-ctl-sm" title="Zoom in" @click="zoomIn">+</button>

      <select
        class="dr-fit-select"
        :value="zoom.mode"
        title="Fit mode"
        @change="onFitModeChange(($event.target as HTMLSelectElement).value)"
      >
        <option value="absolute">—</option>
        <option value="fit-width">Fit width</option>
        <option value="fit-page">Fit page</option>
      </select>

      <span class="dr-spacer" />

      <button
        v-if="annotationsLive"
        class="dr-notes-pill"
        :class="{ 'dr-notes-pill-active': drawerOpen }"
        :title="drawerOpen ? 'Close notes (N)' : 'Open notes (N)'"
        @click="toggleDrawer"
      >
        📝 Notes ({{ anchorsForThisPdf.length }})
      </button>
      <span v-if="annotationsLive" class="dr-spacer" />
      <button class="dr-exit" title="Exit session" @click="requestExit">Exit</button>
    </div>

    <!-- Toast -->
    <transition name="dr-toast">
      <div v-if="toast" class="dr-toast">{{ toast }}</div>
    </transition>

    <!-- Notes drawer -->
    <transition name="dr-drawer">
      <aside
        v-if="drawerOpen && annotationsLive"
        class="dr-drawer"
        @mousemove.stop="bumpDrawerActivity"
        @click="bumpDrawerActivity"
        @keydown="bumpDrawerActivity"
      >
        <header class="dr-drawer-head">
          <span class="dr-drawer-title">Notes</span>
          <button class="dr-drawer-close" title="Close (N)" @click="toggleDrawer">×</button>
        </header>

        <div v-if="anchorsForThisPdf.length === 0 && !textNotes" class="dr-drawer-empty">
          Select PDF text to capture your first note.
        </div>
        <div v-else-if="anchorsForThisPdf.length === 0" class="dr-drawer-empty">
          Select PDF text to capture a highlight.
        </div>

        <div v-if="anchorsForThisPdf.length > 0" class="dr-drawer-list">
          <div
            v-for="a in anchorsForThisPdf"
            :key="a.anchorId"
            class="dr-row"
            :class="{ 'dr-row-expanded': expandedId === a.anchorId }"
            :data-color="a.color"
            @click="onRowClick(a.anchorId)"
          >
            <div class="dr-row-head">
              <span class="dr-row-dot" :data-color="a.color" />
              <span class="dr-row-snippet">{{ a.text }}</span>
              <span class="dr-row-page">p.{{ a.page }}</span>
            </div>

            <div v-if="expandedId === a.anchorId" class="dr-row-body" @click.stop>
              <textarea
                class="dr-row-edit"
                rows="3"
                placeholder="Add your note…"
                :value="getCommentary(a.anchorId)"
                @input="onCommentaryInput(a.anchorId, ($event.target as HTMLTextAreaElement).value)"
              />
              <div class="dr-row-actions">
                <div class="dr-row-colors">
                  <button
                    v-for="c in PDF_QUOTE_COLORS"
                    :key="c"
                    type="button"
                    class="dr-row-color"
                    :data-color="c"
                    :class="{ 'dr-row-color-active': a.color === c }"
                    :title="`Recolor (${c})`"
                    @click.stop="recolor(a.anchorId, c)"
                  />
                </div>
                <button class="dr-row-delete" title="Delete note" @click.stop="removeAnchor(a.anchorId)">
                  ×
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Text notes written in the regular editor -->
        <div v-if="textNotes" class="dr-text-notes">
          <p class="dr-text-notes-label">Text notes</p>
          <p class="dr-text-notes-body">{{ textNotes }}</p>
        </div>
      </aside>
    </transition>

    <!-- Exit guard -->
    <div v-if="showExitGuard" class="dr-guard-backdrop">
      <div class="dr-guard">
        <p class="text-base font-semibold mb-2">Pause focus session?</p>
        <p class="text-sm opacity-80 mb-2">
          You still have <strong>{{ remainingLabel }}</strong> left. Resume to keep going, or exit to leave the session.
        </p>
        <p class="text-xs opacity-70 mb-4">{{ exitSaveLabel }}</p>
        <div class="flex justify-end gap-2">
          <button class="dr-btn-ghost" :disabled="exiting" @click="confirmExit">
            {{ exiting ? 'Saving…' : 'Exit session' }}
          </button>
          <button class="dr-btn-primary" @click="resumeSession">Resume</button>
        </div>
      </div>
    </div>

    <!-- Splash loader: covers everything until PDF + notes + page restore are all ready. -->
    <transition name="dr-splash">
      <div v-if="loading" class="dr-splash" @wheel.prevent @touchmove.prevent>
        <div class="dr-splash-spinner" />
        <p class="dr-splash-label">Preparing your reading session…</p>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import {
  saveProgress, loadProgress, effectiveZoomMode, todayISO,
  type DeepReadSettings, type ZoomMode
} from '@/composables/useDeepRead'
import PdfViewer, { type PdfAnchor, type PdfAnchorMigration } from './PdfViewer.vue'
import { PDF_QUOTE_COLORS, type PdfQuoteColor } from '@/extensions/PdfQuote'
import { api } from '@/api'

const props = defineProps<{
  url: string
  attachmentId: number
  sourceId: number | null
  settings: DeepReadSettings
  filename: string
}>()

const emit = defineEmits<{ exit: []; complete: [] }>()

const root = ref<HTMLElement | null>(null)
const pdfRef = ref<InstanceType<typeof PdfViewer> | null>(null)

// Splash readiness — the overlay stays up until ALL three are true. This
// covers the multi-second gap between mount and "scrolled to last page".
const noteReady = ref(false)
const pdfReady = ref(false)
const restoreReady = ref(false)
const loading = computed(() => !noteReady.value || !pdfReady.value || !restoreReady.value)
const totalPages = ref(0)
const currentPage = ref(1)

// Load saved zoom synchronously in setup so PdfViewer mounts with the correct scale.
const initialSaved = loadProgress(props.attachmentId)
const zoom = ref<{ mode: ZoomMode; scale: number }>({
  mode: effectiveZoomMode(initialSaved),
  scale: initialSaved?.zoom && initialSaved.zoom > 0 ? initialSaved.zoom : 1
})

// Fetch last page from the server in parallel with PDF loading. Server is the
// source of truth — localStorage was unreliable across tabs/private mode/etc.
const lastPagePromise: Promise<number> = props.attachmentId
  ? api.attachments.getProgress(props.attachmentId)
      .then(r => r.last_page || 1)
      .catch(() => initialSaved?.lastPage ?? 1)  // fall back to localStorage if API fails
  : Promise.resolve(1)
const cursorY = ref(0)
const showExitGuard = ref(false)
const exiting = ref(false)

const zoomPct = computed(() => Math.round(zoom.value.scale * 100))

const pageInputEl = ref<HTMLInputElement | null>(null)
const pageInputFocused = ref(false)
const pageInputDraft = ref('')
const pageInputValue = computed(() =>
  pageInputFocused.value ? pageInputDraft.value : String(currentPage.value)
)

let resizeObserver: ResizeObserver | null = null
let resizeTimer: number | null = null
const SCALE_MIN = 0.5
const SCALE_MAX = 3.0

const drawerOpen = ref(false)
const expandedId = ref<string | null>(null)
const toast = ref('')
const sessionCaptureCount = ref(0)
let toastTimer: number | null = null
let drawerHideTimer: number | null = null

// --- Note document state (TipTap JSON) ---
type DocNode = { type: string; attrs?: Record<string, unknown>; content?: DocNode[] }
const noteDoc = ref<DocNode | null>(null)
const noteLocked = ref(false)
const saveStatus = ref<'idle' | 'saving' | 'saved' | 'error' | 'locked'>('idle')
let saveDebounce: number | null = null
let lastSavedJSON = ''

let endTime = 0
let sessionStartTime = 0
let tickInterval: number | null = null
let progressInterval: number | null = null
let pendingExit = false

const annotationsLive = computed(() =>
  props.settings.annotations === 'on' && props.sourceId !== null
)

const totalMs = computed(() => props.settings.durationMin * 60_000)
const remainingMs = ref(props.settings.durationMin * 60_000)
const drainPct = computed(() => {
  if (totalMs.value <= 0) return 0
  return Math.max(0, Math.min(100, (remainingMs.value / totalMs.value) * 100))
})
const remainingLabel = computed(() => {
  const ms = Math.max(0, remainingMs.value)
  const total = Math.ceil(ms / 1000)
  const m = Math.floor(total / 60)
  const s = total % 60
  return `${m}:${String(s).padStart(2, '0')} remaining`
})
const exitSaveLabel = computed(() => {
  if (!annotationsLive.value) return ''
  if (saveStatus.value === 'saving') return 'Saving 1 unsaved note…'
  return 'All notes saved.'
})

const anchorsForThisPdf = computed<PdfAnchor[]>(() => {
  return collectAnchors(noteDoc.value).filter(a => a.attachmentID === props.attachmentId)
})

onMounted(async () => {
  // Fix 3: kick off note fetch immediately — don't block it behind the
  // fullscreen animation (100–300ms wasted). By the time we await below
  // the response is usually already in flight or resolved.
  const notePromise = props.sourceId !== null
    ? api.notes.get(props.sourceId).catch(() => null)
    : Promise.resolve(null)

  await enterFullscreen()
  document.addEventListener('fullscreenchange', onFullscreenChange)
  window.addEventListener('keydown', onKey)
  window.addEventListener('beforeunload', persistProgress)

  sessionStartTime = Date.now()
  endTime = sessionStartTime + totalMs.value
  tickInterval = window.setInterval(tick, 250)
  progressInterval = window.setInterval(persistProgress, 10_000)

  const note = await notePromise
  if (note) {
    noteDoc.value = note.body ? JSON.parse(note.body) : emptyDoc()
    noteLocked.value = !!note.locked
    lastSavedJSON = note.body ?? ''
  } else {
    noteDoc.value = emptyDoc()
  }

  noteReady.value = true
})

onBeforeUnmount(() => {
  if (tickInterval) clearInterval(tickInterval)
  if (progressInterval) clearInterval(progressInterval)
  if (saveDebounce) clearTimeout(saveDebounce)
  if (toastTimer) clearTimeout(toastTimer)
  if (drawerHideTimer) clearTimeout(drawerHideTimer)
  if (resizeTimer) clearTimeout(resizeTimer)
  if (pageChangeSaveTimer) clearTimeout(pageChangeSaveTimer)
  resizeObserver?.disconnect()
  resizeObserver = null
  document.removeEventListener('fullscreenchange', onFullscreenChange)
  window.removeEventListener('keydown', onKey)
  window.removeEventListener('beforeunload', persistProgress)
  persistProgress()
  flushSave()
  if (document.fullscreenElement) {
    document.exitFullscreen().catch(() => { /* ignore */ })
  }
})

// --- Fullscreen ---

async function enterFullscreen() {
  if (!root.value || document.fullscreenElement) return
  try { await root.value.requestFullscreen() } catch { /* still usable without */ }
}
function onFullscreenChange() {
  if (!document.fullscreenElement && !pendingExit) {
    showExitGuard.value = true
  }
}

// --- Timer ---

function tick() {
  remainingMs.value = endTime - Date.now()
  if (remainingMs.value <= 0) {
    remainingMs.value = 0
    finishSession()
  }
}
async function finishSession() {
  if (tickInterval) { clearInterval(tickInterval); tickInterval = null }
  // For open-ended sessions (durationMin === 0), totalMs is 0 so we measure wall-clock elapsed.
  const durationSec = totalMs.value > 0
    ? Math.round(totalMs.value / 1000)
    : Math.round((Date.now() - sessionStartTime) / 1000)
  api.focusSessions.log({ date: todayISO(), duration_sec: durationSec, attachment_id: props.attachmentId, filename: props.filename, completed: true }).catch(() => {})
  playChime()
  persistProgress()
  await flushSave()
  emit('complete')
  pendingExit = true
  if (document.fullscreenElement) {
    try { await document.exitFullscreen() } catch { /* ignore */ }
  }
}
function playChime() {
  try {
    const Ctor = window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext
    const ctx = new Ctor()
    const osc = ctx.createOscillator()
    const gain = ctx.createGain()
    osc.frequency.value = 660
    gain.gain.value = 0.0001
    gain.gain.exponentialRampToValueAtTime(0.2, ctx.currentTime + 0.05)
    gain.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + 1.2)
    osc.connect(gain).connect(ctx.destination)
    osc.start()
    osc.stop(ctx.currentTime + 1.3)
  } catch { /* audio is best-effort */ }
}

// --- Page / Zoom ---

async function onPdfLoaded(info: { numPages: number }) {
  totalPages.value = info.numPages
  pdfReady.value = true

  // Fix 4: run fitMode and last-page fetch concurrently — they're independent.
  const [, lastPage] = await Promise.all([
    zoom.value.mode !== 'absolute' ? applyFitMode() : Promise.resolve(undefined),
    lastPagePromise
  ])

  if (lastPage > 1) {
    pdfRef.value?.setRestorePage(lastPage)
    // Fix 2: wait for the rescale triggered by applyFitMode to fully settle,
    // then scroll once. Replaces the brittle 400ms + 1200ms setTimeout chain.
    await nextTick()  // ensure Vue has flushed the zoom prop update to PdfViewer
    await pdfRef.value?.rescaleDone()
    pdfRef.value?.scrollToPage(lastPage)
  }
  restoreReady.value = true

  // Watch the scroller for size changes (sidebar / drawer / window).
  const scroller = root.value?.querySelector('.dr-pdf') as HTMLElement | null
  if (scroller && !resizeObserver) {
    resizeObserver = new ResizeObserver(() => {
      if (zoom.value.mode === 'absolute') return
      if (resizeTimer) clearTimeout(resizeTimer)
      resizeTimer = window.setTimeout(applyFitMode, 100)
    })
    resizeObserver.observe(scroller)
  }
}

let pageChangeSaveTimer: number | null = null
function onPageChange(p: number) {
  currentPage.value = p
  // Debounced save on every page change. The 1500ms debounce avoids hammering
  // the server while the user is scrolling, but ensures the latest page is
  // always persisted even if the user closes the tab abruptly.
  if (pageChangeSaveTimer) clearTimeout(pageChangeSaveTimer)
  pageChangeSaveTimer = window.setTimeout(() => {
    if (props.attachmentId && currentPage.value >= 1) {
      api.attachments.saveProgress(props.attachmentId, currentPage.value).catch(() => {})
    }
  }, 1500)
}

function clampScale(s: number) {
  return Math.max(SCALE_MIN, Math.min(SCALE_MAX, +s.toFixed(3)))
}

function zoomIn() {
  zoom.value = { mode: 'absolute', scale: clampScale(zoom.value.scale * 1.1) }
}

function zoomOut() {
  zoom.value = { mode: 'absolute', scale: clampScale(zoom.value.scale / 1.1) }
}

function onFitModeChange(mode: string) {
  if (mode === 'fit-width' || mode === 'fit-page') {
    zoom.value = { ...zoom.value, mode }
    applyFitMode()
  } else {
    zoom.value = { ...zoom.value, mode: 'absolute' }
  }
}

async function applyFitMode() {
  const scroller = root.value?.querySelector('.dr-pdf') as HTMLElement | null
  if (!scroller || !pdfRef.value) return
  if (zoom.value.mode === 'fit-width') {
    const w1 = await pdfRef.value.getPageWidthAtScale1()
    if (!w1) return
    const paneWidth = scroller.clientWidth - 48
    const next = clampScale(paneWidth / w1)
    zoom.value = { mode: 'fit-width', scale: next }
  } else if (zoom.value.mode === 'fit-page') {
    const w1 = await pdfRef.value.getPageWidthAtScale1()
    const h1 = await pdfRef.value.getPageHeightAtScale1()
    if (!w1 || !h1) return
    const paneWidth = scroller.clientWidth - 48
    const paneHeight = scroller.clientHeight - 96 // room for control strip + breathing
    const next = clampScale(Math.min(paneWidth / w1, paneHeight / h1))
    zoom.value = { mode: 'fit-page', scale: next }
  }
}

function goPrev() {
  if (currentPage.value <= 1) return
  pdfRef.value?.scrollToPage(currentPage.value - 1)
}

function goNext() {
  if (totalPages.value > 0 && currentPage.value >= totalPages.value) return
  pdfRef.value?.scrollToPage(currentPage.value + 1)
}

function onPageInputType(v: string) {
  pageInputDraft.value = v.replace(/[^0-9]/g, '')
}

function commitPageInput(_e?: Event) {
  if (!pageInputFocused.value) return
  const n = Number(pageInputDraft.value)
  pageInputFocused.value = false
  if (Number.isFinite(n) && n >= 1 && (totalPages.value === 0 || n <= totalPages.value)) {
    pdfRef.value?.scrollToPage(n)
  }
  pageInputEl.value?.blur()
}

function cancelPageInput() {
  pageInputFocused.value = false
  pageInputEl.value?.blur()
}

function persistProgress() {
  // Save zoom/preferences to localStorage (cheap, instant)
  saveProgress(props.attachmentId, {
    lastPage: currentPage.value,
    totalPages: totalPages.value,
    zoom: zoom.value.scale,
    zoomMode: zoom.value.mode
  })
  // Skip server save while the splash is up — currentPage.value is still 1
  // and we'd clobber the user's real last page.
  if (loading.value) return
  // Save last page to server (authoritative source for cross-session restore)
  if (props.attachmentId && currentPage.value >= 1) {
    api.attachments.saveProgress(props.attachmentId, currentPage.value).catch(() => {
      // Network/server error — localStorage above is the fallback
    })
  }
}

// --- Reading ruler tracking ---

function onMouseMove(e: MouseEvent) {
  if (props.settings.ruler === 'off') return
  const rect = root.value?.getBoundingClientRect()
  if (!rect) return
  cursorY.value = e.clientY - rect.top
}

// --- Anchor management (operates on noteDoc JSON) ---

function emptyDoc(): DocNode {
  return { type: 'doc', content: [] }
}
function generateAnchorId() {
  return 'a_' + Math.random().toString(36).slice(2, 10) + Date.now().toString(36).slice(-4)
}

function collectAnchors(doc: DocNode | null): (PdfAnchor & { attachmentID: number })[] {
  const out: (PdfAnchor & { attachmentID: number })[] = []
  if (!doc) return out
  const stack: DocNode[] = [doc]
  while (stack.length) {
    const n = stack.pop()
    if (!n) continue
    if (n.type === 'pdfQuote' && n.attrs) {
      out.push({
        anchorId: String(n.attrs.anchorId ?? ''),
        page: Number(n.attrs.page ?? 1),
        text: String(n.attrs.text ?? ''),
        rects: (n.attrs.rects as PdfAnchor['rects']) ?? [],
        color: (n.attrs.color as PdfQuoteColor) ?? 'yellow',
        v: n.attrs.v ? Number(n.attrs.v) : undefined,
        attachmentID: Number(n.attrs.attachmentID ?? 0)
      })
    }
    if (Array.isArray(n.content)) {
      // walk in reverse so we get document order with stack-pop iteration
      for (let i = n.content.length - 1; i >= 0; i--) stack.push(n.content[i])
    }
  }
  return out
}

function findPdfQuoteIndex(anchorId: string): number | null {
  if (!noteDoc.value || !Array.isArray(noteDoc.value.content)) return null
  for (let i = 0; i < noteDoc.value.content.length; i++) {
    const n = noteDoc.value.content[i]
    if (n.type === 'pdfQuote' && n.attrs?.anchorId === anchorId) return i
  }
  return null
}

function onSelect(payload: { page: number; text: string; rects: PdfAnchor['rects']; color: PdfQuoteColor; v: number }) {
  if (!annotationsLive.value || !noteDoc.value) return
  if (noteLocked.value) {
    showToast('This note is locked. Unlock to capture notes.')
    return
  }
  if (!Array.isArray(noteDoc.value.content)) noteDoc.value.content = []
  const newNode: DocNode = {
    type: 'pdfQuote',
    attrs: {
      attachmentID: props.attachmentId,
      page: payload.page,
      text: payload.text,
      anchorId: generateAnchorId(),
      rects: payload.rects,
      color: payload.color,
      v: payload.v
    },
    content: [{ type: 'paragraph' }]
  }
  noteDoc.value.content.push(newNode)
  noteDoc.value = { ...noteDoc.value }
  scheduleSave()
  sessionCaptureCount.value += 1
  showToast(`Note saved · ${sessionCaptureCount.value} in this session`)
}

// Legacy highlight re-anchored by the viewer: persist corrected text + rects.
function onAnchorMigrated(m: PdfAnchorMigration) {
  if (noteLocked.value || !noteDoc.value) return
  const i = findPdfQuoteIndex(m.anchorId)
  if (i === null || !noteDoc.value.content) return
  const node = noteDoc.value.content[i]
  node.attrs = { ...node.attrs, rects: m.rects, text: m.text, v: m.v }
  noteDoc.value = { ...noteDoc.value }
  scheduleSave()
}

function onAnchorClick(anchorId: string) {
  drawerOpen.value = true
  expandedId.value = anchorId
  bumpDrawerActivity()
  nextTick(() => {
    const row = document.querySelector(`.dr-row[data-anchor="${CSS.escape(anchorId)}"]`)
    row?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  })
}

function onRowClick(anchorId: string) {
  // Toggle expand / collapse
  expandedId.value = expandedId.value === anchorId ? null : anchorId
  pdfRef.value?.scrollToAnchor(anchorId)
  bumpDrawerActivity()
}

function getCommentary(anchorId: string): string {
  const idx = findPdfQuoteIndex(anchorId)
  if (idx === null || !noteDoc.value) return ''
  const node = noteDoc.value.content?.[idx]
  if (!node?.content) return ''
  return node.content
    .filter(p => p.type === 'paragraph')
    .map(p => (p.content ?? []).map(c => String((c as DocNode & { text?: string }).text ?? '')).join(''))
    .join('\n')
}

function onCommentaryInput(anchorId: string, value: string) {
  const idx = findPdfQuoteIndex(anchorId)
  if (idx === null || !noteDoc.value || !noteDoc.value.content) return
  // TipTap text node format: { type: 'text', text: '...' }
  const paragraphs: DocNode[] = value.split('\n').map(line =>
    line
      ? { type: 'paragraph', content: [{ type: 'text', ...({ text: line } as object) } as DocNode] }
      : { type: 'paragraph' }
  )
  if (paragraphs.length === 0) paragraphs.push({ type: 'paragraph' })
  noteDoc.value.content[idx] = {
    ...noteDoc.value.content[idx],
    content: paragraphs
  }
  noteDoc.value = { ...noteDoc.value, content: [...noteDoc.value.content] }
  scheduleSave()
  bumpDrawerActivity()
}

function recolor(anchorId: string, color: PdfQuoteColor) {
  const idx = findPdfQuoteIndex(anchorId)
  if (idx === null || !noteDoc.value || !noteDoc.value.content) return
  const node = noteDoc.value.content[idx]
  noteDoc.value.content[idx] = {
    ...node,
    attrs: { ...(node.attrs ?? {}), color }
  }
  noteDoc.value = { ...noteDoc.value, content: [...noteDoc.value.content] }
  scheduleSave()
  bumpDrawerActivity()
}

function removeAnchor(anchorId: string) {
  const idx = findPdfQuoteIndex(anchorId)
  if (idx === null || !noteDoc.value || !noteDoc.value.content) return
  noteDoc.value.content.splice(idx, 1)
  noteDoc.value = { ...noteDoc.value, content: [...noteDoc.value.content] }
  if (expandedId.value === anchorId) expandedId.value = null
  scheduleSave()
  bumpDrawerActivity()
}

// --- Save (debounced + explicit flush) ---

function scheduleSave() {
  if (props.sourceId === null || !noteDoc.value) return
  if (saveDebounce) clearTimeout(saveDebounce)
  saveStatus.value = 'saving'
  saveDebounce = window.setTimeout(flushSave, 800)
}

async function flushSave() {
  if (props.sourceId === null || !noteDoc.value) return
  if (saveDebounce) { clearTimeout(saveDebounce); saveDebounce = null }
  const body = JSON.stringify(noteDoc.value)
  if (body === lastSavedJSON) {
    saveStatus.value = 'saved'
    return
  }
  try {
    await api.notes.upsert(props.sourceId, body)
    lastSavedJSON = body
    saveStatus.value = 'saved'
  } catch (e) {
    if (e instanceof Error && /locked/i.test(e.message)) {
      // Server flipped the lock (e.g. another tab). Reflect it locally and
      // drop the most-recent capture so the user sees consistent state.
      noteLocked.value = true
      saveStatus.value = 'locked'
      showToast('This note was locked elsewhere. Latest capture not saved.')
      return
    }
    saveStatus.value = 'error'
  }
}

// Periodic explicit flush (Tenet D — throttle-safe)
watch(() => Date.now(), () => { /* placeholder so timer-driven flush is simple */ })

// --- Text notes from the regular editor (non-pdfQuote content) ---

function extractNodeText(node: DocNode): string {
  if (node.type === 'text') return (node as DocNode & { text?: string }).text ?? ''
  const childText = (node.content ?? []).map(extractNodeText).join('')
  if (['paragraph', 'heading', 'listItem', 'blockquote'].includes(node.type)) {
    return childText + '\n'
  }
  return childText
}

const textNotes = computed<string>(() => {
  if (!noteDoc.value || !annotationsLive.value) return ''
  return (noteDoc.value.content ?? [])
    .filter(n => n.type !== 'pdfQuote')
    .map(extractNodeText)
    .join('')
    .trim()
})

// --- Drawer interactions ---

function toggleDrawer() {
  drawerOpen.value = !drawerOpen.value
  if (drawerOpen.value) bumpDrawerActivity()
  else if (drawerHideTimer) clearTimeout(drawerHideTimer)
}

function bumpDrawerActivity() {
  if (!drawerOpen.value || !props.settings.autoHideDrawer) return
  if (drawerHideTimer) clearTimeout(drawerHideTimer)
  drawerHideTimer = window.setTimeout(() => {
    drawerOpen.value = false
  }, 30_000)
}

watch(drawerOpen, (open) => {
  if (open) bumpDrawerActivity()
})

// --- Toast ---

function showToast(msg: string) {
  toast.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = window.setTimeout(() => { toast.value = '' }, 1800)
}

// --- Exit guard ---

function requestExit() { showExitGuard.value = true }
function resumeSession() {
  showExitGuard.value = false
  if (!document.fullscreenElement) enterFullscreen()
}
async function confirmExit() {
  const elapsed = totalMs.value > 0
    ? totalMs.value - Math.max(0, remainingMs.value)
    : Date.now() - sessionStartTime
  const elapsedSec = Math.round(elapsed / 1000)
  // Only log if the user spent at least 3 minutes — shorter exits are a grace period and don't count.
  if (elapsedSec >= 180) {
    api.focusSessions.log({ date: todayISO(), duration_sec: elapsedSec, attachment_id: props.attachmentId, filename: props.filename, completed: false }).catch(() => {})
  }
  exiting.value = true
  pendingExit = true
  persistProgress()
  // Wait for save to finish (max 2s) before tearing down (Tenet D)
  await Promise.race([
    flushSave(),
    new Promise(r => window.setTimeout(r, 2000))
  ])
  exiting.value = false
  showExitGuard.value = false
  if (document.fullscreenElement) {
    try { await document.exitFullscreen() } catch { /* ignore */ }
  }
  emit('exit')
}

// --- Keyboard shortcuts ---

function onKey(e: KeyboardEvent) {
  if (loading.value) return
  if (e.target instanceof HTMLTextAreaElement || e.target instanceof HTMLInputElement) return
  switch (e.key) {
    case '+':
    case '=':
      zoomIn(); e.preventDefault(); break
    case '-':
      zoomOut(); e.preventDefault(); break
    case '0':
      onFitModeChange('fit-width'); e.preventDefault(); break
    case 'PageDown':
    case 'ArrowRight':
    case ' ':
      goNext(); e.preventDefault(); break
    case 'PageUp':
    case 'ArrowLeft':
      goPrev(); e.preventDefault(); break
    case 'Home':
      pdfRef.value?.scrollToPage(1); e.preventDefault(); break
    case 'End':
      if (totalPages.value > 0) { pdfRef.value?.scrollToPage(totalPages.value); e.preventDefault() }
      break
    case 'n':
    case 'N':
      if (annotationsLive.value) { toggleDrawer(); e.preventDefault() }
      break
  }
}
</script>
