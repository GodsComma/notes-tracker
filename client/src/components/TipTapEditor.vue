<template>
  <div
    ref="wrapper"
    class="tiptap-wrapper"
    :class="{ 'tiptap-locked': locked }"
    @click="onWrapperClick"
  >
    <EditorContent :editor="editor" />
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import Typography from '@tiptap/extension-typography'
import Image from '@tiptap/extension-image'
import Mathematics from '@tiptap/extension-mathematics'
import { PdfQuote, type PdfQuoteAttrs, type PdfQuoteColor, PDF_QUOTE_COLORS } from '@/extensions/PdfQuote'

const props = withDefaults(defineProps<{
  initialContent: string
  locked?: boolean
}>(), { locked: false })

const emit = defineEmits<{
  update: [content: string]
  'anchor-click': [anchorId: string]
}>()

const wrapper = ref<HTMLElement | null>(null)

const editor = useEditor({
  extensions: [
    StarterKit,
    Placeholder.configure({ placeholder: 'Start writing…' }),
    Typography,
    Image,
    Mathematics,
    PdfQuote
  ],
  content: props.initialContent ? JSON.parse(props.initialContent) : '',
  editable: !props.locked,
  editorProps: {
    attributes: { class: 'ProseMirror prose max-w-none focus:outline-none' }
  },
  onUpdate({ editor }) {
    emit('update', JSON.stringify(editor.getJSON()))
  }
})

watch(() => props.locked, (l) => {
  editor.value?.setEditable(!l)
})

watch(() => props.initialContent, (val) => {
  if (!editor.value) return
  const current = JSON.stringify(editor.value.getJSON())
  if (current !== val) {
    editor.value.commands.setContent(val ? JSON.parse(val) : '', false)
  }
})

onBeforeUnmount(() => editor.value?.destroy())

function onWrapperClick(e: MouseEvent) {
  const target = e.target as HTMLElement | null
  if (!target) return

  // ↗ Page link → external scroll
  const link = target.closest('[data-pdf-quote-link]') as HTMLElement | null
  if (link) {
    e.preventDefault(); e.stopPropagation()
    const id = link.getAttribute('data-anchor-id') ?? ''
    if (id) emit('anchor-click', id)
    return
  }

  // Recolor (suppressed when locked)
  const colorBtn = target.closest('[data-pdf-quote-color]') as HTMLElement | null
  if (colorBtn) {
    e.preventDefault(); e.stopPropagation()
    if (props.locked) return
    const id = colorBtn.getAttribute('data-anchor-id') ?? ''
    const color = colorBtn.getAttribute('data-pdf-quote-color') as PdfQuoteColor | null
    if (id && color && PDF_QUOTE_COLORS.includes(color)) {
      editor.value?.commands.setPdfQuoteColor(id, color)
    }
    return
  }

  // Delete sectional note (suppressed when locked)
  const delBtn = target.closest('[data-pdf-quote-delete]') as HTMLElement | null
  if (delBtn) {
    e.preventDefault(); e.stopPropagation()
    if (props.locked) return
    const id = delBtn.getAttribute('data-anchor-id') ?? ''
    if (id) editor.value?.commands.removePdfQuote(id)
    return
  }
}

function insertPdfQuote(attrs: PdfQuoteAttrs) {
  editor.value?.commands.insertPdfQuote(attrs)
  editor.value?.commands.focus()
}

function scrollToAnchor(anchorId: string) {
  const root = wrapper.value
  if (!root) return
  const el = root.querySelector(`[data-pdf-quote][data-anchor-id="${CSS.escape(anchorId)}"]`) as HTMLElement | null
  if (!el) return
  el.scrollIntoView({ behavior: 'smooth', block: 'center' })
  el.classList.add('pdf-quote-pulse')
  window.setTimeout(() => el.classList.remove('pdf-quote-pulse'), 1500)
}

/** Persist a re-anchored legacy highlight (PdfViewer `anchor-migrated`). */
function updatePdfQuoteAnchor(anchorId: string, patch: { rects: PdfQuoteAttrs['rects']; text: string; v: number }) {
  editor.value?.commands.updatePdfQuoteAnchor(anchorId, patch)
}

defineExpose({ insertPdfQuote, scrollToAnchor, updatePdfQuoteAnchor })
</script>

<style scoped>
.tiptap-wrapper {
  min-height: 24rem;
}
</style>
