// Shared types & local-storage helpers for the Deep Read focus module.

export type DeepReadTheme = 'default' | 'sepia' | 'dark' | 'contrast'
export type ReadingRulerMode = 'off' | 'underline' | 'masking'

export interface DeepReadSettings {
  durationMin: number       // 0 = no timer (open-ended)
  theme: DeepReadTheme
  ruler: ReadingRulerMode
  suppressNotifications: boolean
  annotations: 'on' | 'off' // Forced 'off' for class-level PDFs (no source)
  autoHideDrawer: boolean
}

export const DEFAULT_SETTINGS: DeepReadSettings = {
  durationMin: 25,
  theme: 'default',
  ruler: 'off',
  suppressNotifications: true,
  annotations: 'on',
  autoHideDrawer: true
}

export type ZoomMode = 'absolute' | 'fit-width' | 'fit-page'

export interface DeepReadProgress {
  lastPage: number
  totalPages: number
  zoom: number
  zoomMode?: ZoomMode
  timestamp: number
}

const SETTINGS_KEY = 'deepread:settings'
const PROGRESS_KEY = (id: number) => `deepread:progress:${id}`
const SESSIONS_KEY = 'deepread:sessions'

export interface FocusSession {
  date: string          // 'YYYY-MM-DD' local time
  durationSec: number   // actual elapsed seconds
  attachmentId: number
  filename: string
  completed: boolean    // true = timer ran to zero; false = early exit
}

export function todayISO(): string {
  return new Date().toISOString().slice(0, 10)
}

export function loadAllSessions(): FocusSession[] {
  try {
    const raw = localStorage.getItem(SESSIONS_KEY)
    return raw ? (JSON.parse(raw) as FocusSession[]) : []
  } catch { return [] }
}

export function logSession(s: FocusSession): void {
  if (s.durationSec < 60) return   // ignore accidental opens
  try {
    const all = loadAllSessions()
    all.push(s)
    localStorage.setItem(SESSIONS_KEY, JSON.stringify(all))
  } catch { /* quota exceeded — silently drop */ }
}

export function loadSettings(): DeepReadSettings {
  try {
    const raw = localStorage.getItem(SETTINGS_KEY)
    if (!raw) return { ...DEFAULT_SETTINGS }
    return { ...DEFAULT_SETTINGS, ...JSON.parse(raw) }
  } catch {
    return { ...DEFAULT_SETTINGS }
  }
}

export function saveSettings(s: DeepReadSettings) {
  localStorage.setItem(SETTINGS_KEY, JSON.stringify(s))
}

export function loadProgress(attachmentId: number): DeepReadProgress | null {
  try {
    const raw = localStorage.getItem(PROGRESS_KEY(attachmentId))
    return raw ? (JSON.parse(raw) as DeepReadProgress) : null
  } catch {
    return null
  }
}

export function saveProgress(attachmentId: number, p: Omit<DeepReadProgress, 'timestamp'>) {
  if (!attachmentId || p.lastPage < 1) return
  try {
    const data: DeepReadProgress = { ...p, timestamp: Date.now() }
    localStorage.setItem(PROGRESS_KEY(attachmentId), JSON.stringify(data))
  } catch {
    // localStorage might be unavailable (private mode, quota exceeded) — non-fatal
  }
}

export function effectiveZoomMode(p: DeepReadProgress | null): ZoomMode {
  return p?.zoomMode ?? 'fit-width'
}

export const THEMES: { id: DeepReadTheme; label: string; bg: string; fg: string }[] = [
  { id: 'default',  label: 'Default',       bg: '#F9F9FB', fg: '#1f2937' },
  { id: 'sepia',    label: 'Warm / Sepia',  bg: '#F4ECD8', fg: '#4a3520' },
  { id: 'dark',     label: 'Dark',          bg: '#1E1E24', fg: '#e0e0e0' },
  { id: 'contrast', label: 'High Contrast', bg: '#000000', fg: '#FFFF66' }
]

export const TIMER_PRESETS = [10, 15, 25, 50] as const
