<!--
  Dev-only harness for the Playwright suite in client/e2e/.
  Mounted at /__e2e/pdf-viewer (registered only when import.meta.env.DEV).

  Query params: scale (default 1.5), selection=0 to disable selection.
  Exposes window.__pdf (control surface) for tests; see e2e/helpers.ts.
-->
<template>
  <div class="harness-root">
    <PdfViewer
      ref="viewer"
      :url="fixtureUrl"
      :anchors="anchors"
      :scale="scale"
      :enable-selection="enableSelection"
      @select="e => record('select', e)"
      @anchor-click="e => record('anchor-click', e)"
      @anchor-migrated="e => record('anchor-migrated', e)"
      @page-change="e => record('page-change', e)"
      @loaded="onLoaded"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import PdfViewer, { type PdfAnchor } from '@/components/PdfViewer.vue'
import fixtureUrl from '../../../e2e/fixtures/selection.pdf?url'

const params = new URLSearchParams(location.search)
const scale = ref(Number(params.get('scale') ?? 1.5))
const enableSelection = ref(params.get('selection') !== '0')
const anchors = ref<PdfAnchor[]>([])
const viewer = ref<InstanceType<typeof PdfViewer> | null>(null)
const events: { type: string; payload: unknown; t: number }[] = []
let loaded = false

function onLoaded(info: unknown) {
  record('loaded', info)
  loaded = true
}

function record(type: string, payload: unknown) {
  events.push({ type, payload: JSON.parse(JSON.stringify(payload ?? null)), t: performance.now() })
}

;(window as unknown as { __pdf: unknown }).__pdf = {
  events,
  isLoaded: () => loaded,
  setScale: (s: number) => { scale.value = s },
  setAnchors: (a: PdfAnchor[]) => { anchors.value = a },
  setSelection: (on: boolean) => { enableSelection.value = on },
  scrollToPage: (p: number) => viewer.value?.scrollToPage(p),
  scrollToAnchor: (id: string) => viewer.value?.scrollToAnchor(id),
  rescaleDone: () => viewer.value?.rescaleDone(),
  clearEvents: () => { events.length = 0 },
}
</script>

<style scoped>
.harness-root {
  position: fixed;
  inset: 0;
}
</style>
