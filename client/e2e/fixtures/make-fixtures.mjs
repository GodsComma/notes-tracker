// Generates e2e/fixtures/selection.pdf and selection.manifest.json.
//
// The manifest records every drawn line (page, text, x, baseline y, font,
// size) in PDF points with a bottom-left origin, so tests can compute the
// exact on-screen box of any substring at any zoom without trusting the
// viewer's own text layer.
//
// Run: node e2e/fixtures/make-fixtures.mjs   (from client/)

import { PDFDocument, StandardFonts, degrees, rgb } from 'pdf-lib'
import { writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const OUT_DIR = dirname(fileURLToPath(import.meta.url))
const W = 612, H = 792, MARGIN = 72

const WORDS = ('lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor ' +
  'incididunt ut labore et dolore magna aliqua enim ad minim veniam quis nostrud exercitation ' +
  'ullamco laboris nisi aliquip ex ea commodo consequat duis aute irure in reprehenderit voluptate ' +
  'velit esse cillum fugiat nulla pariatur excepteur sint occaecat cupidatat non proident').split(' ')

let wordCursor = 0
function nextWords(n) {
  const out = []
  for (let i = 0; i < n; i++) out.push(WORDS[wordCursor++ % WORDS.length])
  return out
}

const doc = await PDFDocument.create()
const fonts = {
  Helvetica: await doc.embedFont(StandardFonts.Helvetica),
  'Helvetica-Bold': await doc.embedFont(StandardFonts.HelveticaBold),
  'Times-Roman': await doc.embedFont(StandardFonts.TimesRoman),
}
const lines = []

function draw(page, pageNum, text, x, y, size, font = 'Helvetica', extra = {}) {
  page.drawText(text, { x, y, size, font: fonts[font], color: rgb(0, 0, 0), ...(extra.rotate ? { rotate: degrees(extra.rotate) } : {}) })
  lines.push({ page: pageNum, text, x, y, size, font, rotate: extra.rotate ?? 0, id: extra.id ?? null })
}

// Wrap `words` into lines no wider than maxW; optional `marker` is injected as
// the first word of the paragraph so tests can find it.
function paragraph(page, pageNum, { x, y, maxW, size = 11, leading = 15, font = 'Helvetica', words, idPrefix }) {
  const f = fonts[font]
  let line = []
  let n = 0
  for (const w of words) {
    const trial = [...line, w].join(' ')
    if (f.widthOfTextAtSize(trial, size) > maxW && line.length) {
      draw(page, pageNum, line.join(' '), x, y, size, font, { id: `${idPrefix}-L${n++}` })
      y -= leading
      line = [w]
    } else {
      line.push(w)
    }
  }
  if (line.length) { draw(page, pageNum, line.join(' '), x, y, size, font, { id: `${idPrefix}-L${n++}` }); y -= leading }
  return y
}

// ── Page 1: heading + three body paragraphs ────────────────────────────────
{
  const p = doc.addPage([W, H]); const pn = 1
  draw(p, pn, 'Selection Fixture Heading', MARGIN, 700, 24, 'Helvetica-Bold', { id: 'p1-heading' })
  let y = 660
  y = paragraph(p, pn, { x: MARGIN, y, maxW: W - 2 * MARGIN, words: ['alphamarker', ...nextWords(70)], idPrefix: 'p1-para1' }) - 12
  y = paragraph(p, pn, { x: MARGIN, y, maxW: W - 2 * MARGIN, words: ['betamarker', ...nextWords(70)], idPrefix: 'p1-para2' }) - 12
  paragraph(p, pn, { x: MARGIN, y, maxW: W - 2 * MARGIN, words: ['omegamarker', ...nextWords(70)], idPrefix: 'p1-para3' })
}

// ── Page 2: two columns ────────────────────────────────────────────────────
{
  const p = doc.addPage([W, H]); const pn = 2
  const colW = (W - 2 * MARGIN - 36) / 2
  paragraph(p, pn, { x: MARGIN, y: 700, maxW: colW, words: ['leftcolumn', ...nextWords(90)], idPrefix: 'p2-left' })
  paragraph(p, pn, { x: MARGIN + colW + 36, y: 700, maxW: colW, words: ['rightcolumn', ...nextWords(90)], idPrefix: 'p2-right' })
}

// ── Page 3: small text + superscript ───────────────────────────────────────
{
  const p = doc.addPage([W, H]); const pn = 3
  draw(p, pn, 'Body sentence with a footnote reference', MARGIN, 700, 11, 'Helvetica', { id: 'p3-body' })
  const supX = MARGIN + fonts.Helvetica.widthOfTextAtSize('Body sentence with a footnote reference', 11) + 1
  draw(p, pn, '1', supX, 704, 7, 'Helvetica', { id: 'p3-sup' })
  paragraph(p, pn, { x: MARGIN, y: 120, maxW: W - 2 * MARGIN, size: 7, leading: 9, font: 'Times-Roman',
    words: ['footnotemarker', ...nextWords(60)], idPrefix: 'p3-foot' })
}

// ── Page 4: rotated line + normal line ─────────────────────────────────────
{
  const p = doc.addPage([W, H]); const pn = 4
  draw(p, pn, 'Normal line on the rotated page', MARGIN, 700, 11, 'Helvetica', { id: 'p4-normal' })
  draw(p, pn, 'Rotated ninety degree line', 300, 200, 14, 'Helvetica', { id: 'p4-rotated', rotate: 90 })
}

// ── Page 5: image-only (vector shapes, no text at all) ─────────────────────
{
  const p = doc.addPage([W, H])
  for (let i = 0; i < 12; i++) {
    p.drawRectangle({ x: MARGIN, y: 700 - i * 50, width: 300 + (i % 3) * 60, height: 30, color: rgb(0.85 - i * 0.04, 0.85, 0.9) })
  }
}

// ── Pages 6–7: paragraph continuing across the page break ──────────────────
{
  const p6 = doc.addPage([W, H])
  paragraph(p6, 6, { x: MARGIN, y: 130, maxW: W - 2 * MARGIN, words: ['crosspagestart', ...nextWords(40)], idPrefix: 'p6-tail' })
  const p7 = doc.addPage([W, H])
  paragraph(p7, 7, { x: MARGIN, y: 720, maxW: W - 2 * MARGIN, words: [...nextWords(40), 'crosspageend'], idPrefix: 'p7-head' })
}

// ── Pages 8–200: filler with a unique marker line each ─────────────────────
for (let pn = 8; pn <= 200; pn++) {
  const p = doc.addPage([W, H])
  draw(p, pn, `Page ${pn} marker line for far selection`, MARGIN, 700, 11, 'Helvetica', { id: `p${pn}-marker` })
  paragraph(p, pn, { x: MARGIN, y: 670, maxW: W - 2 * MARGIN, words: nextWords(120), idPrefix: `p${pn}-fill` })
}

// Font metrics the tests need (per 1pt of size), from pdf-lib's AFM data.
const metrics = Object.fromEntries(Object.entries(fonts).map(([name, f]) => [name, {
  ascent: f.heightAtSize(1, { descender: false }),
  height: f.heightAtSize(1),
}]))

// Character advance widths so tests can compute substring offsets exactly.
const widthTables = {}
for (const [name, f] of Object.entries(fonts)) {
  const chars = new Set(lines.filter(l => l.font === name).flatMap(l => [...l.text]))
  widthTables[name] = Object.fromEntries([...chars].map(c => [c, f.widthOfTextAtSize(c, 1)]))
}

doc.setTitle('Selection fixture')
writeFileSync(join(OUT_DIR, 'selection.pdf'), await doc.save({ useObjectStreams: false }))
writeFileSync(join(OUT_DIR, 'selection.manifest.json'), JSON.stringify({ pageWidth: W, pageHeight: H, metrics, widthTables, lines }, null, 1))
console.log(`wrote selection.pdf (${doc.getPageCount()} pages, ${lines.length} lines)`)
