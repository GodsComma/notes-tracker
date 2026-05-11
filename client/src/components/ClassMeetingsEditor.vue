<template>
  <div class="space-y-3">
    <div v-if="meetings.length === 0" class="text-xs text-slate-400 py-2">
      No meeting slots yet.
    </div>

    <div v-else class="divide-y divide-slate-100 dark:divide-slate-700 border border-slate-200 dark:border-slate-700 rounded-lg overflow-hidden">
      <div
        v-for="m in meetings"
        :key="m.id"
        class="flex items-center gap-3 px-3 py-2.5 bg-white dark:bg-slate-800 text-sm"
      >
        <span class="w-10 font-medium text-slate-700 dark:text-slate-300 shrink-0">{{ dayName(m.day_of_week) }}</span>
        <span class="text-slate-600 dark:text-slate-400">{{ formatTime(m.start_time) }}–{{ formatTime(m.end_time) }}</span>
        <span v-if="m.location" class="text-slate-400 dark:text-slate-500 truncate flex-1">{{ m.location }}</span>
        <span v-else class="flex-1" />
        <button
          class="shrink-0 text-slate-300 hover:text-red-500 transition-colors"
          title="Remove"
          @click="remove(m.id)"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Add form -->
    <div v-if="adding" class="border border-indigo-200 dark:border-indigo-800 rounded-lg p-3 bg-indigo-50 dark:bg-indigo-950 space-y-3">
      <div class="grid grid-cols-2 gap-2">
        <div>
          <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Day</label>
          <select v-model="form.day_of_week" class="input text-sm">
            <option v-for="(name, i) in DAYS" :key="i" :value="i">{{ name }}</option>
          </select>
        </div>
        <div>
          <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Location</label>
          <input v-model="form.location" class="input text-sm" placeholder="Gates B03" />
        </div>
      </div>
      <div class="grid grid-cols-2 gap-2">
        <div>
          <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Start time</label>
          <input v-model="form.start_time" type="time" class="input text-sm" />
        </div>
        <div>
          <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">End time</label>
          <input v-model="form.end_time" type="time" class="input text-sm" />
        </div>
      </div>
      <div class="flex justify-end gap-2">
        <button class="btn-ghost text-xs" @click="adding = false">Cancel</button>
        <button class="btn-primary text-xs" :disabled="!form.start_time || !form.end_time" @click="add">Add</button>
      </div>
    </div>

    <button v-else class="text-sm text-indigo-600 hover:text-indigo-800 font-medium flex items-center gap-1" @click="adding = true">
      <span class="text-base leading-none">+</span> Add time slot
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type ClassMeeting } from '@/api'

const props = defineProps<{ classID: number }>()
const emit = defineEmits<{ change: [] }>()

const DAYS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']

const meetings = ref<ClassMeeting[]>([])
const adding = ref(false)
const form = ref({ day_of_week: 1, start_time: '', end_time: '', location: '' })

onMounted(async () => {
  meetings.value = await api.meetings.list(props.classID)
})

async function add() {
  if (!form.value.start_time || !form.value.end_time) return
  const m = await api.meetings.create(props.classID, form.value)
  meetings.value.push(m)
  form.value = { day_of_week: 1, start_time: '', end_time: '', location: '' }
  adding.value = false
  emit('change')
}

async function remove(id: number) {
  await api.meetings.delete(id)
  meetings.value = meetings.value.filter(m => m.id !== id)
  emit('change')
}

function dayName(dow: number) {
  return DAYS[dow]?.slice(0, 3) ?? ''
}

function formatTime(t: string) {
  if (!t) return ''
  const [h, m] = t.split(':').map(Number)
  const ampm = h >= 12 ? 'pm' : 'am'
  const h12 = h % 12 || 12
  return m === 0 ? `${h12}${ampm}` : `${h12}:${String(m).padStart(2, '0')}${ampm}`
}
</script>
