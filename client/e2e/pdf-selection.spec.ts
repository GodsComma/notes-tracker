// PDF text selection suite — see Spec/spec-pdf-text-selection.md §4.
import { test, expect, type Page } from '@playwright/test'
import { readFileSync, existsSync, writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import {
  manifest, line, linesOnPage, glyphBox, rangeOf, openViewer, waitForTextLayer, setScale,
  dragChars, dragToPoint, selectionText, events, norm, union, expectCovers, boxToScreen, coverage,
  type Box,
} from './helpers'

const ZOOMS = [1, 1.5, 2, 3]
const here = dirname(fileURLToPath(import.meta.url))
const LEGACY_PATH = join(here, 'fixtures/legacy-anchors.json')

// The phrase most tests select: a full-word run in the middle of a body line.
const L1 = line('p1-para1-L1')
const PHRASE = 'labore et dolore magna'
const [P0, P1] = rangeOf(L1, PHRASE)

// ─── A. Text-layer alignment ─────────────────────────────────────────────────
for (const zoom of ZOOMS) {
  test(`A · text layer aligns with glyphs @${zoom}x`, async ({ page }) => {
    await openViewer(page, { scale: zoom })
    const samples: { word: string; pageNum: number; cx: number; cy: number; size: number }[] = []
    for (const pageNum of [1, 2, 3]) {
      for (const l of linesOnPage(pageNum)) {
        let idx = 0
        for (const word of l.text.split(' ')) {
          const b = glyphBox(l, idx, idx + word.length)
          samples.push({ word, pageNum, cx: b.x + b.w / 2, cy: b.y + b.h / 2, size: l.size })
          idx += word.length + 1
        }
      }
    }
    for (const p of [2, 3]) {
      await page.evaluate((n) => (window as any).__pdf.scrollToPage(n), p)
      await waitForTextLayer(page, p)
    }
    const r = await page.evaluate(({ samples, pw }) => {
      const scroller = document.querySelector('.pdf-viewer') as HTMLElement
      let hits = 0
      const sizeErr: number[] = []
      const misses: string[] = []
      for (const s of samples) {
        const canvas = document.querySelector(`[data-page-number="${s.pageNum}"] canvas`) as HTMLElement
        let cr = canvas.getBoundingClientRect()
        let k = cr.width / pw
        const sr = scroller.getBoundingClientRect()
        scroller.scrollTop += cr.top + s.cy * k - (sr.top + sr.height / 2)
        scroller.scrollLeft += cr.left + s.cx * k - (sr.left + sr.width / 2)
        cr = canvas.getBoundingClientRect(); k = cr.width / pw
        const el = document.elementFromPoint(cr.left + s.cx * k, cr.top + s.cy * k) as HTMLElement | null
        const span = el?.closest('.textLayer span') as HTMLElement | null
        if (span && span.textContent?.includes(s.word)) {
          hits++
          sizeErr.push(Math.abs(span.getBoundingClientRect().height / (s.size * k) - 1))
        } else if (misses.length < 5) {
          misses.push(`p${s.pageNum} "${s.word}" → ${el?.tagName}.${el?.className}`)
        }
      }
      sizeErr.sort((a, b) => a - b)
      return { n: samples.length, hitRate: hits / samples.length, medSizeErr: sizeErr[Math.floor(sizeErr.length / 2)] ?? 1, misses }
    }, { samples, pw: manifest.pageWidth })
    test.info().annotations.push({ type: 'metrics', description: JSON.stringify(r) })
    expect.soft(r.medSizeErr, `median span size error (misses: ${r.misses.join('; ')})`).toBeLessThan(0.02)
    expect(r.hitRate, `hit rate over ${r.n} words (misses: ${r.misses.join('; ')})`).toBeGreaterThanOrEqual(0.98)
  })
}

// ─── B. Drag-select a known phrase ───────────────────────────────────────────
for (const zoom of ZOOMS) {
  test(`B · drag-select phrase within a line @${zoom}x`, async ({ page }) => {
    await openViewer(page, { scale: zoom })
    await dragChars(page, { l: L1, i: P0 }, { l: L1, i: P1 })
    expect(await selectionText(page)).toBe(PHRASE)
    await expect(page.locator('.pdf-selection-tooltip')).toBeVisible()
  })

  test(`B · drag-select across two lines @${zoom}x`, async ({ page }) => {
    await openViewer(page, { scale: zoom })
    const L2 = line('p1-para1-L2')
    const [s0] = rangeOf(L1, 'veniam')
    const firstWords = L2.text.split(' ').slice(0, 3).join(' ')
    await dragChars(page, { l: L1, i: s0 }, { l: L2, i: firstWords.length })
    expect(await selectionText(page)).toBe(norm(`${L1.text.slice(s0)} ${firstWords}`))
  })
}

// ─── C. Drag ending in a gap or margin ───────────────────────────────────────
for (const zoom of ZOOMS) {
  test(`C · drag into right margin stays on the line @${zoom}x`, async ({ page }) => {
    await openViewer(page, { scale: zoom })
    // Start near the line end so the line end and the margin past it are both
    // on screen at every zoom (at 3x the page is wider than the viewport, and
    // dragging outside the window is not something a user can do).
    const [q0] = rangeOf(L1, 'quis')
    const b = glyphBox(L1, q0, q0 + 1)
    const lineEnd = glyphBox(L1, 0, L1.text.length)
    const marginX = Math.min(manifest.pageWidth - 10, lineEnd.x + lineEnd.w + 30)
    await dragToPoint(page, { l: L1, i: q0 }, 1, marginX, b.y + b.h * 0.5)
    const sel = await selectionText(page)
    const rest = norm(L1.text.slice(q0))
    expect(sel.startsWith(rest), `selection "${sel.slice(0, 120)}" should start with rest of line`).toBe(true)
    expect(sel.length, 'selection must not run past the next line').toBeLessThanOrEqual(rest.length + line('p1-para1-L2').text.length + 2)
    expect(sel).not.toContain('omegamarker')
  })

  test(`C · drag into paragraph gap does not jump @${zoom}x`, async ({ page }) => {
    await openViewer(page, { scale: zoom })
    const para1 = linesOnPage(1).filter(l => l.id?.startsWith('p1-para1'))
    const last = para1[para1.length - 1]
    const next = line('p1-para2-L0')
    const lb = glyphBox(last, 0, 1), nb = glyphBox(next, 0, 1)
    const gapY = (lb.y + lb.h + nb.y) / 2
    await dragToPoint(page, { l: L1, i: P0 }, 1, manifest.pageWidth / 2, gapY)
    const sel = await selectionText(page)
    expect(sel.startsWith(PHRASE), `selection "${sel.slice(0, 80)}…" should start at the drag origin`).toBe(true)
    expect(sel).not.toContain('omegamarker')
  })
}

// ─── D. Highlight round-trip ─────────────────────────────────────────────────
async function captureHighlight(page: Page) {
  await dragChars(page, { l: L1, i: P0 }, { l: L1, i: P1 })
  await page.locator('.pdf-color-swatch[data-color="yellow"]').click()
  const sel = await events(page, 'select')
  expect(sel.length, 'one select event').toBe(1)
  return sel[0] as { page: number; text: string; rects: Box[]; color: string; v?: number }
}

async function renderedHighlightBox(page: Page, anchorId: string): Promise<Box> {
  const boxes = await page.evaluate((id) => Array.from(document.querySelectorAll(`.pdf-highlight[data-anchor-id="${id}"]`))
    .map(el => { const r = el.getBoundingClientRect(); return { x: r.left, y: r.top, w: r.width, h: r.height } }), anchorId)
  expect(boxes.length, 'highlight elements rendered').toBeGreaterThan(0)
  return union(boxes)
}

for (const zoom of ZOOMS) {
  test(`D · highlight captured @${zoom}x matches the text at every zoom`, async ({ page }) => {
    await openViewer(page, { scale: zoom })
    const expected = glyphBox(L1, P0, P1)
    const sel = await captureHighlight(page)
    expect(norm(sel.text)).toBe(PHRASE)
    // w/h are page points at scale 1 (x,y,w,h). Compare in page space.
    expectCovers(union(sel.rects), expected, 'emitted rects')

    await page.evaluate((a) => (window as any).__pdf.setAnchors([a]),
      { anchorId: 'd1', page: sel.page, text: sel.text, rects: sel.rects, color: 'yellow', v: sel.v })
    for (const z of [zoom, ...ZOOMS.filter(z => z !== zoom)]) {
      if (z !== zoom) await setScale(page, z)
      const exp = await boxToScreen(page, 1, expected)
      const got = await renderedHighlightBox(page, 'd1')
      expectCovers(got, exp, `rendered highlight @${z}x`)
    }
  })
}

// ─── E. Select right after zooming ───────────────────────────────────────────
test('E · selection works right after a zoom change', async ({ page }) => {
  await openViewer(page, { scale: 1.5 })
  await setScale(page, 2)
  await waitForTextLayer(page, 1)
  const t0 = Date.now()
  await dragChars(page, { l: L1, i: P0 }, { l: L1, i: P1 })
  await expect(page.locator('.pdf-selection-tooltip')).toBeVisible({ timeout: 500 })
  expect(await selectionText(page)).toBe(PHRASE)
  expect(Date.now() - t0).toBeLessThan(3000)
  await expect(page.locator('.pdf-rescaling-banner')).toHaveCount(0)
})

// ─── F. Far page ─────────────────────────────────────────────────────────────
test('F · select on a far page right after jumping to it', async ({ page }) => {
  await openViewer(page, { scale: 1.5 })
  await page.evaluate(() => (window as any).__pdf.scrollToPage(150))
  await waitForTextLayer(page, 150, 1000)
  const l = line('p150-marker')
  const [a, b] = rangeOf(l, 'marker line')
  await dragChars(page, { l, i: a }, { l, i: b })
  expect(await selectionText(page)).toBe('marker line')
})

// ─── G. Cross-page drag ──────────────────────────────────────────────────────
test('G · cross-page drag is handled explicitly', async ({ page }) => {
  await openViewer(page, { scale: 1 })
  await page.evaluate(() => (window as any).__pdf.scrollToPage(6))
  await waitForTextLayer(page, 6)
  await waitForTextLayer(page, 7)
  const tail = linesOnPage(6).filter(l => l.id?.startsWith('p6-tail'))
  const from = tail[tail.length - 1]
  const to = line('p7-head-L0')
  await dragChars(page, { l: from, i: 0 }, { l: to, i: 5 })
  const hint = page.locator('.pdf-cross-page-hint')
  const swatch = page.locator('.pdf-color-swatch[data-color="yellow"]')
  await expect(hint.or(swatch).first()).toBeVisible()
  if (await swatch.isVisible()) {
    await swatch.click()
    const sel = await events(page, 'select')
    expect(new Set(sel.map((s: any) => s.page))).toEqual(new Set([6, 7]))
  }
})

// ─── H. Image-only page ──────────────────────────────────────────────────────
test('H · image-only page explains why nothing can be selected', async ({ page }) => {
  const errors: string[] = []
  page.on('pageerror', e => errors.push(String(e)))
  await openViewer(page, { scale: 1 })
  await page.evaluate(() => (window as any).__pdf.scrollToPage(5))
  await page.waitForTimeout(500)
  const c = await boxToScreen(page, 5, { x: 200, y: 300, w: 10, h: 10 })
  await page.mouse.click(c.x, c.y)
  await expect(page.locator('.pdf-no-text-hint')).toBeVisible()
  expect(errors).toEqual([])
})

// ─── I. Legacy anchor migration ──────────────────────────────────────────────
// legacy-anchors.json is captured from the pre-fix viewer (see CAPTURE_LEGACY below).
test('I · legacy highlights are re-anchored to the text', async ({ page }) => {
  test.skip(!existsSync(LEGACY_PATH), 'legacy-anchors.json not captured yet')
  const legacy = JSON.parse(readFileSync(LEGACY_PATH, 'utf8')) as { zoom: number; anchor: any }[]
  await openViewer(page, { scale: 1.5 })
  await page.evaluate((as) => (window as any).__pdf.setAnchors(as), legacy.map(l => l.anchor))
  const expected = glyphBox(L1, P0, P1)
  await expect.poll(async () => (await events(page, 'anchor-migrated')).length).toBe(legacy.length)
  // Legacy captures stored the WRONG text (the misaligned layer shifted it), but
  // their rects sit where the user actually dragged. Migration must recover the
  // phrase the user saw, not reproduce the stored text.
  const migrated = await events(page, 'anchor-migrated')
  for (const m of migrated) {
    expect.soft(norm(m.text), `migrated text ${m.anchorId}`).toBe(PHRASE)
    expectCovers(union(m.rects), expected, `migrated rects ${m.anchorId}`)
  }
  const exp = await boxToScreen(page, 1, expected)
  for (const l of legacy) expectCovers(await renderedHighlightBox(page, l.anchor.anchorId), exp, `rendered ${l.anchor.anchorId}`)
})

// One-off: run against the PRE-FIX viewer with CAPTURE_LEGACY=1 to record what
// old captures look like. Not part of the normal suite.
test('capture legacy anchors (CAPTURE_LEGACY=1 only)', async ({ page }, info) => {
  test.skip(!process.env.CAPTURE_LEGACY || info.project.name !== 'chromium-dpr2', 'capture mode only')
  const out: { zoom: number; anchor: any }[] = []
  for (const zoom of [1, 1.5, 2]) {
    await openViewer(page, { scale: zoom })
    const sel = await captureHighlight(page)
    out.push({ zoom, anchor: { anchorId: `legacy-${zoom}`, page: sel.page, text: sel.text, rects: sel.rects, color: 'yellow' } })
  }
  writeFileSync(LEGACY_PATH, JSON.stringify(out, null, 1))
  const expected = glyphBox(L1, P0, P1)
  for (const o of out) {
    const c = coverage(union(o.anchor.rects), expected)
    test.info().annotations.push({ type: `legacy@${o.zoom}`, description: JSON.stringify(c) })
  }
})

// ─── J. Selection disabled ───────────────────────────────────────────────────
test('J · no capture when selection is disabled', async ({ page }) => {
  await openViewer(page, { scale: 1.5, selection: false })
  await dragChars(page, { l: L1, i: P0 }, { l: L1, i: P1 })
  await page.waitForTimeout(300)
  await expect(page.locator('.pdf-selection-tooltip')).toHaveCount(0)
  expect(await events(page, 'select')).toEqual([])
})

// ─── K. Anchor click / hover ─────────────────────────────────────────────────
test('K · clicking and hovering a highlight', async ({ page }) => {
  await openViewer(page, { scale: 1.5 })
  const b = glyphBox(L1, P0, P1)
  await page.evaluate((a) => (window as any).__pdf.setAnchors([a]),
    { anchorId: 'k1', page: 1, text: PHRASE, rects: [b], color: 'green', v: 2 })
  const s = await boxToScreen(page, 1, b)
  await page.mouse.move(s.x + s.w / 2, s.y + s.h / 2)
  await expect(page.locator('.pdf-highlight[data-anchor-id="k1"]').first()).toHaveClass(/pdf-highlight-hover/)
  await page.mouse.click(s.x + s.w / 2, s.y + s.h / 2)
  expect(await events(page, 'anchor-click')).toEqual(['k1'])
})

// ─── L. Touch smoke ──────────────────────────────────────────────────────────
test('L · touch selection shows the tooltip after the debounce', async ({ page }) => {
  await openViewer(page, { scale: 1.5 })
  const s = await boxToScreen(page, 1, glyphBox(L1, P0, P1))
  await page.evaluate(({ phrase, x, y }) => {
    const spans = Array.from(document.querySelectorAll('[data-page-number="1"] .textLayer span'))
    const span = spans.find(sp => sp.textContent?.includes(phrase) && !sp.querySelector('span'))!
    const node = span.firstChild!
    const i = node.textContent!.indexOf(phrase)
    const r = document.createRange()
    r.setStart(node, i); r.setEnd(node, i + phrase.length)
    const sel = window.getSelection()!
    sel.removeAllRanges(); sel.addRange(r)
    document.dispatchEvent(new PointerEvent('pointerdown', { pointerType: 'touch', bubbles: true, clientX: x, clientY: y }))
    document.dispatchEvent(new PointerEvent('pointerup', { pointerType: 'touch', bubbles: true, clientX: x, clientY: y }))
  }, { phrase: PHRASE, x: s.x, y: s.y })
  await expect(page.locator('.pdf-selection-tooltip')).toBeVisible({ timeout: 1000 })
})
