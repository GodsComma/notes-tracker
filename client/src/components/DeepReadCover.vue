<template>
  <div class="dr-cover" :class="{ 'dr-cover-loading': loading, 'dr-cover-error': error }">
    <canvas ref="canvasEl" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import * as pdfjsLib from 'pdfjs-dist'
import workerUrl from 'pdfjs-dist/build/pdf.worker.min.mjs?url'

pdfjsLib.GlobalWorkerOptions.workerSrc = workerUrl

const props = defineProps<{ url: string }>()

const canvasEl = ref<HTMLCanvasElement | null>(null)
const loading = ref(true)
const error = ref(false)
let task: pdfjsLib.PDFDocumentLoadingTask | null = null
let cancelled = false

onMounted(async () => {
  if (!canvasEl.value) return
  try {
    task = pdfjsLib.getDocument({ url: props.url })
    const pdf = await task.promise
    if (cancelled) return
    const page = await pdf.getPage(1)
    if (cancelled) return
    const baseViewport = page.getViewport({ scale: 1 })
    const targetWidth = 220
    const scale = targetWidth / baseViewport.width
    const viewport = page.getViewport({ scale })
    const canvas = canvasEl.value
    canvas.width = viewport.width
    canvas.height = viewport.height
    const ctx = canvas.getContext('2d')
    if (!ctx) return
    await page.render({ canvasContext: ctx, viewport, canvas }).promise
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
})

onBeforeUnmount(() => {
  cancelled = true
  task?.destroy()
})
</script>

<style scoped>
.dr-cover {
  width: 100%;
  aspect-ratio: 3 / 4;
  background: #f1f5f9;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 0.5rem;
}
.dr-cover canvas {
  max-width: 100%;
  max-height: 100%;
  display: block;
}
.dr-cover-loading::after {
  content: '…';
  color: #94a3b8;
  font-size: 1.5rem;
}
.dr-cover-error::after {
  content: '📄';
  font-size: 2rem;
  opacity: 0.5;
}
</style>
