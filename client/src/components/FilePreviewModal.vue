<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="attachment" class="fixed inset-0 z-50 flex flex-col bg-slate-900/95">
        <!-- Header -->
        <div class="shrink-0 flex items-center justify-between px-6 py-4 border-b border-slate-700">
          <div class="flex items-center gap-3 min-w-0">
            <span class="text-xl">{{ fileIcon(attachment.mime_type) }}</span>
            <span class="text-white font-medium truncate">{{ attachment.filename }}</span>
            <span class="text-slate-400 text-sm">{{ formatBytes(attachment.size_bytes) }}</span>
          </div>
          <div class="flex items-center gap-3 shrink-0">
            <a
              :href="api.attachments.fileUrl(attachment.id, 'attachment')"
              download
              class="flex items-center gap-1.5 text-sm text-slate-300 hover:text-white transition-colors"
            >
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              Download
            </a>
            <button class="text-slate-400 hover:text-white transition-colors" @click="$emit('close')">
              <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <!-- Preview body -->
        <div class="flex-1 overflow-hidden flex items-center justify-center p-4">
          <!-- Image -->
          <img
            v-if="attachment.mime_type.startsWith('image/')"
            :src="api.attachments.fileUrl(attachment.id)"
            :alt="attachment.filename"
            class="max-w-full max-h-full object-contain rounded-lg"
          />

          <!-- PDF -->
          <iframe
            v-else-if="attachment.mime_type === 'application/pdf'"
            :src="api.attachments.fileUrl(attachment.id)"
            class="w-full h-full rounded-lg bg-white"
            title="PDF preview"
          />

          <!-- No preview -->
          <div v-else class="text-center text-slate-400">
            <p class="text-5xl mb-4">{{ fileIcon(attachment.mime_type) }}</p>
            <p class="text-lg font-medium text-slate-200 mb-1">{{ attachment.filename }}</p>
            <p class="text-sm mb-6">No preview available for this file type.</p>
            <a
              :href="api.attachments.fileUrl(attachment.id, 'attachment')"
              download
              class="inline-flex items-center gap-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors"
            >
              Download file
            </a>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { api, type Attachment } from '@/api'
import { fileIcon, formatBytes } from '@/utils/mime'

defineProps<{ attachment: Attachment | null }>()
const emit = defineEmits<{ close: [] }>()

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
}
onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))
</script>

<style scoped>
.modal-enter-active, .modal-leave-active { transition: opacity 0.15s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
</style>
