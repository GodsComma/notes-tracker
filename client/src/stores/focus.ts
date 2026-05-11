import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type PDFLibraryItem, type FocusSettingsAPI } from '@/api'
import { DEFAULT_SETTINGS, type DeepReadSettings } from '@/composables/useDeepRead'

export interface PreflightRequest {
  item: PDFLibraryItem
  returnTo?: string
  initialSettings: DeepReadSettings
}

export interface ActiveSession {
  item: PDFLibraryItem
  settings: DeepReadSettings
  returnTo?: string
}

function apiToSettings(s: FocusSettingsAPI): DeepReadSettings {
  return {
    durationMin: s.duration_min,
    theme: s.theme as DeepReadSettings['theme'],
    ruler: s.ruler as DeepReadSettings['ruler'],
    suppressNotifications: s.suppress_notifications,
    annotations: s.annotations as 'on' | 'off',
    autoHideDrawer: s.auto_hide_drawer
  }
}

export function settingsToAPI(s: DeepReadSettings): FocusSettingsAPI {
  return {
    duration_min: s.durationMin,
    theme: s.theme,
    ruler: s.ruler,
    suppress_notifications: s.suppressNotifications,
    annotations: s.annotations,
    auto_hide_drawer: s.autoHideDrawer
  }
}

async function fetchSettings(item: PDFLibraryItem): Promise<DeepReadSettings> {
  try {
    const raw = await api.attachments.getFocusSettings(item.id)
    const s = apiToSettings(raw)
    if (item.source_id === null) s.annotations = 'off'
    return s
  } catch {
    const s = { ...DEFAULT_SETTINGS }
    if (item.source_id === null) s.annotations = 'off'
    return s
  }
}

export const useFocusStore = defineStore('focus', () => {
  const pendingSettings = ref<PreflightRequest | null>(null)
  const active = ref<ActiveSession | null>(null)

  // Immediately starts a focus session using the attachment's saved settings.
  async function requestPreflight(item: PDFLibraryItem, returnTo?: string) {
    const settings = await fetchSettings(item)
    active.value = { item, settings, returnTo }
  }

  // Opens the settings modal so the user can adjust preferences before (or without) starting.
  async function openSettings(item: PDFLibraryItem, returnTo?: string) {
    const initialSettings = await fetchSettings(item)
    pendingSettings.value = { item, returnTo, initialSettings }
  }

  function closeSettings() {
    pendingSettings.value = null
  }

  function startSession(item: PDFLibraryItem, settings: DeepReadSettings, returnTo?: string) {
    active.value = { item, settings, returnTo }
  }

  function endSession() {
    active.value = null
  }

  return {
    pendingSettings,
    active,
    requestPreflight,
    openSettings,
    closeSettings,
    startSession,
    endSession
  }
})
