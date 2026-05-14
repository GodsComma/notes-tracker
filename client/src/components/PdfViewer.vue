<template>
  <div ref="root" class="pdf-viewer relative" :data-selection-disabled="!enableSelection || undefined" @scroll.passive="onScroll">
    <div v-if="loading" class="absolute inset-0 flex items-center justify-center text-sm text-slate-500">
      Loading PDF…
    </div>
    <div v-if="error" class="absolute inset-0 flex items-center justify-center text-sm text-red-500 px-4 text-center">
      {{ error }}
    </div>

    <!-- Zoom in progress — selection is paused during rescale -->
    <Transition name="pdf-hint">
      <div v-if="rescaling" class="pdf-rescaling-banner">
        Zooming — selection paused
      </div>
    </Transition>

    <div ref="pagesEl" class="pdf-pages" />

    <!-- Cross-page hint (Fix C) — positioned near where the selection ended -->
    <Transition name="pdf-hint">
      <div
        v-if="crossPageHint"
        class="pdf-cross-page-hint"
        :style="{ left: crossPageHint.x + 'px', top: crossPageHint.y + 'px', transform: 'translateX(-50%)' }"
      >
        Select within one page to highlight
      </div>
    </Transition>

    <!-- Floating selection tooltip -->
    <div
      v-if="tooltip && enableSelection"
      class="pdf-selection-tooltip"
      :style="{ left: tooltip.x + 'px', top: tooltip.y + 'px' }"
      @mousedown.prevent
    >
      <button
        v-for="c in PDF_QUOTE_COLORS"
        :key="c"
        type="button"
        class="pdf-color-swatch"
        :data-color="c"
        :title="`Highlight & note (${c})`"
        @click="commitSelection(c)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import * as pdfjsLib from 'pdfjs-dist'
import { PDF_QUOTE_COLORS, type PdfQuoteColor } from '@/extensions/PdfQuote'
import { acquireCachedPdf, releaseCachedPdf } from '@/composables/usePdfCache'

// Worker URL is set in usePdfCache (imported above); no duplicate needed here.

export interface PdfRect {
  x: number
  y: number
  w: number
  h: number
}

export interface PdfAnchor {
  anchorId: string
  page: number
  text: string
  rects: PdfRect[]
  color: PdfQuoteColor
}

const props = withDefaults(defineProps<{
  url: string
  anchors: PdfAnchor[]
  scale?: number
  enableSelection?: boolean
}>(), {
  scale: 1.5,
  enableSelection: true
})

const emit = defineEmits<{
  select: [payload: { page: number; text: string; rects: PdfRect[]; color: PdfQuoteColor }]
  'anchor-click': [anchorId: string]
  'page-change': [page: number]
  loaded: [info: { numPages: number }]
}>()

const root = ref<HTMLElement | null>(null)
const pagesEl = ref<HTMLElement | null>(null)
const loading = ref(true)
const error = ref('')

const tooltip = ref<{ x: number; y: number; page: number; rects: PdfRect[]; text: string } | null>(null)
const pageContainers = new Map<number, HTMLElement>()

// Tracks which pages have had their canvas fully rendered at least once.
// rescale only needs to re-render these; unrendered pages will be filled in
// by the background render loop at the new scale.
const renderedPages = new Set<number>()
const textLayerPages = new Set<number>() // pages whose textLayer has been fully built

let pdfDoc: pdfjsLib.PDFDocumentProxy | null = null
let usingCachedDoc = false  // true if pdfDoc came from the module cache

const renderedScale = ref(props.scale)
const rescaling = ref(false)
let rescaleTimer: number | null = null

// Resolvers awaiting the current rescale to finish (for rescaleDone()).
let rescaleResolvers: Array<() => void> = []

// Epoch guard: incremented on each rescale. Background render checks against
// the epoch it started with and stops if it's been superseded.
let renderEpoch = 0

let touchSelectionTimer: number | null = null
let pendingTouchCapture = false
let hoveredAnchorId: string | null = null
let lastPointerPos = { clientX: 0, clientY: 0 }

onMounted(() => {
  loadPdf()
  // Fix A: capture on pointerup (after drag ends) not on selectionchange (mid-drag).
  document.addEventListener('pointerup', onDocPointerUp)
  document.addEventListener('pointerdown', onDocPointerDown)
  document.addEventListener('mousemove', onDocMouseMove)
  // selectionchange is kept only to clear the tooltip when selection collapses.
  document.addEventListener('selectionchange', onSelectionChangeRaw)
})
watch(() => props.url, () => loadPdf())
watch(() => props.scale, (next) => {
  // Fix D: block selection capture immediately when scale changes.
  // scheduleRescale() also sets this, but doing it here closes the
  // ~300ms window where the CSS transform is live but rescaling is still false.
  rescaling.value = true
  // Phase 1 — instant CSS-scale preview at zero cost.
  const ratio = next / renderedScale.value
  for (const container of pageContainers.values()) {
    container.style.transformOrigin = 'top center'
    container.style.transform = `scale(${ratio})`
  }
  // Phase 2 — debounced real re-render at the new scale.
  scheduleRescale()
})
watch(() => props.anchors, () => renderAllHighlights(), { deep: true })

const currentPage = ref(1)
function onScroll() {
  if (!root.value) return
  const top = root.value.scrollTop
  let active = 1
  for (const [p, el] of pageContainers) {
    if (el.offsetTop - 80 <= top) active = p
    else break
  }
  if (active !== currentPage.value) {
    currentPage.value = active
    emit('page-change', active)
  }
}

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', onDocMouseDown, true)
  document.removeEventListener('pointerup', onDocPointerUp)
  document.removeEventListener('pointerdown', onDocPointerDown)
  document.removeEventListener('mousemove', onDocMouseMove)
  document.removeEventListener('selectionchange', onSelectionChangeRaw)
  if (rescaleTimer) clearTimeout(rescaleTimer)
  if (crossPageHintTimer) clearTimeout(crossPageHintTimer)
  if (touchSelectionTimer) clearTimeout(touchSelectionTimer)
  // Return the cached doc instead of destroying it — the cache manages its
  // own lifecycle. Destroy only if we fetched it ourselves.
  if (pdfDoc) {
    if (usingCachedDoc) {
      releaseCachedPdf(props.url)
    } else {
      pdfDoc.destroy()
    }
    pdfDoc = null
  }
})

// --- Loading ---

async function loadPdf() {
  if (!props.url || !pagesEl.value) return
  loading.value = true
  error.value = ''
  renderedPages.clear()
  textLayerPages.clear()
  pageContainers.clear()
  pagesEl.value.innerHTML = ''
  renderEpoch++
  const myEpoch = renderEpoch

  try {
    // Try module-level cache first (warmed by SourceView.preloadPdf).
    const cached = acquireCachedPdf(props.url)
    if (cached) {
      pdfDoc = cached
      usingCachedDoc = true
    } else {
      usingCachedDoc = false
      pdfDoc = await pdfjsLib.getDocument({ url: props.url }).promise
    }
    if (myEpoch !== renderEpoch) return  // URL changed while loading

    const numPages = pdfDoc.numPages

    // --- Step 1: Create placeholder containers for every page ---
    // We do this upfront so that scrollToPage(p) works for any page immediately,
    // even before that page has been canvas-rendered. We size them all using
    // page 1's dimensions; individual dimensions are corrected when rendered.
    const page1 = await pdfDoc.getPage(1)
    const vp1 = page1.getViewport({ scale: props.scale })
    for (let p = 1; p <= numPages; p++) {
      createPageContainer(p, vp1.width, vp1.height)
    }
    if (myEpoch !== renderEpoch) return

    // --- Step 2: Render pages 1–2 immediately, then emit loaded ---
    // These are the only pages the user sees on open. Everything else is
    // background-rendered so the splash can disappear within ~1s.
    await runPageRender(1, { withTextLayer: true })
    if (numPages >= 2) await runPageRender(2, { withTextLayer: true })
    if (myEpoch !== renderEpoch) return

    renderAllHighlights()
    document.addEventListener('mousedown', onDocMouseDown, true)
    loading.value = false
    emit('loaded', { numPages })

    // --- Step 3: Background-render remaining pages ---
    scheduleBackgroundRender(3, myEpoch)

  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load PDF'
    loading.value = false
  }
}

function createPageContainer(pageNum: number, width: number, height: number) {
  if (!pagesEl.value) return
  const container = document.createElement('div')
  container.className = 'pdf-page-container'
  container.dataset.pageNumber = String(pageNum)
  container.style.width = width + 'px'
  container.style.height = height + 'px'

  const canvas = document.createElement('canvas')
  container.appendChild(canvas)

  const textLayerDiv = document.createElement('div')
  textLayerDiv.className = 'textLayer'
  container.appendChild(textLayerDiv)

  const highlightDiv = document.createElement('div')
  highlightDiv.className = 'pdf-highlight-layer'
  container.appendChild(highlightDiv)

  const label = document.createElement('span')
  label.className = 'pdf-page-label'
  label.textContent = `${pageNum} / ${pdfDoc?.numPages ?? '?'}`
  container.appendChild(label)

  pagesEl.value.appendChild(container)
  pageContainers.set(pageNum, container)
}

// Per-page render tasks so we can cancel mid-flight on rapid zoom.
const renderTasks = new Map<number, { cancel: () => void }>()

async function runPageRender(pageNum: number, { withTextLayer = false } = {}) {
  if (!pdfDoc) return
  const container = pageContainers.get(pageNum)
  if (!container) return

  const prev = renderTasks.get(pageNum)
  if (prev) {
    try { prev.cancel() } catch { /* ignore */ }
    renderTasks.delete(pageNum)
  }

  const page = await pdfDoc.getPage(pageNum)
  const viewport = page.getViewport({ scale: props.scale })

  container.style.width = viewport.width + 'px'
  container.style.height = viewport.height + 'px'

  const canvas = container.querySelector('canvas') as HTMLCanvasElement | null
  if (!canvas) return
  canvas.width = viewport.width
  canvas.height = viewport.height
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const task = page.render({ canvasContext: ctx, viewport, canvas } as unknown as Parameters<typeof page.render>[0])
  renderTasks.set(pageNum, task)
  try {
    await task.promise
  } catch (e: unknown) {
    if ((e as { name?: string })?.name === 'RenderingCancelledException') return
    throw e
  } finally {
    if (renderTasks.get(pageNum) === task) renderTasks.delete(pageNum)
  }

  renderedPages.add(pageNum)

  // Text layer — rendered immediately for visible pages, deferred (withTextLayer=false)
  // for background-rendered pages to avoid blocking the render loop.
  if (withTextLayer) {
    await buildTextLayer(page, viewport, container)
  }
}

async function buildTextLayer(
  page: pdfjsLib.PDFPageProxy,
  viewport: pdfjsLib.PageViewport,
  container: HTMLElement
) {
  const textLayerDiv = container.querySelector('.textLayer') as HTMLElement | null
  if (!textLayerDiv) return
  textLayerDiv.innerHTML = ''
  textLayerDiv.style.width = viewport.width + 'px'
  textLayerDiv.style.height = viewport.height + 'px'
  container.classList.add('text-layer-building')
  try {
    const textContent = await page.getTextContent()
    const TextLayerCtor = (pdfjsLib as unknown as {
      TextLayer?: new (opts: unknown) => { render: () => Promise<void> }
    }).TextLayer
    if (!TextLayerCtor) {
      console.error('[PdfViewer] pdfjsLib.TextLayer not found — text selection unavailable. Check pdfjs-dist version.')
      return
    }
    const tl = new TextLayerCtor({ textContentSource: textContent, container: textLayerDiv, viewport })
    await tl.render()
    attachSnapHandler(textLayerDiv)
    const pageNum = Number(container.dataset.pageNumber ?? 0)
    if (pageNum) textLayerPages.add(pageNum)
  } catch { /* text layer is best-effort */ } finally {
    container.classList.remove('text-layer-building')
  }
}

// Background render: yields to the browser between batches so the UI stays
// responsive. Pages rendered here get withTextLayer=false; text layers are
// added lazily via IntersectionObserver when the page scrolls into view.
const BATCH_SIZE = 4

function scheduleBackgroundRender(startPage: number, epoch: number) {
  if (!pdfDoc || startPage > pdfDoc.numPages) return
  const end = Math.min(startPage + BATCH_SIZE - 1, pdfDoc.numPages)
  const run = async () => {
    for (let p = startPage; p <= end; p++) {
      if (epoch !== renderEpoch || !pdfDoc) return  // superseded by zoom or URL change
      try { await runPageRender(p, { withTextLayer: false }) } catch { /* ignore */ }
    }
    scheduleBackgroundRender(end + 1, epoch)
  }
  if ('requestIdleCallback' in window) {
    requestIdleCallback(run, { timeout: 5000 })
  } else {
    setTimeout(run, 0)
  }
}

// IntersectionObserver: when a page enters the viewport, build its text layer
// if it hasn't been built yet (i.e. was background-rendered without one).
let textLayerObserver: IntersectionObserver | null = null

function setupTextLayerObserver() {
  if (!root.value || textLayerObserver) return
  textLayerObserver = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (!entry.isIntersecting || !pdfDoc) continue
        const container = entry.target as HTMLElement
        const pageNum = Number(container.dataset.pageNumber ?? 0)
        if (!pageNum || !renderedPages.has(pageNum)) continue
        const textLayerDiv = container.querySelector('.textLayer') as HTMLElement | null
        // Only build if empty (background render left it blank) and not already in progress.
        if (!textLayerDiv || textLayerDiv.childElementCount > 0) continue
        if (container.classList.contains('text-layer-building')) continue
        pdfDoc.getPage(pageNum).then(page => {
          const vp = page.getViewport({ scale: props.scale })
          buildTextLayer(page, vp, container)
        }).catch(() => { /* best effort */ })
      }
    },
    { root: root.value, rootMargin: '400px 0px' }
  )
  for (const container of pageContainers.values()) {
    textLayerObserver.observe(container)
  }
}

// --- Rescale ---

// When set, the next rescale scrolls here instead of currentPage.
let pendingRestorePage: number | null = null
function setRestorePage(p: number) { pendingRestorePage = p }

async function rerenderAllPagesAtCurrentScale() {
  if (!pdfDoc || !root.value || renderedPages.size === 0) return
  const myEpoch = ++renderEpoch
  const anchorPage = pendingRestorePage ?? currentPage.value
  pendingRestorePage = null

  // Cancel in-flight tasks before re-rendering.
  for (const t of renderTasks.values()) {
    try { t.cancel() } catch { /* ignore */ }
  }
  renderTasks.clear()

  // Only re-render pages that have already been canvas-rendered. Unrendered
  // placeholders will be filled in correctly by the restarted background loop.
  for (const pageNum of [...renderedPages].sort((a, b) => a - b)) {
    if (myEpoch !== renderEpoch) return
    try { await runPageRender(pageNum, { withTextLayer: true }) } catch { /* swallow */ }
  }
  if (myEpoch !== renderEpoch) return

  renderAllHighlights()
  scrollToPage(anchorPage)

  // Flush rescaleDone() waiters.
  const resolvers = rescaleResolvers.splice(0)
  for (const resolve of resolvers) resolve()

  // Restart background render for pages that haven't been rendered yet.
  scheduleBackgroundRender(firstUnrenderedPage(), myEpoch)
}

function firstUnrenderedPage(): number {
  if (!pdfDoc) return Infinity
  for (let p = 1; p <= pdfDoc.numPages; p++) {
    if (!renderedPages.has(p)) return p
  }
  return Infinity
}

function scheduleRescale() {
  tooltip.value = null
  rescaling.value = true
  if (rescaleTimer) clearTimeout(rescaleTimer)
  rescaleTimer = window.setTimeout(async () => {
    rescaleTimer = null
    for (const container of pageContainers.values()) {
      container.style.transform = ''
    }
    await rerenderAllPagesAtCurrentScale()
    renderedScale.value = props.scale
    rescaling.value = false
  }, 300)
}

/**
 * Returns a Promise that resolves once the current rescale debounce + re-render
 * cycle completes. If no rescale is pending, resolves immediately.
 * Use this instead of hard-coded timeouts when waiting for zoom to settle.
 */
function rescaleDone(): Promise<void> {
  if (!rescaleTimer && !rescaling.value) return Promise.resolve()
  return new Promise<void>(resolve => { rescaleResolvers.push(resolve) })
}

async function getPageWidthAtScale1(): Promise<number> {
  if (!pdfDoc) return 0
  const page = await pdfDoc.getPage(1)
  return page.getViewport({ scale: 1 }).width
}

async function getPageHeightAtScale1(): Promise<number> {
  if (!pdfDoc) return 0
  const page = await pdfDoc.getPage(1)
  return page.getViewport({ scale: 1 }).height
}

// --- Highlights ---

function renderAllHighlights() {
  for (const [pageNum, container] of pageContainers) {
    if (!renderedPages.has(pageNum)) continue
    const layer = container.querySelector('.pdf-highlight-layer') as HTMLElement | null
    if (!layer) continue
    layer.innerHTML = ''
    const onPage = props.anchors.filter(a => a.page === pageNum)
    for (const a of onPage) {
      for (const r of a.rects) {
        const el = document.createElement('div')
        el.className = 'pdf-highlight'
        el.style.left = r.x * props.scale + 'px'
        el.style.top = r.y * props.scale + 'px'
        el.style.width = r.w * props.scale + 'px'
        el.style.height = r.h * props.scale + 'px'
        el.dataset.anchorId = a.anchorId
        el.dataset.color = a.color ?? 'yellow'
        layer.appendChild(el)
      }
    }
  }
}

// --- Selection capture ---

function onDocMouseDown(e: MouseEvent) {
  const target = e.target as HTMLElement | null
  if (target?.closest('.pdf-selection-tooltip')) return
  tooltip.value = null
}

function findPageContainer(node: Node | null): HTMLElement | null {
  let n: Node | null = node
  while (n) {
    if (n instanceof HTMLElement && n.classList.contains('pdf-page-container')) return n
    n = n.parentNode
  }
  return null
}

// Fix B: merge per-span rects from PDF.js TextLayer into one rect per visual
// line. PDF.js positions each text span with a CSS matrix transform, so
// getClientRects() returns one rect per span — 40–120 for a paragraph. Merging
// by vertical proximity reduces that to 3–5 line-level rects, eliminates
// subpixel gaps between adjacent spans, and cuts stored data dramatically.
//
// Both thresholds are now scale-aware:
// - LINE_TOLERANCE grows with scale so same-line subscripts/superscripts stay merged
//   at high zoom (where their vertical offsets are proportionally larger in CSS px).
// - HORIZ_GAP grows with scale so inter-word gaps (which are proportionally larger
//   at high zoom) still merge into full-line rects instead of per-word rects.
function mergeRects(raw: DOMRect[], scale: number): DOMRect[] {
  if (raw.length <= 1) return raw
  const dpr = window.devicePixelRatio ?? 1
  const LINE_TOLERANCE = Math.max(2, (4 / dpr) * Math.max(1, scale))  // scale-and-DPI-aware
  const HORIZ_GAP      = Math.round(8 * Math.max(1, scale))            // proportional to zoom
  const sorted = [...raw].sort((a, b) => a.top - b.top || a.left - b.left)
  const merged: DOMRect[] = []
  let cur = sorted[0]
  for (let i = 1; i < sorted.length; i++) {
    const r = sorted[i]
    const sameTop  = Math.abs(r.top - cur.top) <= LINE_TOLERANCE
    const adjacent = r.left <= cur.right + HORIZ_GAP
    if (sameTop && adjacent) {
      const right  = Math.max(cur.right,  r.right)
      const bottom = Math.max(cur.bottom, r.bottom)
      cur = new DOMRect(cur.left, cur.top, right - cur.left, bottom - cur.top)
    } else {
      merged.push(cur)
      cur = r
    }
  }
  merged.push(cur)
  return merged
}

// Fix A: selectionchange now only clears the tooltip when the user clicks
// away and collapses the selection. Actual rect capture happens in captureSelection
// which is triggered by pointerup — after the drag has fully ended.
// For touch/pen: reset debounce timer while selection handles are still moving.
function onSelectionChangeRaw() {
  const sel = window.getSelection()
  if (!sel || sel.isCollapsed) {
    tooltip.value = null
    return
  }
  if (pendingTouchCapture) {
    if (touchSelectionTimer) clearTimeout(touchSelectionTimer)
    touchSelectionTimer = window.setTimeout(() => {
      pendingTouchCapture = false
      touchSelectionTimer = null
      captureSelection(lastPointerPos.clientX, lastPointerPos.clientY)
    }, 80)
  }
}

// Fix A: pointerup fires exactly once when the mouse button is released.
// Mouse: use requestAnimationFrame so the browser has flushed the final selection
// state (fixes RC-4 scroll race). Also detects anchor clicks on collapsed selections.
// Touch/pen: selection state isn't committed at pointerup time on iOS/Android;
// debounce via selectionchange until handles stop moving.
function onDocPointerUp(e: PointerEvent) {
  const { clientX, clientY } = e
  lastPointerPos = { clientX, clientY }
  if (e.pointerType === 'mouse') {
    if (e.button !== 0) return  // ignore non-left
    // Geometry-based anchor-click: fires only on a plain click (collapsed selection).
    // Drag-to-select leaves a non-collapsed selection, so we skip anchor detection.
    const sel = window.getSelection()
    if (sel?.isCollapsed && root.value && props.anchors.length) {
      for (const [pageNum, container] of pageContainers) {
        const pageRect = container.getBoundingClientRect()
        if (clientX < pageRect.left || clientX > pageRect.right ||
            clientY < pageRect.top  || clientY > pageRect.bottom) continue
        const relX = (clientX - pageRect.left) / props.scale
        const relY = (clientY - pageRect.top)  / props.scale
        let hit = false
        for (const anchor of props.anchors) {
          if (anchor.page !== pageNum) continue
          for (const r of anchor.rects) {
            if (relX >= r.x && relX <= r.x + r.w && relY >= r.y && relY <= r.y + r.h) {
              emit('anchor-click', anchor.anchorId)
              hit = true
              break
            }
          }
          if (hit) break
        }
        if (hit) break
      }
    }
    window.requestAnimationFrame(() => captureSelection(clientX, clientY))
  } else {
    pendingTouchCapture = true
    if (touchSelectionTimer) clearTimeout(touchSelectionTimer)
    touchSelectionTimer = window.setTimeout(() => {
      pendingTouchCapture = false
      touchSelectionTimer = null
      captureSelection(clientX, clientY)
    }, 80)
  }
}

function onDocPointerDown(e: PointerEvent) {
  if (e.pointerType !== 'mouse') {
    if (touchSelectionTimer) {
      clearTimeout(touchSelectionTimer)
      touchSelectionTimer = null
    }
    pendingTouchCapture = false
  }
}

function onDocMouseMove(e: MouseEvent) {
  if (!root.value || !props.anchors.length) return
  let found: string | null = null
  for (const [pageNum, container] of pageContainers) {
    const pageRect = container.getBoundingClientRect()
    if (e.clientX < pageRect.left || e.clientX > pageRect.right ||
        e.clientY < pageRect.top  || e.clientY > pageRect.bottom) continue
    const relX = (e.clientX - pageRect.left) / props.scale
    const relY = (e.clientY - pageRect.top)  / props.scale
    for (const anchor of props.anchors) {
      if (anchor.page !== pageNum) continue
      for (const r of anchor.rects) {
        if (relX >= r.x && relX <= r.x + r.w && relY >= r.y && relY <= r.y + r.h) {
          found = anchor.anchorId
          break
        }
      }
      if (found) break
    }
    if (found) break
  }
  if (found === hoveredAnchorId) return
  if (hoveredAnchorId) {
    const page = props.anchors.find(a => a.anchorId === hoveredAnchorId)?.page ?? -1
    pageContainers.get(page)
      ?.querySelectorAll(`[data-anchor-id="${CSS.escape(hoveredAnchorId)}"]`)
      .forEach(el => el.classList.remove('pdf-highlight-hover'))
  }
  hoveredAnchorId = found
  if (found) {
    const page = props.anchors.find(a => a.anchorId === found)?.page ?? -1
    pageContainers.get(page)
      ?.querySelectorAll(`[data-anchor-id="${CSS.escape(found)}"]`)
      .forEach(el => el.classList.add('pdf-highlight-hover'))
  }
}

// Cross-browser helper: find the text caret position nearest to (clientX, clientY).
// Returns a collapsed Range or null if neither API is available or no text was found.
function caretAtPoint(clientX: number, clientY: number): Range | null {
  if (typeof document.caretRangeFromPoint === 'function') {
    // Chrome, Safari, Edge
    return document.caretRangeFromPoint(clientX, clientY)
  }
  // Firefox
  const pos = (document as unknown as {
    caretPositionFromPoint?: (x: number, y: number) => { offsetNode: Node; offset: number } | null
  }).caretPositionFromPoint?.(clientX, clientY)
  if (!pos) return null
  const r = document.createRange()
  r.setStart(pos.offsetNode, pos.offset)
  r.collapse(true)
  return r
}

// Attaches a mousedown handler to a textLayer div that snaps gap clicks to the
// nearest character. Called after every buildTextLayer so rescales get a fresh
// handler without stacking.
//
// When a click lands on the textLayer div background (an inter-word or inter-line
// gap — no span at that position), the browser creates a collapsed selection at
// the div's empty flow content, effectively selecting nothing. This handler
// intercepts those gap clicks, uses caretRangeFromPoint to snap the anchor to the
// nearest character, then extends the selection on drag via Selection.extend().
// Without preventDefault the browser continues the drag from the snapped anchor.
function attachSnapHandler(textLayerDiv: HTMLElement) {
  // Reuse the stored reference so repeated calls don't stack listeners.
  const prev = (textLayerDiv as HTMLElement & { _snapHandler?: EventListener })._snapHandler
  if (prev) textLayerDiv.removeEventListener('mousedown', prev)

  const handler = (e: MouseEvent) => {
    // Only fire when the click target is the textLayer div itself (the background),
    // not when it's a child <span> or <br> — those have native selection already.
    if (e.target !== textLayerDiv) return

    const range = caretAtPoint(e.clientX, e.clientY)
    if (!range || !textLayerDiv.contains(range.startContainer)) return

    const sel = window.getSelection()
    if (!sel) return
    sel.removeAllRanges()
    sel.addRange(range)
    // Do NOT preventDefault — the browser must receive the mousedown to track
    // the drag start position from our snapped anchor.

    // Extend the selection on drag so moving through gaps also snaps correctly.
    const onDragMove = (ev: MouseEvent) => {
      if (!(ev.buttons & 1)) return  // bail if left button released
      const end = caretAtPoint(ev.clientX, ev.clientY)
      if (!end || !textLayerDiv.contains(end.startContainer)) return
      try {
        sel.extend(end.startContainer, end.startOffset)
      } catch { /* extend can throw across disconnected nodes — safe to ignore */ }
    }
    const onDragEnd = () => {
      document.removeEventListener('mousemove', onDragMove)
      document.removeEventListener('mouseup', onDragEnd)
    }
    document.addEventListener('mousemove', onDragMove)
    document.addEventListener('mouseup', onDragEnd)
  }

  ;(textLayerDiv as HTMLElement & { _snapHandler?: EventListener })._snapHandler = handler as EventListener
  textLayerDiv.addEventListener('mousedown', handler)
}

function captureSelection(clientX = 0, clientY = 0) {
  if (!props.enableSelection || rescaling.value || !root.value) return

  const sel = window.getSelection()
  if (!sel || sel.isCollapsed || sel.rangeCount === 0) return
  const range = sel.getRangeAt(0)

  // Fix G: ignore selections inside the tooltip itself.
  const tooltipEl = root.value.querySelector('.pdf-selection-tooltip')
  if (tooltipEl?.contains(range.commonAncestorContainer)) return

  // Fix C: identify which page each selection boundary is on. Reject if they
  // differ — cross-page rects are relative to the wrong pageRect and produce
  // wildly wrong coordinates.
  const startPage = findPageContainer(range.startContainer)
  const endPage   = findPageContainer(range.endContainer)
  if (!startPage) return
  if (endPage && endPage !== startPage) {
    // Brief visual feedback so the user knows why nothing happened.
    showCrossPageHint(clientX, clientY)
    return
  }
  const pageEl  = startPage
  const pageNum = Number(pageEl.dataset.pageNumber ?? 0)
  if (!pageNum) return

  // Fix B: all three reads happen in the same RAF frame — no scroll race.
  const rootRect  = root.value.getBoundingClientRect()
  const pageRect  = pageEl.getBoundingClientRect()
  const rawRects  = Array.from(range.getClientRects()).filter(r => r.width > 1 && r.height > 1)
  if (rawRects.length === 0) return

  const lineRects = mergeRects(rawRects, props.scale)

  const rects: PdfRect[] = lineRects.map(r => ({
    x: (r.left - pageRect.left) / props.scale,
    y: (r.top  - pageRect.top)  / props.scale,
    w: r.width  / props.scale,
    h: r.height / props.scale
  }))

  // Fix F: use range.toString() scoped to the range, then normalise whitespace.
  const text = range.toString().replace(/\s+/g, ' ').trim()
  if (!text) return

  const first = lineRects[0]
  // Offset the tooltip above the first line. At high zoom the line height grows
  // proportionally, so a fixed 32px can overlap the selection highlight. Use
  // the actual line height + 8px clearance, floored at 32px for small text.
  const tooltipOffset = Math.max(32, first.height + 8)
  tooltip.value = {
    x: first.left - rootRect.left + root.value.scrollLeft,
    y: first.top  - rootRect.top  + root.value.scrollTop - tooltipOffset,
    page: pageNum,
    rects,
    text
  }
}

// Hint shown when selection spans two pages — positioned near the cursor.
let crossPageHintTimer: number | null = null
const crossPageHint = ref<{ x: number; y: number } | null>(null)
function showCrossPageHint(clientX: number, clientY: number) {
  if (!root.value) return
  const rootRect = root.value.getBoundingClientRect()
  crossPageHint.value = {
    x: clientX - rootRect.left + root.value.scrollLeft,
    y: clientY - rootRect.top  + root.value.scrollTop - 40
  }
  if (crossPageHintTimer) clearTimeout(crossPageHintTimer)
  crossPageHintTimer = window.setTimeout(() => { crossPageHint.value = null }, 2000)
}

function commitSelection(color: PdfQuoteColor) {
  if (!tooltip.value) return
  const { page, rects, text } = tooltip.value
  if (text) emit('select', { page, text, rects, color })
  tooltip.value = null
  window.getSelection()?.removeAllRanges()
}

// --- External API ---

function scrollToAnchor(anchorId: string) {
  const anchor = props.anchors.find(a => a.anchorId === anchorId)
  if (!anchor) return
  const container = pageContainers.get(anchor.page)
  if (!container || !root.value) return

  const pageRect = container.getBoundingClientRect()
  const rootRect = root.value.getBoundingClientRect()
  const offsetTop = pageRect.top - rootRect.top + root.value.scrollTop
  const targetTop = offsetTop + (anchor.rects[0]?.y ?? 0) * props.scale - 80
  root.value.scrollTo({ top: Math.max(0, targetTop), behavior: 'smooth' })

  nextTick(() => {
    const els = container.querySelectorAll(`[data-anchor-id="${CSS.escape(anchorId)}"]`)
    els.forEach(el => {
      el.classList.add('pdf-highlight-active')
      window.setTimeout(() => el.classList.remove('pdf-highlight-active'), 1500)
    })
  })
}

function scrollToPage(p: number) {
  const container = pageContainers.get(p)
  if (!container || !root.value) return
  const rootRect = root.value.getBoundingClientRect()
  const pageRect = container.getBoundingClientRect()
  const targetTop = pageRect.top - rootRect.top + root.value.scrollTop - 24
  root.value.scrollTop = Math.max(0, targetTop)

  // Eagerly build text layer if the page is canvas-rendered but has no text layer yet.
  // Fire-and-forget — selection becomes available as soon as the async build completes.
  if (pdfDoc && renderedPages.has(p) && !textLayerPages.has(p)) {
    pdfDoc.getPage(p).then(page => {
      const vp = page.getViewport({ scale: props.scale })
      buildTextLayer(page, vp, container)
    }).catch(() => { /* best effort */ })
  }
}

// Wire up the IntersectionObserver once loadPdf puts containers in the DOM.
watch(() => loading.value, (isLoading) => {
  if (!isLoading) nextTick(setupTextLayerObserver)
})

defineExpose({
  scrollToAnchor,
  scrollToPage,
  setRestorePage,
  rescaleDone,
  currentPage,
  getPageWidthAtScale1,
  getPageHeightAtScale1
})
</script>

<style scoped>
.pdf-pages {
  padding-bottom: 1rem;
}

.pdf-cross-page-hint {
  position: absolute;
  /* left/top/transform set by :style binding — positioned near cursor */
  background: rgba(15, 23, 42, 0.85);
  color: #f1f5f9;
  font-size: 12px;
  padding: 6px 14px;
  border-radius: 9999px;
  pointer-events: none;
  z-index: 20;
  white-space: nowrap;
}

.pdf-hint-enter-active,
.pdf-hint-leave-active {
  transition: opacity 0.2s ease;
}
.pdf-hint-enter-from,
.pdf-hint-leave-to {
  opacity: 0;
}
</style>
