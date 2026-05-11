<template>
  <div class="calendar-wrapper relative">
    <div v-if="loading" class="absolute inset-0 bg-white/70 flex items-center justify-center z-10 rounded-xl">
      <span class="text-sm text-slate-400">Loading…</span>
    </div>
    <FullCalendar :options="calendarOptions" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import FullCalendar from '@fullcalendar/vue3'
import dayGridPlugin from '@fullcalendar/daygrid'
import type { CalendarOptions, DatesSetArg, EventClickArg } from '@fullcalendar/core'
import type { CalendarEvent, CalendarEventProps } from '@/api'

const props = defineProps<{
  events: CalendarEvent[]
  initialDate?: string
  loading?: boolean
}>()

const emit = defineEmits<{
  'month-change': [month: string]
  'event-click': [props: CalendarEventProps & { title: string; start: string }]
}>()

function handleDatesSet(info: DatesSetArg) {
  const d = info.view.currentStart
  const month = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
  emit('month-change', month)
}

function handleEventClick(info: EventClickArg) {
  emit('event-click', {
    ...(info.event.extendedProps as CalendarEventProps),
    title: info.event.title,
    start: info.event.startStr
  })
}

const calendarOptions = computed<CalendarOptions>(() => ({
  plugins: [dayGridPlugin],
  initialView: 'dayGridMonth',
  initialDate: props.initialDate ? props.initialDate + '-01' : undefined,
  events: props.events as CalendarOptions['events'],
  eventClick: handleEventClick,
  datesSet: handleDatesSet,
  headerToolbar: {
    left: 'prev,next today',
    center: 'title',
    right: ''
  },
  height: 'auto',
  eventDisplay: 'block',
  dayMaxEvents: 4,
  eventTimeFormat: { hour: 'numeric', minute: '2-digit', omitZeroMinute: true }
}))
</script>

<style>
/* FullCalendar base overrides — must be global (not scoped) */
.calendar-wrapper .fc {
  font-family: inherit;
}
.calendar-wrapper .fc-toolbar-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: #0f172a;
}
.calendar-wrapper .fc-button {
  background: white !important;
  border: 1px solid #e2e8f0 !important;
  color: #475569 !important;
  font-size: 0.8rem !important;
  padding: 0.25rem 0.625rem !important;
  border-radius: 0.375rem !important;
  box-shadow: none !important;
}
.calendar-wrapper .fc-button:hover {
  background: #f8fafc !important;
  color: #0f172a !important;
}
.calendar-wrapper .fc-button-active {
  background: #f1f5f9 !important;
}
.calendar-wrapper .fc-daygrid-day-number {
  font-size: 0.8rem;
  color: #64748b;
  padding: 4px 6px;
}
.calendar-wrapper .fc-day-today .fc-daygrid-day-number {
  background: #6366f1;
  color: white;
  border-radius: 9999px;
  width: 1.5rem;
  height: 1.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 4px 6px;
  padding: 0;
}
.calendar-wrapper .fc-col-header-cell-cushion {
  font-size: 0.75rem;
  font-weight: 600;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.calendar-wrapper .fc-event {
  border: none !important;
  border-radius: 4px !important;
  font-size: 0.72rem !important;
  font-weight: 500 !important;
  padding: 1px 4px !important;
  cursor: pointer !important;
}
.calendar-wrapper .fc-daygrid-day-frame {
  min-height: 80px;
}
.calendar-wrapper .fc-scrollgrid {
  border-radius: 0.75rem;
  overflow: hidden;
  border-color: #e2e8f0 !important;
}
.calendar-wrapper .fc-scrollgrid td,
.calendar-wrapper .fc-scrollgrid th {
  border-color: #f1f5f9 !important;
}

/* ─── Dark mode overrides ─────────────────────────────────── */
.dark .calendar-wrapper .fc-toolbar-title {
  color: #f1f5f9;
}
.dark .calendar-wrapper .fc-button {
  background: #1e293b !important;
  border-color: #334155 !important;
  color: #94a3b8 !important;
}
.dark .calendar-wrapper .fc-button:hover {
  background: #334155 !important;
  color: #e2e8f0 !important;
}
.dark .calendar-wrapper .fc-col-header-cell-cushion {
  color: #64748b;
}
.dark .calendar-wrapper .fc-daygrid-day-number {
  color: #94a3b8;
}
.dark .calendar-wrapper .fc-day-today {
  background: rgba(99, 102, 241, 0.08) !important;
}
.dark .calendar-wrapper .fc-day-today .fc-daygrid-day-number {
  background: #6366f1;
  color: white;
}
.dark .calendar-wrapper .fc-scrollgrid {
  border-color: #334155 !important;
}
.dark .calendar-wrapper .fc-scrollgrid td,
.dark .calendar-wrapper .fc-scrollgrid th {
  border-color: #1e293b !important;
}
.dark .calendar-wrapper .fc-daygrid-day {
  background: #0f172a;
}
.dark .calendar-wrapper .fc-col-header {
  background: #1e293b;
}
</style>
