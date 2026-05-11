<template>
  <ConfirmModal
    :open="!!pendingComplete"
    title="Mark as complete?"
    message="Mark this item as completed. You can undo this by clicking the green circle again."
    confirm-label="Mark complete"
    :danger="false"
    @confirm="confirmComplete"
    @cancel="pendingComplete = null"
  />
  <div class="rounded-2xl bg-white dark:bg-slate-900/80 border border-slate-200 dark:border-slate-700/50 p-5 flex flex-col gap-4">
    <!-- Header -->
    <div class="flex items-start justify-between gap-2">
      <div>
        <h2 class="text-slate-900 dark:text-white font-semibold text-base">Upcoming</h2>
        <p class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">
          {{ semesterName }}
          <span v-if="totalIncomplete > 0"> · {{ totalIncomplete }} {{ weekFilter === 'this' ? 'this week' : 'next week' }}</span>
        </p>
      </div>
      <div class="flex gap-0.5 shrink-0">
        <button
          v-for="w in WEEK_OPTIONS"
          :key="w.value"
          class="px-2 py-0.5 rounded-full text-[10px] font-medium transition-colors"
          :class="weekFilter === w.value
            ? 'bg-slate-900 dark:bg-white text-white dark:text-slate-900'
            : 'text-slate-400 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200'"
          @click="weekFilter = w.value"
        >{{ w.label }}</button>
      </div>
    </div>

    <div v-if="loading" class="text-sm text-slate-400 dark:text-slate-500">Loading…</div>

    <div v-else-if="incomplete.length === 0 && completed.length === 0" class="text-sm text-slate-400 dark:text-slate-500 py-2">
      Nothing due {{ weekFilter === 'this' ? 'this week' : 'next week' }}.
    </div>

    <template v-else>
      <!-- Incomplete items grouped by priority -->
      <div v-for="group in visibleGroups" :key="group.priority" class="space-y-1.5">
        <p
          v-if="hasMultipleGroups"
          class="text-[10px] font-semibold uppercase tracking-wider"
          :class="groupLabelClass(group.priority)"
        >{{ group.label }}</p>

        <div
          v-for="src in group.items"
          :key="src.id"
          class="flex items-start gap-2 rounded-lg px-3 py-2 bg-slate-50 dark:bg-slate-800/60 hover:bg-slate-100 dark:hover:bg-slate-700/60 transition-colors group"
        >
          <!-- Completion circle -->
          <button
            type="button"
            class="mt-0.5 shrink-0 w-4 h-4 rounded-full border-2 transition-colors flex items-center justify-center border-slate-300 dark:border-slate-600 hover:border-green-500 dark:hover:border-green-400"
            title="Mark as complete"
            @click.prevent="requestComplete(src)"
          />

          <RouterLink :to="`/sources/${src.id}`" class="flex-1 min-w-0">
            <p class="text-sm font-medium text-slate-800 dark:text-slate-200 truncate group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">
              {{ src.title }}
            </p>
            <div class="flex items-center gap-2 mt-0.5">
              <span class="flex items-center gap-1 text-[10px] text-slate-400 dark:text-slate-500 truncate">
                <span class="w-1.5 h-1.5 rounded-full shrink-0" :style="{ backgroundColor: src.class_color }" />
                {{ src.class_name }}
              </span>
              <span v-if="src.due_date" class="text-[10px] font-medium shrink-0" :class="dueDateClass(src.due_date)">
                {{ dueDateLabel(src.due_date) }}
              </span>
            </div>
          </RouterLink>

          <PriorityBadge :priority="src.priority" size="xs" :editable="false" class="mt-0.5 shrink-0" />
        </div>
      </div>

      <!-- Show-all footer for incomplete items -->
      <div v-if="totalIncomplete > MAX_VISIBLE" class="border-t border-slate-100 dark:border-slate-700/50 pt-2">
        <RouterLink
          :to="`/classes/${firstClassId}`"
          class="text-xs text-indigo-500 hover:text-indigo-700 dark:hover:text-indigo-300 font-medium"
        >Show all → ({{ totalIncomplete }} items)</RouterLink>
      </div>

      <!-- Completed items (collapsed section) -->
      <div v-if="completed.length > 0" class="border-t border-slate-100 dark:border-slate-700/50 pt-3 space-y-1">
        <button
          type="button"
          class="text-[10px] font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 flex items-center gap-1.5 hover:text-slate-600 dark:hover:text-slate-300 transition-colors"
          @click="showCompleted = !showCompleted"
        >
          <svg class="w-3 h-3 transition-transform" :class="{ 'rotate-90': showCompleted }" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
          </svg>
          Completed ({{ completed.length }})
        </button>

        <template v-if="showCompleted">
          <div
            v-for="src in completed"
            :key="src.id"
            class="flex items-start gap-2 rounded-lg px-3 py-2 opacity-50 group"
          >
            <!-- Filled circle = done -->
            <button
              type="button"
              class="mt-0.5 shrink-0 w-4 h-4 rounded-full border-2 border-green-500 bg-green-500 flex items-center justify-center transition-colors hover:bg-transparent hover:border-slate-400"
              title="Mark as incomplete"
              @click.prevent="toggle(src)"
            >
              <svg class="w-2.5 h-2.5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
                <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
              </svg>
            </button>
            <RouterLink :to="`/sources/${src.id}`" class="flex-1 min-w-0">
              <p class="text-sm text-slate-400 dark:text-slate-500 truncate line-through">{{ src.title }}</p>
              <span class="text-[10px] text-slate-400 dark:text-slate-500 flex items-center gap-1">
                <span class="w-1.5 h-1.5 rounded-full shrink-0" :style="{ backgroundColor: src.class_color }" />
                {{ src.class_name }}
              </span>
            </RouterLink>
          </div>
        </template>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { api, type SourceWithClass } from '@/api'
import { useSemesterStore } from '@/stores/semester'
import PriorityBadge from './PriorityBadge.vue'
import ConfirmModal from './ConfirmModal.vue'

const MAX_VISIBLE = 8

type WeekFilter = 'this' | 'next'
const WEEK_OPTIONS: { label: string; value: WeekFilter }[] = [
  { label: 'This wk', value: 'this' },
  { label: 'Next wk', value: 'next' }
]

const semesterStore = useSemesterStore()
const sources = ref<SourceWithClass[]>([])
const loading = ref(false)
const showCompleted = ref(false)
const weekFilter = ref<WeekFilter>('this')

function getWeekRange(offset: 0 | 1): { start: string; end: string } {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const day = today.getDay()
  const daysToMonday = day === 0 ? -6 : 1 - day
  const monday = new Date(today)
  monday.setDate(today.getDate() + daysToMonday + offset * 7)
  const sunday = new Date(monday)
  sunday.setDate(monday.getDate() + 6)
  return { start: monday.toISOString().slice(0, 10), end: sunday.toISOString().slice(0, 10) }
}

const semesterName = computed(() => semesterStore.activeSemester?.name ?? '')

async function load() {
  const id = semesterStore.activeSemesterID
  if (!id) { sources.value = []; return }
  loading.value = true
  try {
    sources.value = await api.sources.listForSemester(id)
  } catch {
    sources.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => { load(); window.addEventListener('focus', load) })
onUnmounted(() => window.removeEventListener('focus', load))
watch(() => semesterStore.activeSemesterID, load)

const incomplete = computed(() => {
  const { start, end } = getWeekRange(weekFilter.value === 'this' ? 0 : 1)
  return sources.value.filter(s => {
    if (s.completed) return false
    const d = s.due_date || s.date
    return d >= start && d <= end
  })
})
const completed  = computed(() => sources.value.filter(s => s.completed))
const totalIncomplete = computed(() => incomplete.value.length)
const firstClassId = computed(() => sources.value[0]?.class_id ?? 0)

async function toggle(src: SourceWithClass) {
  const updated = await api.sources.update(src.id, { completed: !src.completed })
  const idx = sources.value.findIndex(s => s.id === src.id)
  if (idx !== -1) sources.value[idx] = { ...sources.value[idx], ...updated }
}

const pendingComplete = ref<SourceWithClass | null>(null)
function requestComplete(src: SourceWithClass) { pendingComplete.value = src }
async function confirmComplete() {
  if (!pendingComplete.value) return
  await toggle(pendingComplete.value)
  pendingComplete.value = null
}

// Grouping — only incomplete items, capped at MAX_VISIBLE
interface Group { priority: number; label: string; items: SourceWithClass[] }

const visibleGroups = computed<Group[]>(() => {
  const visible = incomplete.value.slice(0, MAX_VISIBLE)
  const map = new Map<number, SourceWithClass[]>()
  for (const s of visible) {
    const bucket = map.get(s.priority) ?? []
    bucket.push(s)
    map.set(s.priority, bucket)
  }
  const LABELS: Record<number, string> = { 1: 'High', 2: 'Medium', 3: 'Low', 0: 'No priority' }
  return [1, 2, 3, 0]
    .filter(p => map.has(p))
    .map(p => ({ priority: p, label: LABELS[p], items: map.get(p)! }))
})

const hasMultipleGroups = computed(() => visibleGroups.value.length > 1)

function groupLabelClass(priority: number): string {
  if (priority === 1) return 'text-rose-500'
  if (priority === 2) return 'text-amber-500'
  if (priority === 3) return 'text-blue-400'
  return 'text-slate-400 dark:text-slate-500'
}

function dueDateClass(date: string): string {
  const today = new Date(); today.setHours(0, 0, 0, 0)
  const d = new Date(date + 'T00:00:00')
  const diff = Math.floor((d.getTime() - today.getTime()) / 86_400_000)
  if (diff < 0) return 'text-rose-500'
  if (diff <= 3) return 'text-amber-500'
  return 'text-slate-400 dark:text-slate-500'
}

function dueDateLabel(date: string): string {
  const today = new Date(); today.setHours(0, 0, 0, 0)
  const d = new Date(date + 'T00:00:00')
  const diff = Math.floor((d.getTime() - today.getTime()) / 86_400_000)
  if (diff < 0) return 'Overdue'
  if (diff === 0) return 'Due today'
  if (diff === 1) return 'Due tomorrow'
  return `Due ${d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}`
}
</script>
