// Real-app smoke test: SourceView + Deep Read with the Go backend.
// Opt-in: needs the API on :3001 with a THROWAWAY data dir, e.g.
//   DATA_DIR=$(mktemp -d) JWT_SECRET=e2e go run -tags fts5 ./server &
//   E2E_APP=1 npx playwright test e2e/app-smoke.spec.ts --project=chromium-dpr1
import { test, expect, type Page, type APIRequestContext } from '@playwright/test'
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { line, rangeOf, dragChars, selectionText, norm, setViewerScope, glyphBox, boxToScreen, expectCovers, union } from './helpers'

test.skip(!process.env.E2E_APP, 'set E2E_APP=1 with a backend on :3001 (throwaway DATA_DIR)')

const here = dirname(fileURLToPath(import.meta.url))
const PDF = readFileSync(join(here, 'fixtures/selection.pdf'))
const LEGACY = (JSON.parse(readFileSync(join(here, 'fixtures/legacy-anchors.json'), 'utf8')) as { zoom: number; anchor: any }[])
  .find(l => l.zoom === 1.5)!.anchor
const PHRASE = 'labore et dolore magna'

async function api(req: APIRequestContext, method: string, path: string, data?: unknown) {
  const res = await req.fetch(`/api/v1${path}`, { method, data })
  expect(res.ok(), `${method} ${path} → ${res.status()} ${await res.text()}`).toBeTruthy()
  const t = await res.text()
  return t ? JSON.parse(t) : null
}

async function quotes(req: APIRequestContext, sourceId: number) {
  const note = await api(req, 'GET', `/sources/${sourceId}/note`)
  const doc = note?.body ? JSON.parse(note.body) : { content: [] }
  return (doc.content ?? []).filter((n: any) => n.type === 'pdfQuote').map((n: any) => n.attrs)
}

async function seed(page: Page) {
  const req = page.request
  await api(req, 'POST', '/auth/register', { email: `e2e-${Date.now()}-${Math.random().toString(36).slice(2, 7)}@example.test`, password: 'e2e-password-123' })
  const sem = await api(req, 'POST', '/semesters', { name: 'E2E', start_date: '2026-09-01', end_date: '2026-12-20' })
  const cls = await api(req, 'POST', `/semesters/${sem.id}/classes`, { name: 'E2E Class', color: '#6366f1' })
  const src = await api(req, 'POST', `/classes/${cls.id}/sources`, { title: 'E2E Source' })
  const up = await req.post(`/api/v1/sources/${src.id}/attachments`, {
    multipart: { file: { name: 'selection.pdf', mimeType: 'application/pdf', buffer: PDF } },
  })
  expect(up.ok(), `upload → ${up.status()}`).toBeTruthy()
  const att = await up.json()
  // A note holding one highlight captured by the pre-fix viewer (no `v`, shifted text).
  const body = JSON.stringify({ type: 'doc', content: [
    { type: 'pdfQuote', attrs: { ...LEGACY, attachmentID: att.id, anchorId: 'legacy-1' }, content: [{ type: 'paragraph' }] },
  ] })
  await api(req, 'PUT', `/sources/${src.id}/note`, { body })
  return { sourceId: src.id as number, attachmentId: att.id as number, classId: cls.id as number }
}

test('SourceView: legacy highlight is migrated and saved; new capture is v2 with the right text', async ({ page }) => {
  const errors: string[] = []
  page.on('pageerror', e => errors.push(String(e)))
  const { sourceId } = await seed(page)
  await page.goto(`/sources/${sourceId}`)
  await page.waitForSelector('[data-page-number="1"] .textLayer span', { timeout: 20_000 })

  // Migration persisted through the editor's autosave.
  await expect.poll(async () => (await quotes(page.request, sourceId))[0], { timeout: 15_000 })
    .toMatchObject({ anchorId: 'legacy-1', v: 2, text: PHRASE })

  // New capture.
  const L = line('p1-para2-L0')
  const [a, b] = rangeOf(L, L.text.split(' ').slice(0, 3).join(' '))
  await dragChars(page, { l: L, i: a }, { l: L, i: b })
  const selected = await selectionText(page)
  expect(selected).toBe(norm(L.text.slice(a, b)))
  await page.locator('.pdf-selection-tooltip .pdf-color-swatch[data-color="green"]').click()
  await expect.poll(async () => (await quotes(page.request, sourceId)).length, { timeout: 15_000 }).toBe(2)
  const q = (await quotes(page.request, sourceId)).find((x: any) => x.anchorId !== 'legacy-1')
  expect(q).toMatchObject({ v: 2, text: selected, color: 'green', page: 1 })

  // Both highlights are drawn.
  await expect(page.locator('.pdf-highlight[data-anchor-id="legacy-1"]').first()).toBeVisible()
  await expect(page.locator(`.pdf-highlight[data-anchor-id="${q.anchorId}"]`).first()).toBeVisible()
  await page.screenshot({ path: test.info().outputPath('source-view.png') })
  expect(errors).toEqual([])
})

test('Deep Read: renders, zoom controls work, selection captures', async ({ page }) => {
  const errors: string[] = []
  page.on('pageerror', e => errors.push(String(e)))
  const { classId } = await seed(page)
  // Navigate like a user (class page → source). A direct /sources/:id load
  // leaves SourceView's class unresolved, so Focus Mode is a no-op — a
  // pre-existing bug tracked separately (Spec/spec-pdf-text-selection.md §11).
  await page.goto(`/classes/${classId}`)
  await page.getByText('E2E Source').first().click()
  await page.waitForSelector('[data-page-number="1"] .textLayer span', { timeout: 20_000 })
  // "Focus Mode" starts the session directly (focusStore.requestPreflight).
  await page.getByRole('button', { name: 'Focus Mode' }).click()
  const dr = page.locator('.dr-pdf')
  await expect(dr).toBeVisible()
  await page.waitForSelector('.dr-pdf [data-page-number="1"] .textLayer span', { timeout: 20_000 })
  // The splash waits on PdfViewer's loaded + rescaleDone(); it must clear.
  await expect(page.locator('.dr-splash')).toHaveCount(0, { timeout: 10_000 })
  await page.screenshot({ path: test.info().outputPath('deep-read.png') })
  await setViewerScope(page, '.dr-pdf')

  const L = line('p1-para1-L1')
  const [a, b] = rangeOf(L, PHRASE)
  await dragChars(page, { l: L, i: a }, { l: L, i: b })
  expect(await selectionText(page)).toBe(PHRASE)
  await expect(page.locator('.dr-pdf .pdf-selection-tooltip')).toBeVisible()

  // Zoom with the real control: the migrated highlight must stay on its text
  // and selection must keep working at the new scale.
  await page.mouse.click(5, 5)
  const pct = page.locator('.dr-zoom-pct')
  const before = await pct.textContent()
  await page.getByTitle('Zoom out').click()
  await page.getByTitle('Zoom out').click()
  await expect(pct).not.toHaveText(before ?? '')
  await page.waitForTimeout(600)
  const exp = await boxToScreen(page, 1, glyphBox(L, a, b))
  const boxes = await page.evaluate(() => Array.from(document.querySelectorAll('.dr-pdf .pdf-highlight[data-anchor-id="legacy-1"]'))
    .map(el => { const r = el.getBoundingClientRect(); return { x: r.left, y: r.top, w: r.width, h: r.height } }))
  expectCovers(union(boxes), exp, 'highlight after Deep Read zoom')
  await dragChars(page, { l: L, i: a }, { l: L, i: b })
  expect(await selectionText(page)).toBe(PHRASE)
  await page.screenshot({ path: test.info().outputPath('deep-read-zoomed.png') })
  expect(errors).toEqual([])
})

test('Deep Read: resumes at the last-read page and it is selectable', async ({ page }) => {
  const { classId, attachmentId } = await seed(page)
  await api(page.request, 'PUT', `/attachments/${attachmentId}/progress`, { last_page: 40 })
  await page.goto(`/classes/${classId}`)
  await page.getByText('E2E Source').first().click()
  await page.waitForSelector('[data-page-number="1"] .textLayer span', { timeout: 20_000 })
  await page.getByRole('button', { name: 'Focus Mode' }).click()
  // Deep Read mounts asynchronously; only then is "splash gone" meaningful
  // (its leave transition still intercepts the pointer).
  await expect(page.locator('.dr-pdf')).toBeVisible()
  await expect(page.locator('.dr-splash')).toHaveCount(0, { timeout: 15_000 })
  await setViewerScope(page, '.dr-pdf')
  await page.waitForSelector('.dr-pdf [data-page-number="40"] .textLayer span', { timeout: 5_000 })
  // Page 40 is what the reader sees (its top is within the viewport).
  const top = await page.evaluate(() => document.querySelector('.dr-pdf [data-page-number="40"]')!.getBoundingClientRect().top)
  expect(top).toBeGreaterThan(-200)
  expect(top).toBeLessThan(400)
  const l = line('p40-marker')
  const [a, b] = rangeOf(l, 'marker line')
  await dragChars(page, { l, i: a }, { l, i: b })
  expect(await selectionText(page)).toBe('marker line')
})
