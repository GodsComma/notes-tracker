<template>
  <!-- Editable: dropdown trigger -->
  <div v-if="editable" class="relative inline-block" @keydown.esc="open = false">
    <button
      type="button"
      class="inline-flex items-center gap-1 rounded border transition-colors focus:outline-none"
      :class="[sizeClass, priority === 0 ? 'text-slate-400 dark:text-slate-500 border-transparent hover:border-slate-300 dark:hover:border-slate-600 px-1' : chipClass]"
      @click.stop="open = !open"
    >
      <span v-if="priority !== 0" :class="dotClass">●</span>
      <span>{{ label }}</span>
      <svg class="w-3 h-3 opacity-50 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
      </svg>
    </button>

    <div
      v-if="open"
      class="absolute z-20 mt-1 left-0 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg shadow-lg overflow-hidden min-w-[110px]"
    >
      <button
        v-for="opt in OPTIONS"
        :key="opt.value"
        type="button"
        class="w-full flex items-center gap-2 px-3 py-1.5 text-xs hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
        :class="opt.value === priority ? 'font-semibold' : ''"
        @click.stop="select(opt.value)"
      >
        <span :class="opt.dot">{{ opt.value === 0 ? '○' : '●' }}</span>
        <span class="text-slate-700 dark:text-slate-200">{{ opt.label }}</span>
      </button>
    </div>
  </div>

  <!-- Read-only: just the chip or nothing for None -->
  <span
    v-else-if="priority !== 0"
    class="inline-flex items-center gap-1 rounded border"
    :class="[sizeClass, chipClass]"
  >
    <span :class="dotClass">●</span>
    <span>{{ label }}</span>
  </span>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

const props = withDefaults(defineProps<{
  priority: number
  editable?: boolean
  size?: 'sm' | 'xs'
}>(), {
  editable: true,
  size: 'sm'
})

const emit = defineEmits<{ change: [priority: number] }>()

const open = ref(false)

const OPTIONS = [
  { value: 1, label: 'High',   dot: 'text-rose-500' },
  { value: 2, label: 'Medium', dot: 'text-amber-500' },
  { value: 3, label: 'Low',    dot: 'text-blue-400' },
  { value: 0, label: 'None',   dot: 'text-slate-400' }
]

const CONFIG: Record<number, { label: string; dot: string; chip: string }> = {
  1: { label: 'High',   dot: 'text-rose-500',  chip: 'bg-rose-50 dark:bg-rose-950/60 border-rose-200 dark:border-rose-800 text-rose-700 dark:text-rose-300' },
  2: { label: 'Medium', dot: 'text-amber-500', chip: 'bg-amber-50 dark:bg-amber-950/60 border-amber-200 dark:border-amber-800 text-amber-700 dark:text-amber-300' },
  3: { label: 'Low',    dot: 'text-blue-400',  chip: 'bg-blue-50 dark:bg-blue-950/60 border-blue-200 dark:border-blue-800 text-blue-700 dark:text-blue-300' },
  0: { label: 'None',   dot: 'text-slate-400', chip: '' }
}

const label    = computed(() => CONFIG[props.priority]?.label ?? 'None')
const dotClass = computed(() => CONFIG[props.priority]?.dot ?? '')
const chipClass = computed(() => CONFIG[props.priority]?.chip ?? '')
const sizeClass = computed(() => props.size === 'xs' ? 'text-[10px] px-1 py-0' : 'text-xs px-1.5 py-0.5')

function select(value: number) {
  open.value = false
  emit('change', value)
}

function onDocClick() { open.value = false }
onMounted(() => document.addEventListener('click', onDocClick))
onUnmounted(() => document.removeEventListener('click', onDocClick))
</script>
