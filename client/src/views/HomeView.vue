<template>
  <div class="p-8">
    <!-- Header: title + tab bar + contextual action -->
    <div class="flex items-start justify-between mb-8">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-zinc-100">
          {{ semesterStore.activeSemester?.name ?? 'Select a semester' }}
        </h1>
        <!-- Tab bar -->
        <div class="flex gap-1 mt-3">
          <button
            class="px-3 py-1.5 rounded-lg text-sm font-medium transition-colors"
            :class="tab === 'stats'
              ? 'bg-slate-200 dark:bg-zinc-800 text-slate-900 dark:text-zinc-100'
              : 'text-slate-500 dark:text-zinc-400 hover:text-slate-900 dark:hover:text-zinc-100'"
            @click="tab = 'stats'"
          >
            Focus Stats
          </button>
          <button
            class="px-3 py-1.5 rounded-lg text-sm font-medium transition-colors"
            :class="tab === 'classes'
              ? 'bg-slate-200 dark:bg-zinc-800 text-slate-900 dark:text-zinc-100'
              : 'text-slate-500 dark:text-zinc-400 hover:text-slate-900 dark:hover:text-zinc-100'"
            @click="tab = 'classes'"
          >
            Classes
            <span class="ml-1.5 text-xs font-normal opacity-60">{{ classStore.classes.length }}</span>
          </button>
        </div>
      </div>
      <button
        v-if="tab === 'classes' && semesterStore.activeSemesterID"
        class="flex items-center gap-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors"
        @click="openNewClass"
      >
        <span class="text-lg leading-none">+</span> New Class
      </button>
    </div>

    <!-- Classes tab -->
    <template v-if="tab === 'classes'">
      <div v-if="classStore.loading" class="text-slate-400 text-sm">Loading…</div>

      <!-- Empty state: no semester exists -->
      <div v-else-if="!semesterStore.activeSemesterID" class="max-w-3xl">
        <p class="text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-zinc-500 mb-3">
          Get started
        </p>
        <div class="space-y-2.5">
          <button
            class="w-full text-left flex items-start gap-4 p-4 bg-white dark:bg-[#1e1e1e] border border-slate-200 dark:border-zinc-800 rounded-xl hover:border-indigo-400 dark:hover:border-indigo-500 hover:shadow-sm transition-all group"
            @click="openNewSemester"
          >
            <span class="shrink-0 w-9 h-9 rounded-full bg-indigo-100 dark:bg-indigo-950 text-indigo-600 dark:text-indigo-300 flex items-center justify-center font-semibold">1</span>
            <div class="flex-1 min-w-0">
              <p class="font-medium text-slate-900 dark:text-zinc-100">Create a semester</p>
              <p class="text-xs text-slate-500 dark:text-zinc-400 mt-0.5">Set a name and date range — this groups your classes and powers the calendar.</p>
            </div>
            <span class="text-slate-400 group-hover:text-indigo-500 group-hover:translate-x-0.5 transition-all">→</span>
          </button>

          <div class="w-full text-left flex items-start gap-4 p-4 bg-white dark:bg-[#1e1e1e] border border-slate-200 dark:border-zinc-800 rounded-xl opacity-40 cursor-not-allowed select-none">
            <span class="shrink-0 w-9 h-9 rounded-full bg-slate-100 dark:bg-zinc-800 text-slate-500 flex items-center justify-center font-semibold">2</span>
            <div class="flex-1 min-w-0">
              <p class="font-medium text-slate-900 dark:text-zinc-100">Add your first class</p>
              <p class="text-xs text-slate-500 dark:text-zinc-400 mt-0.5">Available after you create a semester.</p>
            </div>
            <span class="text-slate-300">→</span>
          </div>

          <div class="w-full text-left flex items-start gap-4 p-4 bg-white dark:bg-[#1e1e1e] border border-slate-200 dark:border-zinc-800 rounded-xl opacity-40 cursor-not-allowed select-none">
            <span class="shrink-0 w-9 h-9 rounded-full bg-slate-100 dark:bg-zinc-800 text-slate-500 flex items-center justify-center font-semibold">3</span>
            <div class="flex-1 min-w-0">
              <p class="font-medium text-slate-900 dark:text-zinc-100">Drop a PDF, then Focus Read</p>
              <p class="text-xs text-slate-500 dark:text-zinc-400 mt-0.5">Attach a PDF to any class — it'll show up in your Library, ready for distraction-free reading.</p>
            </div>
            <span class="text-slate-300">→</span>
          </div>
        </div>
      </div>

      <!-- Empty state: semester exists but no classes yet -->
      <div v-else-if="classStore.classes.length === 0" class="max-w-3xl">
        <p class="text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-zinc-500 mb-3">
          Get started
        </p>
        <div class="space-y-2.5">
          <button
            class="w-full text-left flex items-start gap-4 p-4 bg-white dark:bg-[#1e1e1e] border border-slate-200 dark:border-zinc-800 rounded-xl hover:border-indigo-400 dark:hover:border-indigo-500 hover:shadow-sm transition-all group"
            @click="openNewClass"
          >
            <span class="shrink-0 w-9 h-9 rounded-full bg-indigo-100 dark:bg-indigo-950 text-indigo-600 dark:text-indigo-300 flex items-center justify-center font-semibold">1</span>
            <div class="flex-1 min-w-0">
              <p class="font-medium text-slate-900 dark:text-zinc-100">Add your first class</p>
              <p class="text-xs text-slate-500 dark:text-zinc-400 mt-0.5">Set name, color, and instructor. Meeting times come next.</p>
            </div>
            <span class="text-slate-400 group-hover:text-indigo-500 group-hover:translate-x-0.5 transition-all">→</span>
          </button>

          <button
            class="w-full text-left flex items-start gap-4 p-4 bg-white dark:bg-[#1e1e1e] border border-slate-200 dark:border-zinc-800 rounded-xl opacity-60 cursor-not-allowed"
            disabled
          >
            <span class="shrink-0 w-9 h-9 rounded-full bg-slate-100 dark:bg-zinc-800 text-slate-500 flex items-center justify-center font-semibold">2</span>
            <div class="flex-1 min-w-0">
              <p class="font-medium text-slate-900 dark:text-zinc-100">Set up your week</p>
              <p class="text-xs text-slate-500 dark:text-zinc-400 mt-0.5">Add meeting times to populate the calendar. Available after your first class.</p>
            </div>
            <span class="text-slate-300">→</span>
          </button>

          <RouterLink
            to="/library"
            class="w-full text-left flex items-start gap-4 p-4 bg-white dark:bg-[#1e1e1e] border border-slate-200 dark:border-zinc-800 rounded-xl hover:border-indigo-400 dark:hover:border-indigo-500 hover:shadow-sm transition-all group"
          >
            <span class="shrink-0 w-9 h-9 rounded-full bg-indigo-100 dark:bg-indigo-950 text-indigo-600 dark:text-indigo-300 flex items-center justify-center font-semibold">3</span>
            <div class="flex-1 min-w-0">
              <p class="font-medium text-slate-900 dark:text-zinc-100">Drop a PDF, then Focus Read</p>
              <p class="text-xs text-slate-500 dark:text-zinc-400 mt-0.5">Attach a PDF to any class — it'll show up in your Library, ready for distraction-free reading.</p>
            </div>
            <span class="text-slate-400 group-hover:text-indigo-500 group-hover:translate-x-0.5 transition-all">→</span>
          </RouterLink>
        </div>
      </div>

      <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <RouterLink
          v-for="cls in classStore.classes"
          :key="cls.id"
          :to="`/classes/${cls.id}`"
          class="group block bg-white dark:bg-[#1e1e1e] border border-slate-200 dark:border-zinc-800 rounded-xl overflow-hidden hover:shadow-md dark:hover:shadow-zinc-950 transition-shadow"
        >
          <div class="h-2" :style="{ backgroundColor: cls.color }" />
          <div class="p-5">
            <div class="flex items-start justify-between gap-2">
              <div class="min-w-0">
                <p class="font-semibold text-slate-900 dark:text-zinc-100 truncate">{{ cls.name }}</p>
                <p v-if="cls.code" class="text-xs text-slate-500 dark:text-zinc-400 mt-0.5">{{ cls.code }}</p>
              </div>
              <span class="shrink-0 w-3 h-3 rounded-full mt-1" :style="{ backgroundColor: cls.color }" />
            </div>
            <div class="mt-3 space-y-1">
              <p v-if="cls.instructor" class="text-xs text-slate-500 dark:text-zinc-400">{{ cls.instructor }}</p>
              <p v-if="cls.schedule" class="text-xs text-slate-500 dark:text-zinc-400">{{ cls.schedule }}</p>
            </div>
          </div>
        </RouterLink>
      </div>
    </template>

    <!-- Stats tab -->
    <div v-else-if="tab === 'stats'" class="flex flex-col sm:flex-row gap-6 items-start">
      <FocusStatsPanel class="flex-1 min-w-0" />
      <SourceTodoPanel class="w-full sm:w-72 shrink-0" />
    </div>

    <!-- New Class Modal -->
    <BaseModal
      :open="showNewClass"
      :title="step === 1 ? 'New Class' : 'Meeting Times'"
      @close="onModalClose"
    >
      <!-- Step 1: Class Info -->
      <form v-if="step === 1" class="space-y-4" @submit.prevent="goToStep2">
        <div>
          <label class="block text-sm font-medium text-slate-700 dark:text-zinc-300 mb-1">Class name *</label>
          <input v-model="form.name" required class="input" placeholder="e.g. Algorithms" />
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-zinc-300 mb-1">Course code</label>
            <input v-model="form.code" class="input" placeholder="CS 161" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-zinc-300 mb-1">Color</label>
            <input v-model="form.color" type="color" class="h-9 w-full rounded-md border border-slate-300 dark:border-zinc-700 cursor-pointer dark:bg-zinc-800" />
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium text-slate-700 dark:text-zinc-300 mb-1">Instructor</label>
          <input v-model="form.instructor" class="input" placeholder="Prof. Smith" />
        </div>
        <div class="flex justify-end gap-2 pt-2">
          <button type="button" class="btn-ghost" @click="closeModal">Cancel</button>
          <button type="submit" class="btn-primary" :disabled="!form.name || saving">
            {{ saving ? 'Creating…' : 'Next →' }}
          </button>
        </div>
      </form>

      <!-- Step 2: Meeting Times -->
      <div v-else-if="newClassID !== null" class="space-y-4">
        <div>
          <p class="text-sm text-slate-700 dark:text-zinc-300">Add weekly meeting times (optional)</p>
          <p class="text-xs text-slate-500 dark:text-zinc-400 mt-0.5">
            These repeat across the semester and appear on your calendar.
          </p>
        </div>

        <div
          v-if="!hasSemesterDates"
          class="text-xs bg-yellow-50 dark:bg-yellow-900/30 border border-yellow-200 dark:border-yellow-800 text-yellow-800 dark:text-yellow-200 rounded-md px-3 py-2"
        >
          ⚠ Your semester has no start/end dates. Meeting times will be saved but won't appear on the calendar until you set semester dates.
        </div>

        <ClassMeetingsEditor :class-i-d="newClassID" />

        <div class="flex justify-end gap-2 pt-2">
          <button type="button" class="btn-ghost" @click="finishCreate">Skip</button>
          <button type="button" class="btn-primary" @click="finishCreate">Done</button>
        </div>
      </div>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, onMounted, onBeforeUnmount } from 'vue'
import { useSemesterStore } from '@/stores/semester'
import { useClassStore } from '@/stores/class'
import { api, type ClassMeeting } from '@/api'
import BaseModal from '@/components/BaseModal.vue'
import ClassMeetingsEditor from '@/components/ClassMeetingsEditor.vue'
import FocusStatsPanel from '@/components/FocusStatsPanel.vue'
import SourceTodoPanel from '@/components/SourceTodoPanel.vue'

const semesterStore = useSemesterStore()
const classStore = useClassStore()

const tab = ref<'classes' | 'stats'>('stats')

const showNewClass = ref(false)
const step = ref<1 | 2>(1)
const saving = ref(false)
const newClassID = ref<number | null>(null)
const form = ref({ name: '', code: '', instructor: '', color: '#6366f1' })

const hasSemesterDates = computed(() => {
  const s = semesterStore.activeSemester
  return !!(s?.start_date && s?.end_date)
})

watch(() => semesterStore.activeSemesterID, id => {
  if (id) classStore.loadForSemester(id)
}, { immediate: true })

// Sidebar's "+ New class" row dispatches this; HomeView is the modal owner.
function handleNewClassEvent() { openNewClass() }
onMounted(() => window.addEventListener('app:new-class', handleNewClassEvent))
onBeforeUnmount(() => window.removeEventListener('app:new-class', handleNewClassEvent))

function openNewSemester() {
  window.dispatchEvent(new CustomEvent('app:new-semester'))
}

function openNewClass() {
  if (!semesterStore.activeSemesterID) return
  step.value = 1
  newClassID.value = null
  form.value = { name: '', code: '', instructor: '', color: '#6366f1' }
  showNewClass.value = true
}

function closeModal() {
  showNewClass.value = false
}

async function goToStep2() {
  const semID = semesterStore.activeSemesterID
  if (!semID || !form.value.name || saving.value) return
  saving.value = true
  try {
    const c = await classStore.create(semID, form.value)
    newClassID.value = c.id
    step.value = 2
  } finally {
    saving.value = false
  }
}

async function finishCreate() {
  const id = newClassID.value
  closeModal()
  if (id === null) return
  const meetings = await api.meetings.list(id)
  const label = deriveSchedule(meetings)
  if (label) await classStore.update(id, { schedule: label })
}

function onModalClose() {
  if (step.value === 2) finishCreate()
  else closeModal()
}

// --- Schedule label derivation ---
const DAY_ABBR = ['Su', 'M', 'Tu', 'W', 'Th', 'F', 'Sa']

function parseTime(t: string) {
  const [hStr, mStr] = t.split(':')
  const h = Number(hStr)
  const m = Number(mStr)
  const period: 'am' | 'pm' = h >= 12 ? 'pm' : 'am'
  return { h12: h % 12 || 12, m, period }
}

function timeStr(p: { h12: number; m: number }) {
  return p.m === 0 ? String(p.h12) : `${p.h12}:${String(p.m).padStart(2, '0')}`
}

function timeRange(start: string, end: string): string {
  const s = parseTime(start)
  const e = parseTime(end)
  if (s.period === e.period) return `${timeStr(s)}–${timeStr(e)}${e.period}`
  return `${timeStr(s)}${s.period}–${timeStr(e)}${e.period}`
}

function deriveSchedule(meetings: ClassMeeting[]): string {
  if (meetings.length === 0) return ''
  const groups = new Map<string, { start: string; end: string; days: number[] }>()
  for (const m of meetings) {
    const key = `${m.start_time}|${m.end_time}`
    const existing = groups.get(key)
    if (existing) existing.days.push(m.day_of_week)
    else groups.set(key, { start: m.start_time, end: m.end_time, days: [m.day_of_week] })
  }
  const parts: string[] = []
  for (const { start, end, days } of groups.values()) {
    const sorted = [...new Set(days)].sort((a, b) => a - b)
    const dayStr = sorted.map(d => DAY_ABBR[d]).join('')
    parts.push(`${dayStr} ${timeRange(start, end)}`)
  }
  return parts.join(', ')
}
</script>
