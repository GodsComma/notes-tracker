/**
 * Single entry point for pdf.js core. Sets the worker once and publishes the
 * library on globalThis, which pdfjs-dist/web/pdf_viewer.mjs (4.x) reads at
 * module-evaluation time. Import pdf.js from here, never from 'pdfjs-dist'.
 */
import * as pdfjsLib from 'pdfjs-dist'
import workerUrl from 'pdfjs-dist/build/pdf.worker.min.mjs?url'

pdfjsLib.GlobalWorkerOptions.workerSrc = workerUrl
;(globalThis as unknown as { pdfjsLib: typeof pdfjsLib }).pdfjsLib = pdfjsLib

export { pdfjsLib }
