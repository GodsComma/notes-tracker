<template>
  <div class="flex h-screen overflow-hidden bg-slate-50 dark:bg-slate-900">
    <!-- Sidebar -->
    <aside
      v-if="!sidebarCollapsed"
      class="w-60 shrink-0 bg-slate-800 dark:bg-slate-900 flex flex-col overflow-hidden border-r border-slate-700/50"
    >
      <!-- Brand + collapse -->
      <div class="px-5 py-5 flex items-center justify-between">
        <span class="text-white font-bold text-lg tracking-tight">Class Notes</span>
        <button
          class="text-slate-400 hover:text-white transition-colors"
          title="Collapse sidebar"
          @click="toggleSidebar"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
          </svg>
        </button>
      </div>

      <!-- Primary nav -->
      <nav class="px-2 pb-3 space-y-0.5">
        <RouterLink
          v-for="link in primaryNav"
          :key="link.to"
          :to="link.to"
          class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors"
          :class="[
            $route.path === link.to || (link.matchPrefix && $route.path.startsWith(link.matchPrefix))
              ? 'bg-slate-700 text-white'
              : 'text-slate-300 hover:text-white hover:bg-slate-700/50'
          ]"
        >
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" :d="link.icon" />
          </svg>
          <span>{{ link.label }}</span>
        </RouterLink>

        <form class="px-3 pt-2" @submit.prevent="doSearch">
          <div class="relative">
            <svg class="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              v-model="searchQ"
              type="search"
              placeholder="Search notes…"
              class="w-full pl-8 pr-2 py-1.5 text-xs bg-slate-700/50 border border-slate-700 rounded-md text-slate-200 placeholder:text-slate-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
            />
          </div>
        </form>
      </nav>

      <!-- Classes section -->
      <div class="px-4 py-3 border-t border-slate-700/50">
        <button
          class="w-full flex items-center justify-between text-xs font-semibold uppercase tracking-wider text-slate-400 hover:text-slate-200 transition-colors"
          @click="semesterPickerOpen = !semesterPickerOpen"
        >
          <span class="flex items-center gap-1.5 truncate">
            CLASSES
            <span v-if="semesterStore.activeSemester" class="font-medium normal-case tracking-normal text-slate-500 truncate">
              · {{ semesterStore.activeSemester.name }}
            </span>
          </span>
          <svg class="w-3 h-3 shrink-0 transition-transform" :class="{ 'rotate-180': semesterPickerOpen }" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
          </svg>
        </button>

        <!-- Inline semester picker -->
        <div v-if="semesterPickerOpen" class="mt-2 space-y-1.5">
          <select
            v-if="semesterStore.semesters.length > 0"
            :value="semesterStore.activeSemesterID"
            class="w-full bg-slate-700/50 text-slate-200 text-xs rounded px-2 py-1.5 border border-slate-700 focus:outline-none focus:ring-1 focus:ring-indigo-500"
            @change="onSemesterChange"
          >
            <option v-for="s in semesterStore.semesters" :key="s.id" :value="s.id">{{ s.name }}</option>
          </select>
          <button
            class="w-full text-xs text-indigo-300 hover:text-indigo-200 text-left px-1"
            @click="showNewSemester = true; semesterPickerOpen = false"
          >
            + New semester
          </button>
        </div>
      </div>

      <!-- Class list -->
      <nav class="flex-1 overflow-y-auto py-1 space-y-0.5 px-2">
        <RouterLink
          v-for="cls in classStore.classes"
          :key="cls.id"
          :to="`/classes/${cls.id}`"
          class="flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm transition-colors"
          :class="[
            $route.path.startsWith(`/classes/${cls.id}`)
              ? 'bg-slate-700 text-white'
              : 'text-slate-300 hover:text-white hover:bg-slate-700/50'
          ]"
        >
          <span class="w-2.5 h-2.5 rounded-full shrink-0" :style="{ backgroundColor: cls.color }" />
          <span class="truncate">{{ cls.name }}</span>
        </RouterLink>

        <button
          v-if="semesterStore.activeSemesterID"
          class="w-full flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm text-slate-400 hover:text-indigo-300 hover:bg-slate-700/40 transition-colors"
          @click="$router.push('/'); $nextTick(() => emitNewClass())"
        >
          <span class="w-2.5 h-2.5 rounded-full shrink-0 border border-dashed border-slate-500" />
          <span>+ New class</span>
        </button>
      </nav>

      <!-- Date warning -->
      <div v-if="semesterStore.activeSemester && !hasActiveSemesterDates" class="px-4 py-3 border-t border-slate-700/50">
        <button
          class="text-xs text-yellow-400 hover:text-yellow-300 text-left w-full leading-snug"
          @click="toggleDateEdit"
        >
          ⚠ Set semester dates so class meetings appear on the calendar.
        </button>
        <div v-if="editingDates" class="mt-2 space-y-2 bg-slate-700/40 rounded-md p-2 border border-slate-700">
          <div>
            <label class="block text-[10px] uppercase tracking-wider text-slate-500 mb-1">Start</label>
            <input v-model="dateForm.start" type="date" class="w-full bg-slate-800 text-slate-200 text-xs rounded px-2 py-1 border border-slate-700 focus:outline-none focus:ring-1 focus:ring-indigo-500 [color-scheme:dark]" />
          </div>
          <div>
            <label class="block text-[10px] uppercase tracking-wider text-slate-500 mb-1">End</label>
            <input v-model="dateForm.end" type="date" class="w-full bg-slate-800 text-slate-200 text-xs rounded px-2 py-1 border border-slate-700 focus:outline-none focus:ring-1 focus:ring-indigo-500 [color-scheme:dark]" />
          </div>
          <button
            class="w-full bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white text-xs font-medium py-1.5 rounded"
            :disabled="!dateForm.start || !dateForm.end || savingDates"
            @click="saveDates"
          >
            {{ savingDates ? 'Saving…' : 'Save dates' }}
          </button>
        </div>
      </div>
    </aside>

    <!-- Main content -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- Top bar (slimmed: just open-sidebar + theme toggle) -->
      <header class="shrink-0 h-12 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-700 flex items-center justify-between px-6">
        <button
          v-if="sidebarCollapsed"
          class="p-1.5 rounded-md text-slate-500 hover:text-slate-900 hover:bg-slate-100 dark:hover:text-slate-100 dark:hover:bg-slate-800 transition-colors"
          title="Open sidebar"
          @click="toggleSidebar"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
        <span v-else />
        <ThemeToggle />
      </header>

      <!-- Router view -->
      <main class="flex-1 overflow-y-auto">
        <RouterView :key="$route.path" />
      </main>
    </div>

    <!-- New Semester Modal -->
    <BaseModal :open="showNewSemester" title="New Semester" @close="showNewSemester = false">
      <form class="space-y-4" @submit.prevent="createSemester">
        <div>
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Name *</label>
          <input v-model="semForm.name" required class="input" placeholder="Fall 2026" />
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Start date</label>
            <input v-model="semForm.startDate" type="date" class="input" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">End date</label>
            <input v-model="semForm.endDate" type="date" class="input" />
          </div>
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400 -mt-2">
          Required for calendar event generation.
        </p>
        <div class="flex justify-end gap-2 pt-2">
          <button type="button" class="btn-ghost" @click="showNewSemester = false">Cancel</button>
          <button type="submit" class="btn-primary" :disabled="!semForm.name">Create</button>
        </div>
      </form>
    </BaseModal>

    <!-- Global Deep Read pre-flight + focus session.
         Mounted here so any view can launch them via the focus store. -->
    <DeepReadPreflight
      v-if="focusStore.pendingPreflight"
      :open="!!focusStore.pendingPreflight"
      :item="focusStore.pendingPreflight.item"
      @close="focusStore.cancelPreflight()"
      @start="onPreflightStart"
    />
    <DeepReadFocus
      v-if="focusStore.active"
      :url="api.attachments.fileUrl(focusStore.active.item.id, 'inline')"
      :attachment-id="focusStore.active.item.id"
      :source-id="focusStore.active.item.source_id"
      :settings="focusStore.active.settings"
      :filename="focusStore.active.item.filename"
      @exit="onSessionEnd"
      @complete="onSessionEnd"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useSemesterStore } from '@/stores/semester'
import { useClassStore } from '@/stores/class'
import { useFocusStore } from '@/stores/focus'
import BaseModal from './BaseModal.vue'
import ThemeToggle from './ThemeToggle.vue'
import DeepReadPreflight from './DeepReadPreflight.vue'
import DeepReadFocus from './DeepReadFocus.vue'
import { api } from '@/api'
import type { DeepReadSettings } from '@/composables/useDeepRead'

const router = useRouter()
const semesterStore = useSemesterStore()
const classStore = useClassStore()
const focusStore = useFocusStore()

const showNewSemester = ref(false)
const semesterPickerOpen = ref(false)
const searchQ = ref('')
const semForm = ref({ name: '', startDate: '', endDate: '' })

const sidebarCollapsed = ref(localStorage.getItem('sidebar:collapsed') === '1')

const editingDates = ref(false)
const savingDates = ref(false)
const dateForm = ref({ start: '', end: '' })

const hasActiveSemesterDates = computed(() => {
  const s = semesterStore.activeSemester
  return !!(s?.start_date && s?.end_date)
})

const primaryNav = [
  {
    to: '/',
    label: 'Home',
    icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0h6'
  },
  {
    to: '/calendar',
    label: 'Calendar',
    icon: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z'
  },
  {
    to: '/library',
    label: 'Library',
    icon: 'M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253',
    matchPrefix: '/deepread'
  }
]

onMounted(async () => {
  await semesterStore.load()
  if (semesterStore.activeSemesterID) {
    await classStore.loadForSemester(semesterStore.activeSemesterID)
  }
})

watch(() => semesterStore.activeSemesterID, async (id) => {
  editingDates.value = false
  if (id) await classStore.loadForSemester(id)
})

async function createSemester() {
  await semesterStore.create(semForm.value.name, semForm.value.startDate, semForm.value.endDate)
  semForm.value = { name: '', startDate: '', endDate: '' }
  showNewSemester.value = false
}

function onSemesterChange(e: Event) {
  const id = Number((e.target as HTMLSelectElement).value)
  semesterStore.setActive(id)
  semesterPickerOpen.value = false
}

function toggleDateEdit() {
  if (!editingDates.value) {
    const s = semesterStore.activeSemester
    dateForm.value = { start: s?.start_date ?? '', end: s?.end_date ?? '' }
  }
  editingDates.value = !editingDates.value
}

async function saveDates() {
  const s = semesterStore.activeSemester
  if (!s || !dateForm.value.start || !dateForm.value.end) return
  savingDates.value = true
  try {
    await semesterStore.update(s.id, s.name, dateForm.value.start, dateForm.value.end)
    editingDates.value = false
  } finally {
    savingDates.value = false
  }
}

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
  localStorage.setItem('sidebar:collapsed', sidebarCollapsed.value ? '1' : '0')
}

function doSearch() {
  if (searchQ.value.trim()) {
    router.push({ path: '/search', query: { q: searchQ.value.trim() } })
  }
}

function emitNewClass() {
  window.dispatchEvent(new CustomEvent('app:new-class'))
}

onMounted(() => window.addEventListener('app:new-semester', () => { showNewSemester.value = true }))
// Note: no cleanup needed — AppShell lives for the full app lifetime.

function onPreflightStart(settings: DeepReadSettings) {
  focusStore.startSession(settings)
}

function onSessionEnd() {
  const returnTo = focusStore.active?.returnTo
  const sourceId = focusStore.active?.item.source_id
  focusStore.endSession()
  if (sourceId !== null && sourceId !== undefined) {
    window.dispatchEvent(new CustomEvent('app:focus-session-ended', { detail: { sourceId } }))
  }
  if (returnTo) {
    nextTick(() => router.replace(returnTo))
  }
}
</script>
