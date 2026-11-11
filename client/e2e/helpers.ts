import { expect, type Page } from '@playwright/test'
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

// ─── Fixture manifest ────────────────────────────────────────────────────────

export interface FixtureLine {
  page: number; text: string; x: number; y: number; size: number
  font: string; rotate: number; id: string | null
}
interface Manifest {
  pageWidth: number; pageHeight: number
  metrics: Record<string, { ascent: number; height: number }>
  widthTables: Record<string, Record<string, number>>
  lines: FixtureLine[]
}

const here = dirname(fileURLToPath(import.meta.url))
export const manifest: Manifest = JSON.parse(readFileSync(join(here, 'fixtures/selection.manifest.json'), 'utf8'))

export function line(id: string): FixtureLine {
  const l = manifest.lines.find(l => l.id === id)
  if (!l) throw new Error(`no fixture line ${id}`)
  return l
}

export function linesOnPage(page: number): FixtureLine[] {
  return manifest.lines.filter(l => l.page === page && !l.rotate)
}

export function textWidth(l: FixtureLine, s: string): number {
  const t = manifest.widthTables[l.font]
  let w = 0
  for (const c of s) w += t[c] ?? 0
  return w * l.size
}

/** Box in PDF points, top-left origin, for chars [start, end) of a line. */
export interface Box { x: number; y: number; w: number; h: number }
export function glyphBox(l: FixtureLine, start: number, end: number): Box {
  const m = manifest.metrics[l.font]
  return {
    x: l.x + textWidth(l, l.text.slice(0, start)),
    y: manifest.pageHeight - l.y - m.ascent * l.size,
    w: textWidth(l, l.text.slice(start, end)),
    h: m.height * l.size,
  }
}

/** Index range of `phrase` within a line (throws if absent). */
export function rangeOf(l: FixtureLine, phrase: string): [number, number] {
  const i = l.text.indexOf(phrase)
  if (i < 0) throw new Error(`"${phrase}" not in line ${l.id}`)
  return [i, i + phrase.length]
}

// ─── Page lifecycle ──────────────────────────────────────────────────────────

export async function openViewer(page: Page, opts: { scale?: number; selection?: boolean } = {}) {
  // No backend in e2e: make the auth probe fail fast so the router renders the harness.
  await page.route('**/api/**', r => r.fulfill({ status: 401, body: '{}' }))
  const q = new URLSearchParams({ scale: String(opts.scale ?? 1.5) })
  if (opts.selection === false) q.set('selection', '0')
  await page.goto(`/__e2e/pdf-viewer?${q}`)
  await page.waitForFunction(() => (window as any).__pdf?.isLoaded?.(), null, { timeout: 30_000 })
  await waitForTextLayer(page, 1)
}

export async function waitForTextLayer(page: Page, pageNum: number, timeout = 10_000) {
  await page.waitForFunction((n) => {
    const el = document.querySelector(`[data-page-number="${n}"] .textLayer`)
    return !!el && el.querySelectorAll('span:not(.markedContent)').length > 0
  }, pageNum, { timeout })
}

export async function setScale(page: Page, scale: number) {
  await page.evaluate(async (s) => {
    const pdf = (window as any).__pdf
    pdf.setScale(s)
    await new Promise(r => setTimeout(r, 0))
    await pdf.rescaleDone()
  }, scale)
}

// ─── Geometry: PDF points → screen ───────────────────────────────────────────

/**
 * Scrolls the viewer so (ptX, ptY) on `pageNum` is in view, then returns its
 * client coordinates. Uses the page canvas as the visual source of truth.
 */
export async function toScreen(page: Page, pageNum: number, ptX: number, ptY: number) {
  return page.evaluate(({ pageNum, ptX, ptY, pw }) => {
    const scroller = document.querySelector(`${(window as any).__pdfScope ?? ''} .pdf-viewer`) as HTMLElement
    const canvas = () => document.querySelector(`${(window as any).__pdfScope ?? ''} [data-page-number="${pageNum}"] canvas`) as HTMLElement
    const at = () => {
      const r = canvas().getBoundingClientRect()
      const k = r.width / pw
      return { x: r.left + ptX * k, y: r.top + ptY * k, k }
    }
    let p = at()
    const sr = scroller.getBoundingClientRect()
    scroller.scrollTop += p.y - (sr.top + sr.height / 2)
    scroller.scrollLeft += p.x - (sr.left + sr.width / 2)
    p = at()
    return p
  }, { pageNum, ptX, ptY, pw: manifest.pageWidth })
}

/** Scroll so a box is visible; return its screen box. */
export async function boxToScreen(page: Page, pageNum: number, b: Box) {
  const c = await toScreen(page, pageNum, b.x + b.w / 2, b.y + b.h / 2)
  return { x: c.x - (b.w / 2) * c.k, y: c.y - (b.h / 2) * c.k, w: b.w * c.k, h: b.h * c.k, k: c.k }
}

/** Restrict geometry helpers to one viewer (e.g. '.dr-pdf') when several are mounted. */
export async function setViewerScope(page: Page, selector: string) {
  await page.evaluate((s) => { (window as any).__pdfScope = s }, selector)
}

// ─── Selection ───────────────────────────────────────────────────────────────

/** Drag from just inside the first char to just inside the last char of [start,end). */
export async function dragChars(page: Page, from: { l: FixtureLine; i: number }, to: { l: FixtureLine; i: number }) {
  const startBox = glyphBox(from.l, from.i, from.i + 1)
  const endBox = glyphBox(to.l, to.i - 1, to.i)
  const a = await pointFor(page, from.l, startBox.x + startBox.w * 0.25)
  await page.mouse.move(a.x, a.y)
  await page.mouse.down()
  const b = await pointFor(page, to.l, endBox.x + endBox.w * 0.75, false)
  const steps = 12
  for (let s = 1; s <= steps; s++) {
    await page.mouse.move(a.x + ((b.x - a.x) * s) / steps, a.y + ((b.y - a.y) * s) / steps)
  }
  await page.mouse.up()
}

/** Drag from a char to an arbitrary page point (for gap/margin tests). */
export async function dragToPoint(page: Page, from: { l: FixtureLine; i: number }, pageNum: number, ptX: number, ptY: number) {
  const startBox = glyphBox(from.l, from.i, from.i + 1)
  const a = await pointFor(page, from.l, startBox.x + startBox.w * 0.25)
  await page.mouse.move(a.x, a.y)
  await page.mouse.down()
  const b = await currentScreen(page, pageNum, ptX, ptY)
  const steps = 12
  for (let s = 1; s <= steps; s++) {
    await page.mouse.move(a.x + ((b.x - a.x) * s) / steps, a.y + ((b.y - a.y) * s) / steps)
  }
  await page.mouse.up()
}

/** Screen point at x (pt) on a line's x-height middle. Scrolls only when `scroll`. */
async function pointFor(page: Page, l: FixtureLine, ptX: number, scroll = true) {
  const m = manifest.metrics[l.font]
  const ptY = manifest.pageHeight - l.y - m.ascent * l.size * 0.45
  return scroll ? toScreen(page, l.page, ptX, ptY) : currentScreen(page, l.page, ptX, ptY)
}

/** Client coords of a page point without scrolling (mid-drag). */
async function currentScreen(page: Page, pageNum: number, ptX: number, ptY: number) {
  return page.evaluate(({ pageNum, ptX, ptY, pw }) => {
    const r = (document.querySelector(`${(window as any).__pdfScope ?? ''} [data-page-number="${pageNum}"] canvas`) as HTMLElement).getBoundingClientRect()
    const k = r.width / pw
    return { x: r.left + ptX * k, y: r.top + ptY * k, k }
  }, { pageNum, ptX, ptY, pw: manifest.pageWidth })
}

export const norm = (s: string) => s.replace(/\s+/g, ' ').trim()

export async function selectionText(page: Page) {
  return norm(await page.evaluate(() => window.getSelection()?.toString() ?? ''))
}

export async function events(page: Page, type: string) {
  return page.evaluate((t) => (window as any).__pdf.events.filter((e: any) => e.type === t).map((e: any) => e.payload), type)
}

// ─── Box comparison ──────────────────────────────────────────────────────────

export function union(bs: Box[]): Box {
  const x0 = Math.min(...bs.map(b => b.x)), y0 = Math.min(...bs.map(b => b.y))
  const x1 = Math.max(...bs.map(b => b.x + b.w)), y1 = Math.max(...bs.map(b => b.y + b.h))
  return { x: x0, y: y0, w: x1 - x0, h: y1 - y0 }
}

/** Fraction of `expected` covered horizontally / vertically by `actual`, and overshoot. */
export function coverage(actual: Box, expected: Box) {
  const ox = Math.max(0, Math.min(actual.x + actual.w, expected.x + expected.w) - Math.max(actual.x, expected.x))
  const oy = Math.max(0, Math.min(actual.y + actual.h, expected.y + expected.h) - Math.max(actual.y, expected.y))
  return {
    x: ox / expected.w,
    y: oy / expected.h,
    overshootX: Math.max(0, actual.w - ox) / expected.w,
  }
}

export function expectCovers(actual: Box, expected: Box, label: string) {
  const c = coverage(actual, expected)
  expect.soft(c.x, `${label}: horizontal coverage`).toBeGreaterThanOrEqual(0.95)
  expect.soft(c.overshootX, `${label}: horizontal overshoot`).toBeLessThanOrEqual(0.1)
  expect.soft(c.y, `${label}: vertical coverage`).toBeGreaterThanOrEqual(0.7)
}
