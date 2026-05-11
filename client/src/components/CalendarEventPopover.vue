<template>
  <BaseModal :open="!!event" :title="popoverTitle" @close="$emit('close')">
    <div v-if="event" class="space-y-3">
      <!-- Type badge -->
      <div class="flex items-center gap-2">
        <span
          class="text-xs font-medium px-2 py-0.5 rounded-full"
          :class="typeBadgeClass"
        >{{ typeLabel }}</span>
        <span v-if="event.className" class="text-xs text-slate-500">{{ event.className }}</span>
      </div>

      <!-- Date / time -->
      <div class="text-sm text-slate-600">
        <span v-if="event.type === 'meeting'">
          {{ formatDate(event.start) }} · {{ formatTime(event.start) }}
          <span v-if="event.end"> – {{ formatTime(event.end) }}</span>
        </span>
        <span v-else>{{ formatDate(event.start) }}</span>
      </div>

      <!-- Location (meetings only) -->
      <div v-if="event.type === 'meeting' && event.location" class="text-sm text-slate-500 flex items-center gap-1.5">
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
        {{ event.location }}
      </div>

      <!-- Source type (for source / deadline events) -->
      <div v-if="event.sourceType && event.type !== 'meeting'" class="text-xs text-slate-400">
        {{ event.sourceType }}
      </div>

      <!-- Action -->
      <div v-if="event.type !== 'meeting' && event.sourceID" class="pt-1">
        <RouterLink
          :to="`/sources/${event.sourceID}`"
          class="inline-flex items-center gap-1.5 text-sm text-indigo-600 hover:text-indigo-800 font-medium"
          @click="$emit('close')"
        >
          Open notes
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
          </svg>
        </RouterLink>
      </div>
      <div v-else-if="event.type === 'meeting' && event.classID" class="pt-1">
        <RouterLink
          :to="`/classes/${event.classID}`"
          class="inline-flex items-center gap-1.5 text-sm text-indigo-600 hover:text-indigo-800 font-medium"
          @click="$emit('close')"
        >
          Open class
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
          </svg>
        </RouterLink>
      </div>
    </div>
  </BaseModal>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import BaseModal from './BaseModal.vue'
import type { CalendarEventProps } from '@/api'

type PopoverEvent = CalendarEventProps & { title: string; start: string; end?: string; location?: string }

const props = defineProps<{ event: PopoverEvent | null }>()
defineEmits<{ close: [] }>()

const popoverTitle = computed(() => props.event?.title ?? '')

const typeLabel = computed(() => {
  switch (props.event?.type) {
    case 'source': return props.event.sourceType ?? 'Source'
    case 'deadline': return 'Deadline'
    case 'meeting': return 'Class Meeting'
    default: return ''
  }
})

const typeBadgeClass = computed(() => {
  switch (props.event?.type) {
    case 'deadline': return 'bg-rose-100 text-rose-700'
    case 'meeting': return 'bg-slate-100 text-slate-600'
    default: {
      const map: Record<string, string> = {
        Lecture: 'bg-blue-100 text-blue-700',
        Reading: 'bg-purple-100 text-purple-700',
        'Problem Set': 'bg-orange-100 text-orange-700',
        Discussion: 'bg-green-100 text-green-700',
        Lab: 'bg-teal-100 text-teal-700',
      }
      return map[props.event?.sourceType ?? ''] ?? 'bg-slate-100 text-slate-600'
    }
  }
})

function formatDate(iso: string) {
  const d = iso.slice(0, 10)
  return new Date(d + 'T00:00:00').toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric' })
}

function formatTime(iso: string) {
  if (!iso.includes('T')) return ''
  const [, time] = iso.split('T')
  const [h, m] = time.split(':').map(Number)
  const ampm = h >= 12 ? 'pm' : 'am'
  const h12 = h % 12 || 12
  return m === 0 ? `${h12}${ampm}` : `${h12}:${String(m).padStart(2, '0')}${ampm}`
}
</script>
