/**
 * Module-level PDF document cache. Keeps a PDFDocumentProxy alive across
 * SourceView → FocusMode navigation so PdfViewer doesn't re-parse the file.
 *
 * Lifecycle: each consumer calls acquireCachedPdf / releaseCachedPdf.
 * When the ref-count drops to zero the document is destroyed and evicted.
 */
import * as pdfjsLib from 'pdfjs-dist'
import workerUrl from 'pdfjs-dist/build/pdf.worker.min.mjs?url'

pdfjsLib.GlobalWorkerOptions.workerSrc = workerUrl

interface CacheEntry {
  doc: pdfjsLib.PDFDocumentProxy
  refs: number
}

const cache = new Map<string, CacheEntry>()

/** Start loading a PDF in the background. No-op if already cached. */
export function preloadPdf(url: string): void {
  if (cache.has(url)) return
  pdfjsLib.getDocument({ url }).promise.then(doc => {
    // Only store if nobody called preloadPdf for a different URL meanwhile
    if (!cache.has(url)) cache.set(url, { doc, refs: 0 })
  }).catch(() => { /* preload is best-effort */ })
}

/**
 * Returns the cached PDFDocumentProxy for `url` and increments its ref-count,
 * or null if it isn't cached yet (PdfViewer must call getDocument itself).
 */
export function acquireCachedPdf(url: string): pdfjsLib.PDFDocumentProxy | null {
  const entry = cache.get(url)
  if (!entry) return null
  entry.refs++
  return entry.doc
}

/**
 * Decrement ref-count for `url`. Destroys and evicts when count reaches 0.
 * Call this from onBeforeUnmount in any component that called acquireCachedPdf.
 */
export function releaseCachedPdf(url: string): void {
  const entry = cache.get(url)
  if (!entry) return
  entry.refs--
  if (entry.refs <= 0) {
    entry.doc.destroy()
    cache.delete(url)
  }
}
