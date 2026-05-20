<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="$emit('close')">
      <div class="bg-white dark:bg-[#1e1e1e] rounded-xl shadow-xl w-full max-w-md mx-4 p-6">
        <div class="flex items-start justify-between mb-5">
          <div>
            <h2 class="text-base font-semibold text-slate-900 dark:text-zinc-100">Assign to…</h2>
            <p class="text-sm text-slate-500 dark:text-zinc-400 mt-0.5 truncate max-w-xs">{{ item.filename }}</p>
          </div>
          <button
            type="button"
            class="text-slate-400 hover:text-slate-600 dark:hover:text-zinc-200 ml-4 shrink-0"
            @click="$emit('close')"
          >
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div v-if="loadingClasses" class="text-sm text-slate-400 py-4 text-center">Loading…</div>

        <template v-else>
          <!-- Mode toggle -->
          <div class="flex gap-2 mb-4">
            <button
              type="button"
              :class="mode === 'class' ? 'bg-indigo-600 text-white' : 'bg-slate-100 dark:bg-zinc-800 text-slate-700 dark:text-zinc-300 hover:bg-slate-200 dark:hover:bg-zinc-700'"
              class="flex-1 text-sm font-medium py-1.5 rounded-lg transition-colors"
              @click="mode = 'class'; selectedSourceId = null"
            >
              Class
            </button>
            <button
              type="button"
              :class="mode === 'source' ? 'bg-indigo-600 text-white' : 'bg-slate-100 dark:bg-zinc-800 text-slate-700 dark:text-zinc-300 hover:bg-slate-200 dark:hover:bg-zinc-700'"
              class="flex-1 text-sm font-medium py-1.5 rounded-lg transition-colors"
              @click="mode = 'source'; selectedSourceId = null"
            >
              Source (inside a class)
            </button>
          </div>

          <!-- No classes state -->
          <div v-if="allClasses.length === 0" class="text-sm text-slate-500 dark:text-zinc-400 py-4 text-center">
            <p>You don't have any classes yet.</p>
            <RouterLink to="/" class="text-indigo-600 dark:text-indigo-400 hover:underline mt-1 inline-block">
              Create one →
            </RouterLink>
          </div>

          <template v-else>
            <!-- Class selector -->
            <div class="mb-3">
              <label class="block text-xs font-medium text-slate-600 dark:text-zinc-400 mb-1">Class</label>
              <select
                v-model="selectedClassId"
                class="w-full border border-slate-300 dark:border-zinc-700 rounded-lg px-3 py-2 text-sm bg-white dark:bg-zinc-800 text-slate-900 dark:text-zinc-100 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                @change="onClassChange"
              >
                <option :value="null" disabled>Select a class…</option>
                <optgroup
                  v-for="sem in semesterGroups"
                  :key="sem.name"
                  :label="sem.name"
                >
                  <option v-for="cls in sem.classes" :key="cls.id" :value="cls.id">
                    {{ cls.name }}{{ cls.code ? ` (${cls.code})` : '' }}
                  </option>
                </optgroup>
              </select>
            </div>

            <!-- Source selector (only shown in source mode) -->
            <div v-if="mode === 'source' && selectedClassId" class="mb-3">
              <label class="block text-xs font-medium text-slate-600 dark:text-zinc-400 mb-1">Source</label>
              <div v-if="loadingSources" class="text-xs text-slate-400 py-2">Loading sources…</div>
              <select
                v-else-if="sources.length > 0"
                v-model="selectedSourceId"
                class="w-full border border-slate-300 dark:border-zinc-700 rounded-lg px-3 py-2 text-sm bg-white dark:bg-zinc-800 text-slate-900 dark:text-zinc-100 focus:outline-none focus:ring-2 focus:ring-indigo-500"
              >
                <option :value="null" disabled>Select a source…</option>
                <option v-for="src in sources" :key="src.id" :value="src.id">
                  {{ src.title || src.type }}
                </option>
              </select>
              <p v-else class="text-xs text-slate-400 py-2">No sources in this class yet.</p>
            </div>
          </template>

          <!-- Error -->
          <p v-if="error" class="text-sm text-red-500 mb-3">{{ error }}</p>

          <!-- Actions -->
          <div class="flex justify-end gap-3 mt-5">
            <button
              type="button"
              class="text-sm font-medium px-4 py-2 rounded-lg text-slate-600 dark:text-zinc-300 hover:bg-slate-100 dark:hover:bg-zinc-800 transition-colors"
              @click="$emit('close')"
            >
              Cancel
            </button>
            <button
              type="button"
              :disabled="!canSubmit || saving"
              class="text-sm font-medium px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed text-white transition-colors"
              @click="submit"
            >
              {{ saving ? 'Assigning…' : 'Assign' }}
            </button>
          </div>
        </template>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { api, type PDFLibraryItem, type Class, type Source, type Semester } from '@/api'

const props = defineProps<{ item: PDFLibraryItem }>()
const emit = defineEmits<{ close: []; assigned: [] }>()

type SemesterGroup = { name: string; classes: Class[] }

const mode = ref<'class' | 'source'>('class')
const allClasses = ref<Class[]>([])
const semesterGroups = ref<SemesterGroup[]>([])
const sources = ref<Source[]>([])
const selectedClassId = ref<number | null>(null)
const selectedSourceId = ref<number | null>(null)
const loadingClasses = ref(true)
const loadingSources = ref(false)
const saving = ref(false)
const error = ref('')

const canSubmit = computed(() => {
  if (!selectedClassId.value) return false
  if (mode.value === 'source') return !!selectedSourceId.value
  return true
})

onMounted(async () => {
  try {
    const semesters: Semester[] = await api.semesters.list()
    const grouped: SemesterGroup[] = []
    const all: Class[] = []
    await Promise.all(semesters.map(async (sem) => {
      const cls = await api.classes.list(sem.id)
      if (cls.length > 0) {
        grouped.push({ name: sem.name, classes: cls })
        all.push(...cls)
      }
    }))
    semesterGroups.value = grouped
    allClasses.value = all
  } finally {
    loadingClasses.value = false
  }
})

async function onClassChange() {
  selectedSourceId.value = null
  if (!selectedClassId.value || mode.value !== 'source') return
  loadingSources.value = true
  try {
    sources.value = await api.sources.list(selectedClassId.value)
  } finally {
    loadingSources.value = false
  }
}

watch(mode, async (m) => {
  selectedSourceId.value = null
  if (m === 'source' && selectedClassId.value) {
    await onClassChange()
  }
})

async function submit() {
  error.value = ''
  if (!canSubmit.value) return
  saving.value = true
  try {
    if (mode.value === 'class') {
      await api.attachments.assign(props.item.id, { class_id: selectedClassId.value! })
    } else {
      await api.attachments.assign(props.item.id, { source_id: selectedSourceId.value! })
    }
    emit('assigned')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to assign'
  } finally {
    saving.value = false
  }
}
</script>
