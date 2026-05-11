import type { GradeCategory, GradeRubric } from '@/api'

export const DEFAULT_THRESHOLDS = {
  letter: { A: 90, B: 80, C: 70, D: 60, F: 0 } as Record<string, number>,
  letter_plus_minus: {
    'A+': 97, 'A': 93, 'A-': 90,
    'B+': 87, 'B': 83, 'B-': 80,
    'C+': 77, 'C': 73, 'C-': 70,
    'D+': 67, 'D': 63, 'D-': 60,
    'F': 0
  } as Record<string, number>,
  pass_fail: { P: 60, F: 0 } as Record<string, number>
}

export function defaultThresholdsFor(
  gradingType: 'letter' | 'pass_fail',
  plusMinus: boolean
): Record<string, number> {
  if (gradingType === 'pass_fail') return DEFAULT_THRESHOLDS.pass_fail
  return plusMinus ? DEFAULT_THRESHOLDS.letter_plus_minus : DEFAULT_THRESHOLDS.letter
}

export function effectiveThresholds(rubric: GradeRubric): Record<string, number> {
  const defaults = defaultThresholdsFor(rubric.grading_type, rubric.plus_minus)
  if (!rubric.thresholds) return defaults
  try {
    const custom = JSON.parse(rubric.thresholds) as Record<string, number>
    return { ...defaults, ...custom }
  } catch {
    return defaults
  }
}

export function letterGrade(pct: number, thresholds: Record<string, number>): string {
  const sorted = Object.entries(thresholds).sort((a, b) => b[1] - a[1])
  return sorted.find(([, min]) => pct >= min)?.[0] ?? 'F'
}

export interface GradeBreakdown {
  /** Earned percentage in [0, cat.weight], or null if not yet scored. */
  contribution: number | null
  /** Raw percentage on the work itself (e.g. 87 for 87/100). */
  pctEarned: number | null
}

export function categoryBreakdown(c: GradeCategory): GradeBreakdown {
  if (c.earned === null || c.earned === undefined) {
    return { contribution: null, pctEarned: null }
  }
  const pct = c.possible > 0 ? (c.earned / c.possible) * 100 : 0
  return {
    pctEarned: pct,
    contribution: (pct * c.weight) / 100
  }
}

export interface GradeSummary {
  current: number | null   // rescaled to scored weight (or null if nothing scored)
  projected: number        // assumes 0 on ungraded
  best: number             // assumes 100% on ungraded non-EC
  scoredWeight: number     // sum of weights for scored, non-EC categories
  totalWeight: number      // sum of weights for all non-EC categories
  hasAnyScored: boolean
}

export function computeGrade(rubric: GradeRubric): GradeSummary {
  let scoredWeight = 0
  let totalWeight = 0
  let weightedEarned = 0
  let bestUnscoredAdditional = 0
  let hasAnyScored = false

  for (const c of rubric.categories) {
    const contrib = categoryBreakdown(c).contribution
    if (!c.is_extra_credit) totalWeight += c.weight

    if (contrib !== null) {
      hasAnyScored = true
      weightedEarned += contrib
      if (!c.is_extra_credit) scoredWeight += c.weight
    } else if (!c.is_extra_credit) {
      bestUnscoredAdditional += c.weight
    }
  }

  const current = scoredWeight > 0
    ? (weightedEarned / scoredWeight) * 100
    : null

  return {
    current,
    projected: weightedEarned,
    best: weightedEarned + bestUnscoredAdditional,
    scoredWeight,
    totalWeight,
    hasAnyScored
  }
}

/** Tailwind class string for a colored badge for the given letter grade. */
export function letterBadgeClass(letter: string): string {
  const first = letter.charAt(0).toUpperCase()
  switch (first) {
    case 'A':
    case 'P':
      return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
    case 'B':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
    case 'C':
      return 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200'
    case 'D':
      return 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200'
    case 'F':
    default:
      return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
  }
}
