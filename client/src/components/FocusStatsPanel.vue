<template>
  <div v-if="allSessions.length === 0" class="rounded-2xl bg-white dark:bg-[#141414]/80 border border-slate-200 dark:border-zinc-800/50 p-8 text-center max-w-3xl">
    <p class="text-3xl mb-3">📖</p>
    <p class="font-semibold text-slate-800 dark:text-zinc-200">No focus sessions yet</p>
    <p class="text-sm text-slate-500 dark:text-zinc-400 mt-1">Start a Deep Read session from the Library to see your stats here.</p>
  </div>

  <div v-else class="rounded-2xl bg-white dark:bg-[#141414]/80 border border-slate-200 dark:border-zinc-800/50 p-6 max-w-3xl space-y-5">
    <!-- Header row -->
    <div class="flex items-center justify-between gap-4">
      <h2 class="text-slate-900 dark:text-white font-semibold text-base">Focus Stats</h2>
      <div class="flex gap-1">
        <button
          v-for="p in PERIODS"
          :key="p.value"
          class="px-3 py-1 rounded-full text-xs font-medium transition-colors"
          :class="period === p.value
            ? 'bg-slate-900 dark:bg-white text-white dark:text-slate-900'
            : 'text-slate-400 dark:text-zinc-400 hover:text-slate-700 dark:hover:text-zinc-200'"
          @click="period = p.value"
        >{{ p.label }}</button>
      </div>
    </div>

    <!-- Stat cards: 4-column grid -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
      <div
        v-for="card in statCards"
        :key="card.label"
        class="bg-slate-50 dark:bg-zinc-800/40 border border-slate-100 dark:border-transparent rounded-xl p-3 space-y-1 min-w-0"
        :title="card.title ?? ''"
      >
        <p class="text-xs text-slate-500 dark:text-zinc-400 leading-tight truncate">{{ card.label }}</p>
        <p class="text-xl font-semibold text-slate-900 dark:text-white leading-tight truncate">{{ card.value }}</p>
      </div>
    </div>

    <!-- Empty period message -->
    <p v-if="filtered.length === 0 && period !== 'all'" class="text-sm text-slate-400 dark:text-zinc-500 text-center py-2">
      No sessions in the last {{ period === '30d' ? '30' : '7' }} days.
    </p>

    <!-- Heatmap -->
    <div>
      <p class="text-xs text-slate-400 dark:text-zinc-500 mb-2 uppercase tracking-wider">Activity — past 5 months</p>
      <div class="overflow-x-auto">
        <div class="flex gap-[3px]" style="min-width: max-content">
          <div
            v-for="(week, wi) in heatmapWeeks"
            :key="wi"
            class="flex flex-col gap-[3px]"
          >
            <div
              v-for="(cell, di) in week"
              :key="di"
              class="w-[11px] h-[11px] rounded-sm transition-colors"
              :class="cellColor(cell.totalSec)"
              :title="cell.date ? cellTooltip(cell) : ''"
            />
          </div>
        </div>
      </div>
      <div class="flex items-center gap-2 mt-2">
        <span class="text-[10px] text-slate-400 dark:text-zinc-600">Less</span>
        <div v-for="c in LEGEND" :key="c" class="w-[11px] h-[11px] rounded-sm" :class="c" />
        <span class="text-[10px] text-slate-400 dark:text-zinc-600">More</span>
      </div>
    </div>

    <!-- Fun comparison -->
    <p class="text-xs text-slate-400 dark:text-zinc-500 italic border-t border-slate-100 dark:border-zinc-800/50 pt-4">{{ funLine }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { todayISO } from '@/composables/useDeepRead'
import { api, type FocusSession } from '@/api'

type Period = 'all' | '30d' | '7d'

const PERIODS: { label: string; value: Period }[] = [
  { label: 'All', value: 'all' },
  { label: '30d', value: '30d' },
  { label: '7d', value: '7d' }
]

const LEGEND = [
  'bg-slate-200 dark:bg-zinc-800/40',
  'bg-indigo-200 dark:bg-indigo-900',
  'bg-indigo-400 dark:bg-indigo-600',
  'bg-indigo-600 dark:bg-indigo-400'
]

const period = ref<Period>('all')
const allSessions = ref<FocusSession[]>([])

onMounted(async () => {
  await migrateLocalStorage()
  await refreshSessions()
  window.addEventListener('focus', refreshSessions)
  window.addEventListener('app:focus-session-ended', refreshSessions)
})
onUnmounted(() => {
  window.removeEventListener('focus', refreshSessions)
  window.removeEventListener('app:focus-session-ended', refreshSessions)
})

async function refreshSessions() {
  try {
    allSessions.value = await api.focusSessions.list()
  } catch { /* server unavailable — keep existing data */ }
}

// One-time migration of localStorage sessions to the DB
async function migrateLocalStorage() {
  const SESSIONS_KEY = 'deepread:sessions'
  const raw = localStorage.getItem(SESSIONS_KEY)
  if (!raw) return
  try {
    type OldSession = { date: string; durationSec: number; attachmentId: number; filename: string; completed: boolean }
    const old = JSON.parse(raw) as OldSession[]
    await Promise.all(old
      .filter(s => s.duration_sec >= 60)
      .map(s => api.focusSessions.log({
        date: s.date,
        duration_sec: s.duration_sec,
        attachment_id: s.attachmentId,
        filename: s.filename,
        completed: s.completed
      }))
    )
    localStorage.removeItem(SESSIONS_KEY)
  } catch { /* leave localStorage intact if migration fails */ }
}

// --- Filtered sessions for stat cards ---

const filtered = computed<FocusSession[]>(() => {
  if (period.value === 'all') return allSessions.value
  const cutoff = new Date()
  cutoff.setDate(cutoff.getDate() - (period.value === '30d' ? 30 : 7))
  const cutoffISO = cutoff.toISOString().slice(0, 10)
  return allSessions.value.filter(s => s.date >= cutoffISO)
})

// --- Derived stats ---

function fmtDuration(sec: number): string {
  if (sec < 60) return `${sec}s`
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  if (h === 0) return `${m}m`
  return m === 0 ? `${h}h` : `${h}h ${m}m`
}

const totalSec = computed(() => filtered.value.reduce((a, s) => a + s.duration_sec, 0))
const activeDays = computed(() => new Set(filtered.value.map(s => s.date)).size)

const streaks = computed(() => computeStreaks(allSessions.value))

function computeStreaks(sessions: FocusSession[]): { current: number; longest: number } {
  const dates = [...new Set(sessions.map(s => s.date))].sort()
  if (dates.length === 0) return { current: 0, longest: 0 }

  let longest = 1, run = 1
  for (let i = 1; i < dates.length; i++) {
    const diff = (new Date(dates[i]).getTime() - new Date(dates[i - 1]).getTime()) / 86_400_000
    if (diff === 1) { run++; if (run > longest) longest = run }
    else run = 1
  }

  const today = todayISO()
  const yesterday = new Date(Date.now() - 86_400_000).toISOString().slice(0, 10)
  const dateSet = new Set(dates)
  let current = 0
  if (dateSet.has(today) || dateSet.has(yesterday)) {
    let check = dateSet.has(today) ? today : yesterday
    while (dateSet.has(check)) {
      current++
      const d = new Date(check)
      d.setDate(d.getDate() - 1)
      check = d.toISOString().slice(0, 10)
    }
  }
  return { current, longest }
}

const peakDay = computed(() => {
  const DAY = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
  const totals = new Array(7).fill(0)
  for (const s of filtered.value) {
    totals[new Date(s.date + 'T12:00:00').getDay()] += s.duration_sec
  }
  const max = Math.max(...totals)
  return max === 0 ? '—' : DAY[totals.indexOf(max)]
})

// Sessions that exceeded the 3-minute grace period count toward the Completed ratio.
const countableSessions = computed(() => filtered.value.filter(s => s.duration_sec >= 180))
const completedSessions = computed(() => countableSessions.value.filter(s => s.completed))

const statCards = computed(() => [
  { label: 'Total time',    value: fmtDuration(totalSec.value) },
  { label: 'Active days',   value: String(activeDays.value) },
  { label: 'Cur. streak',   value: streaks.value.current + 'd' },
  { label: 'Longest streak',value: streaks.value.longest + 'd' },
  { label: 'Peak day',      value: peakDay.value },
  { label: 'Completed',     value: countableSessions.value.length === 0 ? '—' : `${completedSessions.value.length} / ${countableSessions.value.length}` }
])

// --- Heatmap (always rolling year, ignores period filter) ---

interface HeatCell { date: string; totalSec: number; count: number }

const heatmapWeeks = computed<HeatCell[][]>(() => {
  // Build a map of date → totals from all sessions
  const byDate = new Map<string, { totalSec: number; count: number }>()
  for (const s of allSessions.value) {
    const e = byDate.get(s.date) ?? { totalSec: 0, count: 0 }
    e.totalSec += s.duration_sec
    e.count++
    byDate.set(s.date, e)
  }

  // Build 20 full weeks (~5 months) ending with the current week.
  const WEEKS = 20
  const todayDate = new Date()
  todayDate.setHours(0, 0, 0, 0)
  // Anchor the end to Saturday of the current week
  const dayOfWeek = todayDate.getDay()
  const endSunday = new Date(todayDate)
  endSunday.setDate(todayDate.getDate() - dayOfWeek + 6)

  const startDate = new Date(endSunday)
  startDate.setDate(endSunday.getDate() - WEEKS * 7 + 1)
  // align to Sunday
  startDate.setDate(startDate.getDate() - startDate.getDay())

  const weeks: HeatCell[][] = []
  let cur = new Date(startDate)
  while (cur <= endSunday) {
    const week: HeatCell[] = []
    for (let d = 0; d < 7; d++) {
      const iso = cur.toISOString().slice(0, 10)
      const isFuture = cur > todayDate
      const entry = byDate.get(iso)
      week.push({ date: isFuture ? '' : iso, totalSec: isFuture ? -1 : (entry?.totalSec ?? 0), count: entry?.count ?? 0 })
      cur.setDate(cur.getDate() + 1)
    }
    weeks.push(week)
  }
  return weeks
})

function cellColor(totalSec: number): string {
  if (totalSec < 0) return 'bg-slate-100 dark:bg-[#1e1e1e]/30'          // future cell
  if (totalSec === 0) return 'bg-slate-200 dark:bg-zinc-800/40'        // no activity
  if (totalSec < 900) return 'bg-indigo-200 dark:bg-indigo-900'         // < 15 min
  if (totalSec < 2700) return 'bg-indigo-400 dark:bg-indigo-600'        // < 45 min
  return 'bg-indigo-600 dark:bg-indigo-400'                              // 45+ min
}

function cellTooltip(cell: HeatCell): string {
  if (!cell.date) return ''
  const d = new Date(cell.date + 'T12:00:00')
  const label = d.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })
  if (cell.totalSec === 0) return `${label} · no sessions`
  return `${label} · ${fmtDuration(cell.totalSec)} · ${cell.count} session${cell.count !== 1 ? 's' : ''}`
}

// --- Fun comparison (all-time) ---

const allTimeSec = computed(() => allSessions.value.reduce((a, s) => a + s.duration_sec, 0))

const funLine = computed(() => {
  const sec = allTimeSec.value
  if (sec < 1800) return 'Every minute counts. Keep going.'
  if (sec < 7_200) return "You've focused long enough to watch a feature film."
  if (sec < 18_000) return "You've read for the equivalent of a full school day."
  if (sec < 36_000) return "You've accumulated a full work day of deep reading."
  if (sec < 86_400) return "You've put in more focused hours than most students do in a week."
  const days = Math.floor(sec / 86_400)
  return `You've spent ${days} full day${days !== 1 ? 's' : ''} in deep focus. Remarkable.`
})
</script>
