/**
 * pdf.js viewer components (TextLayerBuilder, PDFViewer, EventBus, …).
 * './pdfjs' must be evaluated first: pdf_viewer.mjs destructures
 * globalThis.pdfjsLib when it loads. ES modules evaluate imports in source
 * order, so keep the side-effect import on the first line.
 */
import './pdfjs'
export * from 'pdfjs-dist/web/pdf_viewer.mjs'
