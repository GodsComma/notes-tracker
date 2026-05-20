<template>
  <BaseModal :open="open" :title="rubric ? 'Edit rubric' : 'Set up rubric'" @close="$emit('close')">
    <div v-if="loading" class="text-sm text-slate-400">Loading…</div>

    <form v-else class="space-y-5" @submit.prevent="save">
      <!-- Grading type -->
      <div>
        <label class="block text-sm font-medium text-slate-700 dark:text-zinc-300 mb-1.5">Grading type</label>
        <div class="flex rounded-lg border border-slate-200 dark:border-zinc-800 overflow-hidden text-sm w-fit">
          <button
            type="button"
            class="px-4 py-1.5 font-medium transition-colors"
            :class="draft.grading_type === 'letter'
              ? 'bg-slate-900 text-white dark:bg-zinc-800'
              : 'bg-white dark:bg-[#1e1e1e] text-slate-600 dark:text-zinc-300 hover:bg-slate-50 dark:hover:bg-zinc-800'"
            @click="draft.grading_type = 'letter'"
          >Letter grades</button>
          <button
            type="button"
            class="px-4 py-1.5 font-medium transition-colors border-l border-slate-200 dark:border-zinc-800"
            :class="draft.grading_type === 'pass_fail'
              ? 'bg-slate-900 text-white dark:bg-zinc-800'
              : 'bg-white dark:bg-[#1e1e1e] text-slate-600 dark:text-zinc-300 hover:bg-slate-50 dark:hover:bg-zinc-800'"
            @click="draft.grading_type = 'pass_fail'"
          >Pass/Fail</button>
        </div>
      </div>

      <!-- Plus/minus toggle -->
      <div v-if="draft.grading_type === 'letter'">
        <label class="flex items-center gap-2 text-sm text-slate-700 dark:text-zinc-300 cursor-pointer">
          <input v-model="draft.plus_minus" type="checkbox" class="rounded" />
          Use sub-grades (A+/A/A−, B+/B/B−, …)
        </label>
      </div>

      <!-- Custom thresholds (collapsible) -->
      <div>
        <button
          type="button"
          class="text-xs font-medium text-slate-500 hover:text-slate-700 dark:text-zinc-400 dark:hover:text-zinc-200 flex items-center gap-1"
          @click="showThresholds = !showThresholds"
        >
          <span class="inline-block w-3">{{ showThresholds ? '▾' : '▸' }}</span>
          Custom thresholds (advanced)
        </button>
        <div v-if="showThresholds" class="mt-2 grid grid-cols-3 gap-2">
          <div
            v-for="row in thresholdRows"
            :key="row.letter"
            class="flex items-center gap-2"
          >
            <span class="text-sm font-mono w-8 text-slate-700 dark:text-zinc-300">{{ row.letter }}</span>
            <input
              type="number"
              min="0"
              max="100"
              step="0.5"
              class="input"
              :value="thresholdMap[row.letter] ?? row.defaultValue"
              @input="onThresholdChange(row.letter, ($event.target as HTMLInputElement).value)"
            />
          </div>
        </div>
      </div>

      <!-- Categories -->
      <div>
        <div class="flex items-center justify-between mb-2">
          <label class="block text-sm font-medium text-slate-700 dark:text-zinc-300">Categories</label>
          <button
            type="button"
            class="text-sm text-indigo-600 hover:text-indigo-800 font-medium"
            @click="addCategory"
          >+ Add category</button>
        </div>

        <div v-if="draft.categories.length === 0" class="text-sm text-slate-400 py-3 text-center border border-dashed border-slate-200 dark:border-zinc-800 rounded-md">
          No categories yet. Add at least one.
        </div>

        <div v-else class="space-y-2 max-h-64 overflow-y-auto pr-1">
          <div
            v-for="(cat, i) in draft.categories"
            :key="cat.tempId"
            class="grid grid-cols-12 gap-2 items-center"
          >
            <input
              v-model="cat.name"
              type="text"
              placeholder="Category name"
              class="input col-span-5"
              :class="{ 'border-red-400': blankNames && !cat.name.trim() }"
            />
            <div class="col-span-2 relative">
              <input
                v-model.number="cat.weight"
                type="number"
                min="0"
                step="0.5"
                placeholder="0"
                class="input pr-6"
                :class="{ 'border-red-400': cat.weight <= 0 }"
              />
              <span class="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-slate-400">%</span>
            </div>
            <div class="col-span-2">
              <input
                v-model.number="cat.possible"
                type="number"
                min="1"
                step="1"
                placeholder="100"
                class="input"
                title="Possible points"
              />
            </div>
            <label class="col-span-2 flex items-center gap-1.5 text-xs text-slate-600 dark:text-zinc-400 cursor-pointer">
              <input v-model="cat.is_extra_credit" type="checkbox" class="rounded" />
              Extra
            </label>
            <button
              type="button"
              class="col-span-1 text-slate-300 hover:text-red-500 flex justify-center"
              title="Remove category"
              @click="removeCategory(i)"
            >
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      <!-- Weight summary -->
      <div>
        <div class="flex items-center justify-between text-xs mb-1.5">
          <span class="text-slate-600 dark:text-zinc-400">
            Weights total: <span class="font-medium" :class="weightStatusClass">{{ totalWeight.toFixed(1) }}%</span>
            <span v-if="extraCreditTotal > 0" class="text-slate-400"> · {{ extraCreditTotal.toFixed(1) }}% extra credit</span>
          </span>
        </div>
        <div class="h-2 rounded-full bg-slate-100 dark:bg-zinc-800 overflow-hidden">
          <div
            class="h-full transition-all"
            :class="weightStatusBarClass"
            :style="{ width: Math.min(100, totalWeight) + '%' }"
          />
        </div>
        <p v-if="warning" class="text-xs mt-1.5" :class="warningClass">{{ warning }}</p>
      </div>

      <!-- Footer -->
      <div class="flex justify-between items-center pt-3 border-t border-slate-100 dark:border-zinc-800">
        <button
          v-if="rubric"
          type="button"
          class="text-red-500 hover:text-red-700 text-sm font-medium"
          @click="confirmDelete"
        >Delete rubric</button>
        <span v-else />
        <div class="flex gap-2">
          <button type="button" class="btn-ghost" @click="$emit('close')">Cancel</button>
          <button
            type="submit"
            class="btn-primary"
            :disabled="!canSave || saving"
          >{{ saving ? 'Saving…' : 'Save rubric' }}</button>
        </div>
      </div>
    </form>
  </BaseModal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import BaseModal from '@/components/BaseModal.vue'
import { api, type GradeRubric } from '@/api'
import { defaultThresholdsFor } from '@/composables/useGrades'

const props = defineProps<{
  open: boolean
  classID: number
  rubric: GradeRubric | null
}>()

const emit = defineEmits<{
  close: []
  saved: [rubric: GradeRubric]
  deleted: []
}>()

// `id` present = persisted row, absent = new row to POST.
type DraftCategory = {
  id?: number
  tempId: number
  name: string
  weight: number
  is_extra_credit: boolean
  possible: number
}

const loading = ref(false)
const saving = ref(false)
const showThresholds = ref(false)
const blankNames = ref(false)
let nextTempId = 1

const draft = ref<{
  grading_type: 'letter' | 'pass_fail'
  plus_minus: boolean
  categories: DraftCategory[]
}>({
  grading_type: 'letter',
  plus_minus: false,
  categories: []
})

// thresholdMap holds only user-overridden values. We compare against defaults
// at save time to decide whether to persist '' (defaults) or a JSON map.
const thresholdMap = ref<Record<string, number>>({})

// Snapshot of category IDs that exist server-side, so we can compute deletes.
let originalIDs: number[] = []

watch(() => props.open, (open) => {
  if (open) reset()
})

function reset() {
  blankNames.value = false
  saving.value = false
  showThresholds.value = false
  if (props.rubric) {
    draft.value = {
      grading_type: props.rubric.grading_type,
      plus_minus: props.rubric.plus_minus,
      categories: props.rubric.categories.map(c => ({
        id: c.id,
        tempId: nextTempId++,
        name: c.name,
        weight: c.weight,
        is_extra_credit: c.is_extra_credit,
        possible: c.possible
      }))
    }
    originalIDs = props.rubric.categories.map(c => c.id)
    if (props.rubric.thresholds) {
      try {
        thresholdMap.value = JSON.parse(props.rubric.thresholds)
        showThresholds.value = true
      } catch { thresholdMap.value = {} }
    } else {
      thresholdMap.value = {}
    }
  } else {
    draft.value = {
      grading_type: 'letter',
      plus_minus: false,
      categories: []
    }
    originalIDs = []
    thresholdMap.value = {}
  }
}

// --- Thresholds ---

const thresholdRows = computed(() => {
  const defaults = defaultThresholdsFor(draft.value.grading_type, draft.value.plus_minus)
  return Object.entries(defaults)
    .sort((a, b) => b[1] - a[1])
    .map(([letter, defaultValue]) => ({ letter, defaultValue }))
})

function onThresholdChange(letter: string, raw: string) {
  const v = parseFloat(raw)
  if (Number.isFinite(v)) {
    thresholdMap.value = { ...thresholdMap.value, [letter]: v }
  }
}

function thresholdsForSave(): string {
  const defaults = defaultThresholdsFor(draft.value.grading_type, draft.value.plus_minus)
  const customized: Record<string, number> = {}
  let anyDifferent = false
  for (const [letter, defaultValue] of Object.entries(defaults)) {
    const v = thresholdMap.value[letter]
    if (v !== undefined && v !== defaultValue) {
      customized[letter] = v
      anyDifferent = true
    } else {
      customized[letter] = defaultValue
    }
  }
  return anyDifferent ? JSON.stringify(customized) : ''
}

// --- Categories ---

function addCategory() {
  draft.value.categories.push({
    tempId: nextTempId++,
    name: '',
    weight: 10,
    is_extra_credit: false,
    possible: 100
  })
}

function removeCategory(i: number) {
  draft.value.categories.splice(i, 1)
}

// --- Validation summary ---

const totalWeight = computed(() =>
  draft.value.categories
    .filter(c => !c.is_extra_credit)
    .reduce((s, c) => s + (Number(c.weight) || 0), 0)
)

const extraCreditTotal = computed(() =>
  draft.value.categories
    .filter(c => c.is_extra_credit)
    .reduce((s, c) => s + (Number(c.weight) || 0), 0)
)

const weightStatusClass = computed(() => {
  if (totalWeight.value === 0) return 'text-red-500'
  if (totalWeight.value < 100) return 'text-yellow-600 dark:text-yellow-400'
  if (totalWeight.value > 100) return 'text-amber-600 dark:text-amber-400'
  return 'text-green-600 dark:text-green-400'
})

const weightStatusBarClass = computed(() => {
  if (totalWeight.value === 0) return 'bg-red-400'
  if (totalWeight.value < 100) return 'bg-yellow-400'
  if (totalWeight.value > 100) return 'bg-amber-400'
  return 'bg-green-500'
})

const warning = computed(() => {
  if (draft.value.categories.length === 0) return ''
  if (totalWeight.value < 100) {
    return `Rubric covers ${totalWeight.value.toFixed(1)}% — needs 100% to calculate a full grade.`
  }
  if (totalWeight.value > 100) {
    return `Weights sum to ${totalWeight.value.toFixed(1)}% — some categories may overlap.`
  }
  return ''
})

const warningClass = computed(() =>
  totalWeight.value < 100
    ? 'text-yellow-600 dark:text-yellow-400'
    : 'text-amber-600 dark:text-amber-400'
)

const canSave = computed(() => {
  if (draft.value.categories.length === 0) return false
  if (totalWeight.value === 0) return false
  for (const c of draft.value.categories) {
    if (!c.name.trim()) return false
    if (!Number.isFinite(c.weight) || c.weight <= 0) return false
    if (!Number.isFinite(c.possible) || c.possible <= 0) return false
  }
  return true
})

// --- Save / Delete ---

async function save() {
  if (!canSave.value) {
    blankNames.value = true
    return
  }
  saving.value = true
  try {
    const rubric = await api.grades.putRubric(props.classID, {
      grading_type: draft.value.grading_type,
      plus_minus: draft.value.plus_minus,
      thresholds: thresholdsForSave()
    })

    // Diff categories: delete removed, update existing, create new.
    const draftIDs = draft.value.categories.filter(c => c.id).map(c => c.id!)
    const toDelete = originalIDs.filter(id => !draftIDs.includes(id))
    await Promise.all(toDelete.map(id => api.grades.deleteCategory(id)))

    for (let i = 0; i < draft.value.categories.length; i++) {
      const c = draft.value.categories[i]
      const payload = {
        name: c.name.trim(),
        weight: Number(c.weight),
        is_extra_credit: c.is_extra_credit,
        possible: Number(c.possible) || 100
      }
      if (c.id) {
        await api.grades.updateCategory(c.id, { ...payload, sort_order: i })
      } else {
        await api.grades.addCategory(rubric.id, payload)
      }
    }

    // Re-fetch the canonical rubric so the parent has current state.
    const fresh = await api.grades.getRubric(props.classID)
    emit('saved', fresh)
    emit('close')
  } catch (e) {
    console.error('failed to save rubric', e)
  } finally {
    saving.value = false
  }
}

async function confirmDelete() {
  if (!props.rubric) return
  if (!window.confirm('Delete this rubric and all its categories? Scores will be lost.')) return
  await api.grades.deleteRubric(props.classID)
  emit('deleted')
  emit('close')
}

</script>
