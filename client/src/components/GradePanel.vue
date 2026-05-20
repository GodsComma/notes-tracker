<template>
  <div>
    <!-- Loading -->
    <div v-if="loading" class="text-sm text-slate-400">Loading…</div>

    <!-- Empty state -->
    <div
      v-else-if="!rubric"
      class="text-center py-16 px-4 max-w-md mx-auto"
    >
      <p class="text-4xl mb-3">📊</p>
      <p class="text-base font-medium text-slate-700 dark:text-zinc-300 mb-1">No grading rubric set up</p>
      <p class="text-sm text-slate-500 dark:text-zinc-400 mb-5">
        Set up a rubric to track your running grade for this class.
      </p>
      <button class="btn-primary" @click="openEditor">Set up rubric</button>
    </div>

    <!-- Rubric exists -->
    <div v-else class="space-y-6">
      <!-- Grade overview -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <GradeStat label="Current grade" :rubric="rubric" :pct="summary.current" :primary="true" />
        <GradeStat label="Projected (worst)" :rubric="rubric" :pct="summary.hasAnyScored ? summary.projected : null" />
        <GradeStat label="Best possible" :rubric="rubric" :pct="summary.best" />
      </div>

      <!-- Breakdown -->
      <div>
        <div class="flex items-center justify-between mb-2.5">
          <h3 class="text-sm font-semibold text-slate-700 dark:text-zinc-300">Grading breakdown</h3>
          <button
            class="text-sm text-indigo-600 hover:text-indigo-800 font-medium"
            @click="openEditor"
          >Edit rubric</button>
        </div>

        <div class="rounded-lg border border-slate-200 dark:border-zinc-800 overflow-hidden">
          <table class="w-full text-sm">
            <thead class="bg-slate-50 dark:bg-[#1e1e1e] text-slate-500 dark:text-zinc-400">
              <tr>
                <th class="text-left px-4 py-2 font-medium">Category</th>
                <th class="text-right px-4 py-2 font-medium w-20">Weight</th>
                <th class="text-right px-4 py-2 font-medium w-44">Score</th>
                <th class="text-right px-4 py-2 font-medium w-24">Contribution</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="cat in rubric.categories"
                :key="cat.id"
                class="border-t border-slate-100 dark:border-zinc-800"
              >
                <td class="px-4 py-2.5 text-slate-800 dark:text-zinc-200">
                  <div class="flex items-center gap-2">
                    <span>{{ cat.name }}</span>
                    <span
                      v-if="cat.is_extra_credit"
                      class="text-[10px] uppercase tracking-wide font-medium px-1.5 py-0.5 rounded bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-200"
                    >Extra</span>
                  </div>
                </td>
                <td class="px-4 py-2.5 text-right text-slate-600 dark:text-zinc-400">{{ cat.weight }}%</td>
                <td class="px-4 py-2.5">
                  <ScoreInput
                    :category="cat"
                    @change="updateScore(cat, $event)"
                  />
                </td>
                <td class="px-4 py-2.5 text-right tabular-nums" :class="contribClass(cat)">
                  {{ contribLabel(cat) }}
                </td>
              </tr>
            </tbody>
            <tfoot class="bg-slate-50 dark:bg-[#1e1e1e] text-slate-600 dark:text-zinc-400 text-xs">
              <tr>
                <td class="px-4 py-2 font-medium">Total</td>
                <td class="px-4 py-2 text-right font-medium">
                  {{ summary.totalWeight.toFixed(0) }}%
                  <span v-if="extraCredit > 0" class="text-slate-400">
                    (+{{ extraCredit.toFixed(0) }}% EC)
                  </span>
                </td>
                <td colspan="2"></td>
              </tr>
            </tfoot>
          </table>
        </div>

        <p
          v-if="summary.totalWeight > 0 && summary.totalWeight < 100"
          class="text-xs text-yellow-600 dark:text-yellow-400 mt-2"
        >
          Weights sum to {{ summary.totalWeight.toFixed(1) }}%. Add more categories or adjust weights for a complete grade.
        </p>
      </div>
    </div>

    <!-- Editor modal -->
    <GradeRubricEditor
      :open="editorOpen"
      :class-i-d="classID"
      :rubric="rubric"
      @close="editorOpen = false"
      @saved="onRubricSaved"
      @deleted="onRubricDeleted"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, h, defineComponent, onMounted, type PropType } from 'vue'
import { api, type GradeRubric, type GradeCategory } from '@/api'
import {
  computeGrade, categoryBreakdown,
  effectiveThresholds, letterGrade, letterBadgeClass
} from '@/composables/useGrades'
import GradeRubricEditor from './GradeRubricEditor.vue'

const props = defineProps<{ classID: number }>()

const loading = ref(true)
const rubric = ref<GradeRubric | null>(null)
const editorOpen = ref(false)

const summary = computed(() =>
  rubric.value
    ? computeGrade(rubric.value)
    : { current: null, projected: 0, best: 0, scoredWeight: 0, totalWeight: 0, hasAnyScored: false }
)

const extraCredit = computed(() =>
  rubric.value
    ? rubric.value.categories.filter(c => c.is_extra_credit).reduce((s, c) => s + c.weight, 0)
    : 0
)

onMounted(async () => {
  await fetchRubric()
})

async function fetchRubric() {
  loading.value = true
  try {
    rubric.value = await api.grades.getRubric(props.classID)
  } catch (e) {
    // 404 = no rubric yet, anything else is an actual error.
    rubric.value = null
    if (!(e instanceof Error) || !/no rubric/i.test(e.message)) {
      console.error('failed to fetch rubric', e)
    }
  } finally {
    loading.value = false
  }
}

function openEditor() {
  editorOpen.value = true
}

function onRubricSaved(updated: GradeRubric) {
  rubric.value = updated
}

function onRubricDeleted() {
  rubric.value = null
}

// --- Score editing ---

const saveTimers = new Map<number, number>()

async function updateScore(cat: GradeCategory, payload: { earned: number | null; possible: number }) {
  // Optimistic local update
  const idx = rubric.value?.categories.findIndex(c => c.id === cat.id) ?? -1
  if (idx >= 0 && rubric.value) {
    const updated = { ...rubric.value.categories[idx], ...payload }
    rubric.value.categories[idx] = updated
  }
  // Debounced save (300ms — typing in a number field shouldn't hammer the server)
  const existing = saveTimers.get(cat.id)
  if (existing) window.clearTimeout(existing)
  saveTimers.set(cat.id, window.setTimeout(async () => {
    try {
      const saved = await api.grades.updateCategory(cat.id, payload)
      if (rubric.value) {
        const i = rubric.value.categories.findIndex(c => c.id === cat.id)
        if (i >= 0) rubric.value.categories[i] = saved
      }
    } catch (e) {
      console.error('failed to save score', e)
    }
  }, 300))
}

// --- Per-row contribution display ---

function contribLabel(cat: GradeCategory): string {
  const b = categoryBreakdown(cat)
  if (b.contribution === null) return '—'
  return `${b.contribution.toFixed(1)} pts`
}

function contribClass(cat: GradeCategory): string {
  return cat.earned === null || cat.earned === undefined
    ? 'text-slate-300 dark:text-zinc-600'
    : 'text-slate-700 dark:text-zinc-200 font-medium'
}

// --- Inline subcomponents ---
// Kept inline because they are tightly coupled to this panel's rendering.

const GradeStat = defineComponent({
  props: {
    label: { type: String, required: true },
    rubric: { type: Object as PropType<GradeRubric>, required: true },
    pct: { type: Number as PropType<number | null>, default: null },
    primary: { type: Boolean, default: false }
  },
  setup(p) {
    const letter = computed(() => {
      if (p.pct === null || p.pct === undefined) return null
      return letterGrade(p.pct, effectiveThresholds(p.rubric))
    })
    return () =>
      h('div', {
        class: [
          'rounded-xl border p-4 transition-colors',
          p.primary
            ? 'border-slate-300 dark:border-zinc-700 bg-white dark:bg-[#1e1e1e]'
            : 'border-slate-200 dark:border-zinc-800 bg-slate-50 dark:bg-[#1e1e1e]/50'
        ]
      }, [
        h('p', { class: 'text-xs font-medium text-slate-500 dark:text-zinc-400 mb-2' }, p.label),
        p.pct === null
          ? h('div', { class: 'flex items-baseline gap-2' }, [
              h('span', { class: 'text-2xl font-semibold text-slate-300 dark:text-zinc-600' }, '—')
            ])
          : h('div', { class: 'flex items-baseline gap-2 flex-wrap' }, [
              letter.value
                ? h('span', {
                    class: [
                      'px-2.5 py-0.5 rounded-md font-bold',
                      p.primary ? 'text-2xl' : 'text-base',
                      letterBadgeClass(letter.value)
                    ]
                  }, letter.value)
                : null,
              h('span', {
                class: [
                  'tabular-nums text-slate-600 dark:text-zinc-300',
                  p.primary ? 'text-xl font-semibold' : 'text-base'
                ]
              }, `${p.pct.toFixed(1)}%`)
            ])
      ])
  }
})

const ScoreInput = defineComponent({
  props: {
    category: { type: Object as PropType<GradeCategory>, required: true }
  },
  emits: {
    change: (_p: { earned: number | null; possible: number }) => true
  },
  setup(p, { emit }) {
    // For possible=100 we show a single percentage input; otherwise two fields.
    const isPercentMode = computed(() => p.category.possible === 100)
    const earnedDraft = ref(p.category.earned !== null && p.category.earned !== undefined ? String(p.category.earned) : '')
    const possibleDraft = ref(String(p.category.possible))

    function commit() {
      const earnedRaw = earnedDraft.value.trim()
      let earned = earnedRaw === '' ? null : Number(earnedRaw)
      if (earned !== null && !Number.isFinite(earned)) return
      const possible = Number(possibleDraft.value) || 100
      // Clamp earned to [0, possible]
      if (earned !== null) {
        earned = Math.max(0, Math.min(possible, earned))
        earnedDraft.value = String(earned)
      }
      emit('change', { earned, possible })
    }

    const inputClass = 'text-right rounded border border-slate-200 dark:border-zinc-800 bg-white dark:bg-[#141414] text-slate-900 dark:text-zinc-100 px-2 py-1 text-sm tabular-nums focus:outline-none focus:ring-2 focus:ring-indigo-500'

    return () => h('div', { class: 'flex items-center gap-1 justify-end' }, [
      h('input', {
        type: 'number',
        step: '0.5',
        min: '0',
        max: String(p.category.possible),
        placeholder: '—',
        value: earnedDraft.value,
        class: 'w-16 ' + inputClass,
        onInput: (e: Event) => { earnedDraft.value = (e.target as HTMLInputElement).value },
        onChange: commit,
        onBlur: commit
      }),
      h('span', { class: 'text-slate-400 dark:text-zinc-500 text-xs' }, '/'),
      isPercentMode.value
        ? h('span', { class: 'text-slate-500 dark:text-zinc-400 text-xs w-9' }, '100')
        : h('input', {
            type: 'number',
            step: '1',
            min: '1',
            value: possibleDraft.value,
            class: 'w-12 ' + inputClass,
            onInput: (e: Event) => { possibleDraft.value = (e.target as HTMLInputElement).value },
            onChange: commit,
            onBlur: commit
          })
    ])
  }
})
</script>
