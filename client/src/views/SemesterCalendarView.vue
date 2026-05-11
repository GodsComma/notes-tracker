<template>
  <div class="p-8">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">Semester Calendar</h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">{{ semesterStore.activeSemester?.name }}</p>
      </div>
    </div>

    <!-- Class legend / filter -->
    <div v-if="classStore.classes.length > 0" class="flex flex-wrap gap-2 mb-5">
      <button
        v-for="cls in classStore.classes"
        :key="cls.id"
        class="flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-full border transition-all"
        :class="hiddenClassIDs.has(cls.id)
          ? 'border-slate-200 dark:border-slate-700 text-slate-400 dark:text-slate-500 bg-white dark:bg-slate-800'
          : 'border-transparent text-white'"
        :style="hiddenClassIDs.has(cls.id) ? {} : { backgroundColor: cls.color }"
        @click="toggleClass(cls.id)"
      >
        <span
          class="w-2 h-2 rounded-full shrink-0"
          :style="{ backgroundColor: hiddenClassIDs.has(cls.id) ? '#cbd5e1' : 'white' }"
        />
        {{ cls.name }}
      </button>
    </div>

    <CalendarView
      :events="filteredEvents"
      :loading="loading"
      :initial-date="currentMonth"
      @month-change="onMonthChange"
      @event-click="popoverEvent = $event"
    />

    <CalendarEventPopover :event="popoverEvent" @close="popoverEvent = null" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { api, type CalendarEvent, type CalendarEventProps } from '@/api'
import { useSemesterStore } from '@/stores/semester'
import { useClassStore } from '@/stores/class'
import CalendarView from '@/components/CalendarView.vue'
import CalendarEventPopover from '@/components/CalendarEventPopover.vue'

const semesterStore = useSemesterStore()
const classStore = useClassStore()

const allEvents = ref<CalendarEvent[]>([])
const loading = ref(false)
const currentMonth = ref(todayMonth())
const hiddenClassIDs = ref(new Set<number>())
const popoverEvent = ref<(CalendarEventProps & { title: string; start: string }) | null>(null)

function todayMonth() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
}

const filteredEvents = computed(() =>
  hiddenClassIDs.value.size === 0
    ? allEvents.value
    : allEvents.value.filter(e => !hiddenClassIDs.value.has(e.extendedProps.classID ?? -1))
)

function toggleClass(id: number) {
  const next = new Set(hiddenClassIDs.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  hiddenClassIDs.value = next
}

async function fetchEvents(month: string) {
  const semID = semesterStore.activeSemesterID
  if (!semID) return
  loading.value = true
  try {
    allEvents.value = await api.calendar.forSemester(semID, month)
  } finally {
    loading.value = false
  }
}

async function onMonthChange(month: string) {
  if (month !== currentMonth.value) {
    currentMonth.value = month
    await fetchEvents(month)
  }
}

onMounted(() => fetchEvents(currentMonth.value))

watch(() => semesterStore.activeSemesterID, () => fetchEvents(currentMonth.value))
</script>
