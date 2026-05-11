<template>
  <div
    class="flex items-center gap-0.5 rounded-full p-0.5 bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700"
    role="group"
    aria-label="Color theme"
  >
    <!-- Light -->
    <button
      :class="[
        'flex items-center justify-center w-7 h-7 rounded-full transition-all duration-200',
        theme === 'light'
          ? 'bg-white dark:bg-slate-600 shadow-sm text-amber-500'
          : 'text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-400'
      ]"
      title="Light mode"
      @click="setTheme('light')"
    >
      <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364-6.364l-.707.707M6.343 17.657l-.707.707M17.657 17.657l-.707-.707M6.343 6.343l-.707-.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
      </svg>
    </button>

    <!-- Dark -->
    <button
      :class="[
        'flex items-center justify-center w-7 h-7 rounded-full transition-all duration-200',
        theme === 'dark'
          ? 'bg-slate-700 shadow-sm text-indigo-400'
          : 'text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-400'
      ]"
      title="Dark mode"
      @click="setTheme('dark')"
    >
      <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
        <path d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useTheme } from '@/composables/useTheme'

const { theme, setTheme } = useTheme()
</script>
