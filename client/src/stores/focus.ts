import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { PDFLibraryItem } from '@/api'
import type { DeepReadSettings } from '@/composables/useDeepRead'

// Global Pinia store for the Deep Read focus session.
// Lets any view (Home, Class, Source, Library) launch a session through
// a single mount point in AppShell, instead of inlining the focus reader.

export interface PreflightRequest {
  item: PDFLibraryItem
  returnTo?: string
}

export interface ActiveSession {
  item: PDFLibraryItem
  settings: DeepReadSettings
  returnTo?: string
}

export const useFocusStore = defineStore('focus', () => {
  const pendingPreflight = ref<PreflightRequest | null>(null)
  const active = ref<ActiveSession | null>(null)

  function requestPreflight(item: PDFLibraryItem, returnTo?: string) {
    pendingPreflight.value = { item, returnTo }
  }

  function cancelPreflight() {
    pendingPreflight.value = null
  }

  function startSession(settings: DeepReadSettings) {
    if (!pendingPreflight.value) return
    active.value = {
      item: pendingPreflight.value.item,
      settings,
      returnTo: pendingPreflight.value.returnTo
    }
    pendingPreflight.value = null
  }

  function endSession() {
    active.value = null
  }

  return {
    pendingPreflight,
    active,
    requestPreflight,
    cancelPreflight,
    startSession,
    endSession
  }
})
