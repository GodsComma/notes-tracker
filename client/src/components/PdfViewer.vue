<!--
  PDF viewer with text selection → highlight capture.

  Rendering, lazy page loading, zoom and the text layer are pdf.js's own
  PDFViewer (pdfjs-dist/web/pdf_viewer.mjs) — the viewer Firefox ships. This
  component adds the app layer on top: highlight overlays, selection capture,
  legacy-anchor migration and hints. See Spec/spec-pdf-text-selection.md.

  Units: `scale` is CSS px per PDF point (1 = 72 dpi). pdf.js's own "100%" is
  96 dpi, so its currentScale = scale / PDF_TO_CSS_UNITS and --scale-factor on
  the viewer equals `scale`. Anchor rects are PDF points, top-left origin.
-->
<template>
  <div class="pdf-viewer-host">
    <div
      ref="root"
      class="pdf-viewer"
      :data-selection-disabled="!enableSelection || undefined"
    >
      <div ref="viewerEl" class="pdfViewer removePageBorders" />

      <Transition name="pdf-hint">
        <div
          v-if="hint"
          :class="hint.kind === 'cross-page' ? 'pdf-cross-page-hint' : 'pdf-no-text-hint'"
          class="pdf-hint"
          :style="{ left: hint.x + 'px', top: hint.y + 'px' }"
        >
          {{ hint.kind === 'cross-page'
            ? 'Select within one page to highlight'
            : 'This page is an image — it has no selectable text' }}
        </div>
      </Transition>

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

    <div v-if="loading" class="pdf-viewer-status text-slate-500">Loading PDF…</div>
    <div v-if="error" class="pdf-viewer-status text-red-500 px-4 text-center">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import { pdfjsLib } from '@/lib/pdfjs'
import { EventBus, PDFLinkService, PDFViewer } from '@/lib/pdfjsViewer'
import 'pdfjs-dist/web/pdf_viewer.css'
import { PDF_QUOTE_COLORS, PDF_ANCHOR_VERSION, type PdfQuoteColor } from '@/extensions/PdfQuote'
import { acquireCachedPdf, releaseCachedPdf } from '@/composables/usePdfCache'

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
  /**
   * Anchor format version. Missing = captured before the text-layer fix: those
   * rects mark where the user dragged but the stored text is shifted, so they
   * are re-anchored by geometry (see migrateLegacyAnchors).
   */
  v?: number
}

export interface PdfAnchorMigration {
  anchorId: string
  rects: PdfRect[]
  text: string
  v: number
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
  select: [payload: { page: number; text: string; rects: PdfRect[]; color: PdfQuoteColor; v: number }]
  'anchor-click': [anchorId: string]
  'anchor-migrated': [payload: PdfAnchorMigration]
  'page-change': [page: number]
  loaded: [info: { numPages: number }]
}>()

const PDF_TO_CSS = pdfjsLib.PixelsPerInch.PDF_TO_CSS_UNITS

const root = ref<HTMLElement | null>(null)
const viewerEl = ref<HTMLElement | null>(null)
const loading = ref(true)
const error = ref('')
const currentPage = ref(1)

const tooltip = ref<{ x: number; y: number; page: number; rects: PdfRect[]; text: string } | null>(null)
const hint = ref<{ kind: 'cross-page' | 'no-text'; x: number; y: number } | null>(null)
let hintTimer: number | null = null

let viewer: PDFViewer | null = null
let eventBus: EventBus | null = null
let linkService: PDFLinkService | null = null
let pdfDoc: pdfjsLib.PDFDocumentProxy | null = null
let docUrl: string | null = null
let usingCachedDoc = false
let loadEpoch = 0

const textLayerPages = new Set<number>()   // pages whose text layer has rendered
const noTextPages = new Set<number>()      // rendered text layer with zero spans (scanned/image page)
const migratedAnchors = new Map<string, PdfAnchorMigration>()
const migrationAttempted = new Set<string>()

// rescaleDone(): resolves once the first page re-render after a scale change.
let scalePending = false
let rescaleResolvers: Array<() => void> = []

let touchSelectionTimer: number | null = null
let pendingTouchCapture = false
let hoveredAnchorId: string | null = null
let lastPointerPos = { clientX: 0, clientY: 0 }

// ─── Lifecycle ────────────────────────────────────────────────────────────────

onMounted(() => {
  if (!root.value || !viewerEl.value) return
  eventBus = new EventBus()
  linkService = new PDFLinkService({ eventBus })
  viewer = new PDFViewer({
    container: root.value as HTMLDivElement,
    viewer: viewerEl.value as HTMLDivElement,
    eventBus,
    linkService,
    textLayerMode: 1,                                // enabled
    annotationMode: pdfjsLib.AnnotationMode.DISABLE, // app draws its own highlights
    removePageBorders: true,
  })
  linkService.setViewer(viewer)

  eventBus.on('pagesinit', onPagesInit)
  eventBus.on('pagerendered', onPageRendered)
  eventBus.on('textlayerrendered', onTextLayerRendered)
  eventBus.on('pagechanging', (e: { pageNumber: number }) => {
    if (e.pageNumber !== currentPage.value) {
      currentPage.value = e.pageNumber
      emit('page-change', e.pageNumber)
    }
  })
  eventBus.on('scalechanging', () => { tooltip.value = null })

  document.addEventListener('pointerup', onDocPointerUp)
  document.addEventListener('pointerdown', onDocPointerDown)
  document.addEventListener('pointermove', onDocPointerMove)
  document.addEventListener('mousedown', onDocMouseDown, true)
  document.addEventListener('mousemove', onDocMouseMove)
  document.addEventListener('selectionchange', onSelectionChange)

  loadPdf()
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerup', onDocPointerUp)
  document.removeEventListener('pointerdown', onDocPointerDown)
  document.removeEventListener('pointermove', onDocPointerMove)
  document.removeEventListener('mousedown', onDocMouseDown, true)
  document.removeEventListener('mousemove', onDocMouseMove)
  document.removeEventListener('selectionchange', onSelectionChange)
  if (hintTimer) clearTimeout(hintTimer)
  if (touchSelectionTimer) clearTimeout(touchSelectionTimer)
  flushRescaleResolvers()
  releaseDoc()
  viewer = null
})

watch(() => props.url, () => loadPdf())

watch(() => props.scale, (s) => {
  if (!viewer?.pdfDocument) return
  const next = s / PDF_TO_CSS
  if (Math.abs(viewer.currentScale - next) < 1e-6) return
  tooltip.value = null
  scalePending = true
  viewer.currentScale = next
  // Safety net: never leave rescaleDone() waiters hanging.
  window.setTimeout(() => { if (scalePending) { scalePending = false; flushRescaleResolvers() } }, 3000)
})

watch(() => props.anchors, () => {
  for (const p of textLayerPages) migrateLegacyAnchors(p)
  renderAllHighlights()
}, { deep: true })

// ─── Loading ──────────────────────────────────────────────────────────────────

function releaseDoc() {
  viewer?.setDocument(null as unknown as pdfjsLib.PDFDocumentProxy)
  linkService?.setDocument(null)
  if (pdfDoc && docUrl) {
    if (usingCachedDoc) releaseCachedPdf(docUrl)
    else pdfDoc.loadingTask.destroy()
  }
  pdfDoc = null
  docUrl = null
}

async function loadPdf() {
  if (!props.url || !viewer) return
  const epoch = ++loadEpoch
  loading.value = true
  error.value = ''
  tooltip.value = null
  textLayerPages.clear()
  noTextPages.clear()
  migratedAnchors.clear()
  migrationAttempted.clear()
  releaseDoc()

  try {
    const url = props.url
    const cached = acquireCachedPdf(url)
    const doc = cached ?? await pdfjsLib.getDocument({ url }).promise
    if (epoch !== loadEpoch) {
      if (cached) releaseCachedPdf(url); else doc.loadingTask.destroy()
      return
    }
    pdfDoc = doc
    docUrl = url
    usingCachedDoc = !!cached
    viewer.setDocument(doc)
    linkService?.setDocument(doc, null)
  } catch (e: unknown) {
    if (epoch !== loadEpoch) return
    error.value = e instanceof Error ? e.message : 'Failed to load PDF'
    loading.value = false
  }
}

function onPagesInit() {
  if (!viewer || !pdfDoc) return
  viewer.currentScale = props.scale / PDF_TO_CSS
  loading.value = false
  emit('loaded', { numPages: pdfDoc.numPages })
}

// ─── Page decorations (highlight layer + label) ───────────────────────────────

function pageView(pageNum: number) {
  return viewer?.getPageView(pageNum - 1) as
    | { div: HTMLDivElement; viewport: { width: number; height: number; scale: number }; textLayer?: { div: HTMLDivElement } | null }
    | undefined
}

/** CSS px per PDF point for a page, measured from its rendered box. */
function pxPerUnit(pageNum: number, pageRect: DOMRect): number {
  const vp = pageView(pageNum)?.viewport
  // viewport.scale already includes PDF_TO_CSS, so width / scale = width in points.
  const w1 = vp ? vp.width / vp.scale : 0
  return w1 ? pageRect.width / w1 : props.scale
}

/** pdf.js removes unknown children when it re-renders a page, so re-add ours. */
function decoratePage(pageNum: number) {
  const pv = pageView(pageNum)
  if (!pv || !pdfDoc) return
  const div = pv.div
  if (!div.querySelector(':scope > .pdf-highlight-layer')) {
    const layer = document.createElement('div')
    layer.className = 'pdf-highlight-layer'
    div.appendChild(layer)
  }
  if (!div.querySelector(':scope > .pdf-page-label')) {
    const label = document.createElement('span')
    label.className = 'pdf-page-label'
    label.textContent = `${pageNum} / ${pdfDoc.numPages}`
    div.appendChild(label)
  }
  renderHighlights(pageNum)
}

function onPageRendered(e: { pageNumber: number }) {
  decoratePage(e.pageNumber)
  if (scalePending) {
    scalePending = false
    flushRescaleResolvers()
  }
}

function onTextLayerRendered(e: { pageNumber: number }) {
  const pv = pageView(e.pageNumber)
  const tl = pv?.textLayer?.div
  if (!tl) return
  textLayerPages.add(e.pageNumber)
  if (tl.querySelector('span:not(.markedContent)')) noTextPages.delete(e.pageNumber)
  else noTextPages.add(e.pageNumber)
  migrateLegacyAnchors(e.pageNumber)
}

function flushRescaleResolvers() {
  const rs = rescaleResolvers.splice(0)
  for (const r of rs) r()
}

// ─── Highlights ───────────────────────────────────────────────────────────────

/** Rects to draw for an anchor: migrated ones for legacy anchors when available. */
function anchorRects(a: PdfAnchor): PdfRect[] {
  if (!a.v) return migratedAnchors.get(a.anchorId)?.rects ?? a.rects
  return a.rects
}

function renderHighlights(pageNum: number) {
  const layer = pageView(pageNum)?.div.querySelector(':scope > .pdf-highlight-layer')
  if (!layer) return
  layer.replaceChildren()
  for (const a of props.anchors) {
    if (a.page !== pageNum) continue
    for (const r of anchorRects(a)) {
      const el = document.createElement('div')
      el.className = 'pdf-highlight'
      // Page units × the viewer's --scale-factor (= props.scale): highlights
      // follow every zoom without being recomputed.
      el.style.left = `calc(var(--scale-factor) * ${r.x}px)`
      el.style.top = `calc(var(--scale-factor) * ${r.y}px)`
      el.style.width = `calc(var(--scale-factor) * ${r.w}px)`
      el.style.height = `calc(var(--scale-factor) * ${r.h}px)`
      el.dataset.anchorId = a.anchorId
      el.dataset.color = a.color ?? 'yellow'
      layer.appendChild(el)
    }
  }
}

function renderAllHighlights() {
  if (!pdfDoc) return
  const pages = new Set(props.anchors.map(a => a.page))
  // Clear layers on pages that no longer have anchors.
  viewerEl.value?.querySelectorAll('.page').forEach(div => {
    const n = Number((div as HTMLElement).dataset.pageNumber)
    if (n && !pages.has(n)) div.querySelector(':scope > .pdf-highlight-layer')?.replaceChildren()
  })
  for (const p of pages) if (p >= 1 && p <= pdfDoc.numPages) renderHighlights(p)
}

/** Anchor under a client point, if any (geometry-based; highlights don't take events). */
function anchorAt(clientX: number, clientY: number): string | null {
  for (const pageNum of new Set(props.anchors.map(a => a.page))) {
    const div = pageView(pageNum)?.div
    if (!div) continue
    const pr = div.getBoundingClientRect()
    if (clientX < pr.left || clientX > pr.right || clientY < pr.top || clientY > pr.bottom) continue
    const k = pxPerUnit(pageNum, pr)
    const x = (clientX - pr.left) / k, y = (clientY - pr.top) / k
    for (const a of props.anchors) {
      if (a.page !== pageNum) continue
      if (anchorRects(a).some(r => x >= r.x && x <= r.x + r.w && y >= r.y && y <= r.y + r.h)) return a.anchorId
    }
  }
  return null
}

// ─── Geometry ─────────────────────────────────────────────────────────────────

/**
 * Merge a range's client rects into one rect per visual line, in page units.
 * Thresholds are relative to line height, so they hold at any zoom and font
 * size: rects share a line when they overlap vertically by at least half the
 * smaller height, and join when the gap is under one line height (word gaps
 * between text items merge; column gutters don't).
 */
function mergeLineRects(rects: PdfRect[]): PdfRect[] {
  if (rects.length <= 1) return rects
  const sorted = [...rects].sort((a, b) => a.y - b.y || a.x - b.x)
  const lines: PdfRect[][] = []
  for (const r of sorted) {
    const line = lines.find(l => {
      const ref = l[0]
      const overlap = Math.min(ref.y + ref.h, r.y + r.h) - Math.max(ref.y, r.y)
      return overlap >= 0.5 * Math.min(ref.h, r.h)
    })
    if (line) line.push(r)
    else lines.push([r])
  }
  const out: PdfRect[] = []
  for (const l of lines) {
    l.sort((a, b) => a.x - b.x)
    let cur = { ...l[0] }
    for (let i = 1; i < l.length; i++) {
      const r = l[i]
      if (r.x <= cur.x + cur.w + Math.max(cur.h, r.h)) {
        const right = Math.max(cur.x + cur.w, r.x + r.w)
        const top = Math.min(cur.y, r.y)
        const bottom = Math.max(cur.y + cur.h, r.y + r.h)
        cur = { x: cur.x, y: top, w: right - cur.x, h: bottom - top }
      } else {
        out.push(cur)
        cur = { ...r }
      }
    }
    out.push(cur)
  }
  return out.sort((a, b) => a.y - b.y || a.x - b.x)
}

/** Page-unit line rects for a Range that lies within one page. */
function rectsFromRange(range: Range, pageEl: HTMLElement, pageNum: number): PdfRect[] {
  const pr = pageEl.getBoundingClientRect()
  const k = pxPerUnit(pageNum, pr)
  const raw = Array.from(range.getClientRects())
    .filter(r => r.width > 1 && r.height > 1)
    .map(r => ({ x: (r.left - pr.left) / k, y: (r.top - pr.top) / k, w: r.width / k, h: r.height / k }))
  return mergeLineRects(raw)
}

function pageOf(node: Node | null): HTMLElement | null {
  const el = node instanceof Element ? node : node?.parentElement
  const page = el?.closest('.page[data-page-number]') as HTMLElement | null
  return page && viewerEl.value?.contains(page) ? page : null
}

// ─── Legacy anchor migration ──────────────────────────────────────────────────

/**
 * Re-anchor highlights captured before the text-layer fix (no `v`).
 *
 * Their stored *text* is wrong (the misaligned layer shifted it by a
 * zoom-dependent amount) but their *rects* mark where the user dragged. So we
 * take the characters of the now-correct text layer under those rects and
 * rebuild text + rects from them. A character counts when its vertical centre
 * is inside a rect and at least 30% of its width overlaps it.
 */
function migrateLegacyAnchors(pageNum: number) {
  const legacy = props.anchors.filter(a => a.page === pageNum && !a.v && !migrationAttempted.has(a.anchorId))
  if (!legacy.length) return
  const pv = pageView(pageNum)
  const textLayer = pv?.textLayer?.div
  if (!pv || !textLayer) return
  const pr = pv.div.getBoundingClientRect()
  const k = pxPerUnit(pageNum, pr)

  const nodes: { node: Text; box: PdfRect }[] = []
  const walker = document.createTreeWalker(textLayer, NodeFilter.SHOW_TEXT)
  for (let n = walker.nextNode(); n; n = walker.nextNode()) {
    const el = n.parentElement
    if (!el || !n.textContent) continue
    const r = el.getBoundingClientRect()
    nodes.push({ node: n as Text, box: { x: (r.left - pr.left) / k, y: (r.top - pr.top) / k, w: r.width / k, h: r.height / k } })
  }
  const probe = document.createRange()
  const intersects = (a: PdfRect, b: PdfRect) => a.x < b.x + b.w && b.x < a.x + a.w && a.y < b.y + b.h && b.y < a.y + a.h

  for (const a of legacy) {
    migrationAttempted.add(a.anchorId)
    let first: { node: Text; i: number } | null = null
    let last: { node: Text; i: number } | null = null
    for (const { node, box } of nodes) {
      if (!a.rects.some(r => intersects(r, box))) continue
      for (let i = 0; i < node.length; i++) {
        probe.setStart(node, i); probe.setEnd(node, i + 1)
        const cr = probe.getBoundingClientRect()
        if (cr.width <= 0) continue
        const cx0 = (cr.left - pr.left) / k, cw = cr.width / k
        const cy = (cr.top + cr.height / 2 - pr.top) / k
        const hit = a.rects.some(r => {
          if (cy < r.y || cy > r.y + r.h) return false
          return Math.min(cx0 + cw, r.x + r.w) - Math.max(cx0, r.x) >= 0.3 * cw
        })
        if (!hit) continue
        if (!first) first = { node, i }
        last = { node, i }
      }
    }
    if (!first || !last) {
      console.warn(`[PdfViewer] could not re-anchor legacy highlight ${a.anchorId}; keeping stored rects`)
      continue
    }
    const range = document.createRange()
    range.setStart(first.node, first.i)
    range.setEnd(last.node, last.i + 1)
    const text = range.toString().replace(/\s+/g, ' ').trim()
    const rects = rectsFromRange(range, pv.div, pageNum)
    if (!text || !rects.length) continue
    const m: PdfAnchorMigration = { anchorId: a.anchorId, rects, text, v: PDF_ANCHOR_VERSION }
    migratedAnchors.set(a.anchorId, m)
    emit('anchor-migrated', m)
  }
  renderHighlights(pageNum)
}

// ─── Selection capture ────────────────────────────────────────────────────────

function onDocMouseDown(e: MouseEvent) {
  if ((e.target as HTMLElement | null)?.closest('.pdf-selection-tooltip')) return
  tooltip.value = null
}

function onSelectionChange() {
  const sel = window.getSelection()
  if (!sel || sel.isCollapsed) {
    tooltip.value = null
    return
  }
  // Touch/pen: selection handles keep moving after pointerup; capture once they settle.
  if (pendingTouchCapture) {
    if (touchSelectionTimer) clearTimeout(touchSelectionTimer)
    touchSelectionTimer = window.setTimeout(() => {
      pendingTouchCapture = false
      touchSelectionTimer = null
      captureSelection(lastPointerPos.clientX, lastPointerPos.clientY)
    }, 80)
  }
}

function onDocPointerDown(e: PointerEvent) {
  if (e.pointerType !== 'mouse') {
    if (touchSelectionTimer) { clearTimeout(touchSelectionTimer); touchSelectionTimer = null }
    pendingTouchCapture = false
  }
  // Scanned / image-only page: explain why nothing will select.
  if (props.enableSelection) {
    const page = pageOf(e.target as Node)
    const n = Number(page?.dataset.pageNumber ?? 0)
    if (n && noTextPages.has(n)) showHint('no-text', e.clientX, e.clientY)
  }
}

function onDocPointerMove(e: PointerEvent) {
  if (e.pointerType === 'mouse' && e.buttons & 1) snapFocusToLineEdge(e.clientX, e.clientY)
}

function onDocPointerUp(e: PointerEvent) {
  const { clientX, clientY } = e
  lastPointerPos = { clientX, clientY }
  if (e.pointerType === 'mouse') {
    if (e.button !== 0) return
    // Plain click (collapsed selection) on a highlight → anchor-click.
    if (window.getSelection()?.isCollapsed && props.anchors.length) {
      const id = anchorAt(clientX, clientY)
      if (id) emit('anchor-click', id)
    }
    snapFocusToLineEdge(clientX, clientY)
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

function onDocMouseMove(e: MouseEvent) {
  if (!props.anchors.length) return
  const found = anchorAt(e.clientX, e.clientY)
  if (found === hoveredAnchorId) return
  const toggle = (id: string, on: boolean) => {
    viewerEl.value?.querySelectorAll(`.pdf-highlight[data-anchor-id="${CSS.escape(id)}"]`)
      .forEach(el => el.classList.toggle('pdf-highlight-hover', on))
  }
  if (hoveredAnchorId) toggle(hoveredAnchorId, false)
  hoveredAnchorId = found
  if (found) toggle(found, true)
}

/**
 * Line-edge snap. Browsers don't extend a selection while the pointer is over
 * empty space (the margin past a line end, or before its start), so a fast
 * drag past the end of a line leaves the last few characters unselected —
 * whatever the last pointer sample over text happened to be. When the pointer
 * is beside the span holding the selection's focus (same line, past its edge),
 * move the focus to that edge.
 */
function snapFocusToLineEdge(clientX: number, clientY: number) {
  if (!props.enableSelection) return
  const sel = window.getSelection()
  if (!sel || sel.isCollapsed || !sel.focusNode) return
  const focusEl = sel.focusNode.nodeType === Node.TEXT_NODE ? sel.focusNode.parentElement : sel.focusNode as Element
  const span = focusEl?.closest('.textLayer span') as HTMLElement | null
  if (!span || !viewerEl.value?.contains(span)) return
  const text = Array.from(span.childNodes).find((n): n is Text => n.nodeType === Node.TEXT_NODE)
  if (!text) return
  const r = span.getBoundingClientRect()
  if (clientY < r.top || clientY > r.bottom) return
  try {
    if (clientX > r.right && !(sel.focusNode === text && sel.focusOffset === text.length)) {
      sel.extend(text, text.length)
    } else if (clientX < r.left && !(sel.focusNode === text && sel.focusOffset === 0)) {
      sel.extend(text, 0)
    }
  } catch { /* extend() can throw on detached nodes — safe to ignore */ }
}

function captureSelection(clientX = 0, clientY = 0) {
  if (!props.enableSelection || !root.value) return
  const sel = window.getSelection()
  if (!sel || sel.isCollapsed || sel.rangeCount === 0) return
  const range = sel.getRangeAt(0)
  if (root.value.querySelector('.pdf-selection-tooltip')?.contains(range.commonAncestorContainer)) return

  const startPage = pageOf(range.startContainer)
  const endPage = pageOf(range.endContainer)
  if (!startPage) return
  if (endPage && endPage !== startPage) {
    showHint('cross-page', clientX, clientY)
    return
  }
  const pageNum = Number(startPage.dataset.pageNumber ?? 0)
  if (!pageNum) return

  const rootRect = root.value.getBoundingClientRect()
  const pr = startPage.getBoundingClientRect()
  const k = pxPerUnit(pageNum, pr)
  const rects = rectsFromRange(range, startPage, pageNum)
  if (!rects.length) return
  const text = range.toString().replace(/\s+/g, ' ').trim()
  if (!text) return

  const first = rects[0]
  const offset = Math.max(32, first.h * k + 8)
  tooltip.value = {
    x: pr.left + first.x * k - rootRect.left + root.value.scrollLeft,
    y: pr.top + first.y * k - rootRect.top + root.value.scrollTop - offset,
    page: pageNum,
    rects,
    text
  }
}

function showHint(kind: 'cross-page' | 'no-text', clientX: number, clientY: number) {
  if (!root.value) return
  const rr = root.value.getBoundingClientRect()
  hint.value = { kind, x: clientX - rr.left + root.value.scrollLeft, y: clientY - rr.top + root.value.scrollTop - 40 }
  if (hintTimer) clearTimeout(hintTimer)
  hintTimer = window.setTimeout(() => { hint.value = null }, 2500)
}

function commitSelection(color: PdfQuoteColor) {
  if (!tooltip.value) return
  const { page, rects, text } = tooltip.value
  if (text) emit('select', { page, text, rects, color, v: PDF_ANCHOR_VERSION })
  tooltip.value = null
  window.getSelection()?.removeAllRanges()
}

// ─── External API ─────────────────────────────────────────────────────────────

function scrollToPage(p: number) {
  if (!viewer?.pdfDocument || p < 1 || p > viewer.pagesCount) return
  viewer.scrollPageIntoView({ pageNumber: p })
}

function scrollToAnchor(anchorId: string) {
  const anchor = props.anchors.find(a => a.anchorId === anchorId)
  const div = anchor && pageView(anchor.page)?.div
  if (!anchor || !div || !root.value) return
  const pr = div.getBoundingClientRect()
  const rr = root.value.getBoundingClientRect()
  const y = anchorRects(anchor)[0]?.y ?? 0
  const top = pr.top - rr.top + root.value.scrollTop + y * pxPerUnit(anchor.page, pr) - 80
  root.value.scrollTo({ top: Math.max(0, top), behavior: 'smooth' })
  window.setTimeout(() => {
    div.querySelectorAll(`.pdf-highlight[data-anchor-id="${CSS.escape(anchorId)}"]`).forEach(el => {
      el.classList.add('pdf-highlight-active')
      window.setTimeout(() => el.classList.remove('pdf-highlight-active'), 1500)
    })
  }, 50)
}

/** Kept for API compatibility: PDFViewer preserves the reading position across zoom. */
function setRestorePage(_p: number) { /* no-op */ }

/** Resolves after the next page render following a scale change (immediately if none pending). */
function rescaleDone(): Promise<void> {
  if (!scalePending) return Promise.resolve()
  return new Promise<void>(resolve => { rescaleResolvers.push(resolve) })
}

async function pageSizeAtScale1(): Promise<{ w: number; h: number }> {
  if (!pdfDoc) return { w: 0, h: 0 }
  const vp = (await pdfDoc.getPage(1)).getViewport({ scale: 1 })
  return { w: vp.width, h: vp.height }
}
async function getPageWidthAtScale1() { return (await pageSizeAtScale1()).w }
async function getPageHeightAtScale1() { return (await pageSizeAtScale1()).h }

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
.pdf-viewer-status {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.875rem;
}

.pdf-hint {
  position: absolute;
  transform: translateX(-50%);
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
