import { Node, mergeAttributes } from '@tiptap/core'

export interface PdfRect {
  x: number
  y: number
  w: number
  h: number
}

export type PdfQuoteColor = 'yellow' | 'green' | 'blue' | 'pink'

export const PDF_QUOTE_COLORS: PdfQuoteColor[] = ['yellow', 'green', 'blue', 'pink']

export interface PdfQuoteAttrs {
  attachmentID: number
  page: number
  text: string
  anchorId: string
  rects: PdfRect[]
  color: PdfQuoteColor
}

declare module '@tiptap/core' {
  interface Commands<ReturnType> {
    pdfQuote: {
      insertPdfQuote: (attrs: PdfQuoteAttrs) => ReturnType
      setPdfQuoteColor: (anchorId: string, color: PdfQuoteColor) => ReturnType
      removePdfQuote: (anchorId: string) => ReturnType
    }
  }
}

export const PdfQuote = Node.create({
  name: 'pdfQuote',
  group: 'block',
  content: 'paragraph+',
  defining: true,
  selectable: true,
  draggable: false,
  isolating: true,

  addAttributes() {
    return {
      attachmentID: {
        default: null as number | null,
        parseHTML: el => {
          const v = (el as HTMLElement).getAttribute('data-attachment-id')
          return v ? Number(v) : null
        },
        renderHTML: (attrs: Record<string, unknown>) => ({ 'data-attachment-id': String(attrs.attachmentID ?? '') })
      },
      page: {
        default: 1,
        parseHTML: el => Number((el as HTMLElement).getAttribute('data-page') ?? 1),
        renderHTML: (attrs: Record<string, unknown>) => ({ 'data-page': String(attrs.page ?? 1) })
      },
      text: {
        default: '',
        parseHTML: el => (el as HTMLElement).querySelector('blockquote')?.textContent ?? '',
        renderHTML: (attrs: Record<string, unknown>) => ({ 'data-text': String(attrs.text ?? '') })
      },
      anchorId: {
        default: '',
        parseHTML: el => (el as HTMLElement).getAttribute('data-anchor-id') ?? '',
        renderHTML: (attrs: Record<string, unknown>) => ({ 'data-anchor-id': String(attrs.anchorId ?? '') })
      },
      rects: {
        default: [] as PdfRect[],
        parseHTML: el => {
          const raw = (el as HTMLElement).getAttribute('data-rects')
          if (!raw) return []
          try { return JSON.parse(raw) as PdfRect[] } catch { return [] }
        },
        renderHTML: (attrs: Record<string, unknown>) => ({
          'data-rects': JSON.stringify((attrs.rects as PdfRect[] | undefined) ?? [])
        })
      },
      color: {
        default: 'yellow' as PdfQuoteColor,
        parseHTML: el => {
          const raw = (el as HTMLElement).getAttribute('data-color') ?? 'yellow'
          return (PDF_QUOTE_COLORS.includes(raw as PdfQuoteColor) ? raw : 'yellow') as PdfQuoteColor
        },
        renderHTML: (attrs: Record<string, unknown>) => ({ 'data-color': String(attrs.color ?? 'yellow') })
      }
    }
  },

  parseHTML() {
    return [{
      tag: 'div[data-pdf-quote]',
      // Editable content area is everything inside .pdf-quote-commentary
      contentElement: el => (el as HTMLElement).querySelector('.pdf-quote-commentary') as HTMLElement ?? (el as HTMLElement)
    }]
  },

  renderHTML({ HTMLAttributes, node }) {
    const attrs = node.attrs as PdfQuoteAttrs
    const colorButtons = PDF_QUOTE_COLORS.map(c => [
      'button',
      {
        class: 'pdf-quote-color-btn' + (c === attrs.color ? ' pdf-quote-color-btn-active' : ''),
        type: 'button',
        'data-pdf-quote-color': c,
        'data-anchor-id': attrs.anchorId,
        'data-color': c,
        title: `Recolor (${c})`
      }
    ])
    return [
      'div',
      mergeAttributes(HTMLAttributes, { 'data-pdf-quote': '', class: 'pdf-quote-block' }),
      [
        'div',
        { class: 'pdf-quote-header', contenteditable: 'false' },
        ['blockquote', { class: 'pdf-quote-text' }, attrs.text],
        [
          'div',
          { class: 'pdf-quote-toolbar' },
          [
            'button',
            {
              class: 'pdf-quote-link',
              type: 'button',
              'data-pdf-quote-link': '',
              'data-anchor-id': attrs.anchorId
            },
            `↗ Page ${attrs.page}`
          ],
          ['div', { class: 'pdf-quote-actions' },
            ...colorButtons,
            [
              'button',
              {
                class: 'pdf-quote-delete',
                type: 'button',
                'data-pdf-quote-delete': '',
                'data-anchor-id': attrs.anchorId,
                title: 'Delete sectional note'
              },
              '×'
            ]
          ]
        ]
      ],
      ['div', { class: 'pdf-quote-commentary' }, 0]
    ]
  },

  addCommands() {
    return {
      insertPdfQuote:
        (attrs: PdfQuoteAttrs) =>
        ({ commands }) =>
          commands.insertContent({
            type: this.name,
            attrs,
            content: [{ type: 'paragraph' }]
          }),

      setPdfQuoteColor:
        (anchorId: string, color: PdfQuoteColor) =>
        ({ tr, state, dispatch }) => {
          let updated = false
          state.doc.descendants((node, pos) => {
            if (updated) return false
            if (node.type.name === 'pdfQuote' && node.attrs.anchorId === anchorId) {
              if (dispatch) tr.setNodeMarkup(pos, undefined, { ...node.attrs, color })
              updated = true
              return false
            }
            return true
          })
          return updated
        },

      removePdfQuote:
        (anchorId: string) =>
        ({ tr, state, dispatch }) => {
          let removed = false
          state.doc.descendants((node, pos) => {
            if (removed) return false
            if (node.type.name === 'pdfQuote' && node.attrs.anchorId === anchorId) {
              if (dispatch) tr.delete(pos, pos + node.nodeSize)
              removed = true
              return false
            }
            return true
          })
          return removed
        }
    }
  }
})
