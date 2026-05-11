<template>
  <BaseModal :open="open" title="Start Focus Session" @close="$emit('close')">
    <div class="space-y-5">
      <!-- File header -->
      <div class="flex items-center gap-3 -mt-2">
        <div class="w-2 h-10 rounded shrink-0" :style="{ backgroundColor: item.class_color || '#6366f1' }" />
        <div class="min-w-0">
          <p class="font-medium text-slate-900 dark:text-slate-100 truncate">{{ item.filename }}</p>
          <p class="text-xs text-slate-500 dark:text-slate-400 truncate">
            {{ item.class_name }}<span v-if="item.source_title"> · {{ item.source_title }}</span>
          </p>
        </div>
      </div>

      <!-- Timer -->
      <div>
        <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">Session length</label>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="m in TIMER_PRESETS"
            :key="m"
            type="button"
            class="px-3 py-1.5 text-sm rounded-md border transition-colors"
            :class="settings.durationMin === m
              ? 'bg-indigo-600 border-indigo-600 text-white'
              : 'bg-white dark:bg-slate-700 border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-300 hover:border-indigo-400'"
            @click="setDuration(m)"
          >
            {{ m }}m
          </button>
          <button
            type="button"
            class="px-3 py-1.5 text-sm rounded-md border transition-colors"
            :class="customMode
              ? 'bg-indigo-600 border-indigo-600 text-white'
              : 'bg-white dark:bg-slate-700 border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-300 hover:border-indigo-400'"
            @click="customMode = true"
          >
            Custom
          </button>
        </div>
        <input
          v-if="customMode"
          v-model.number="settings.durationMin"
          type="number"
          min="1"
          max="240"
          class="input mt-2 w-32 text-sm"
          placeholder="minutes"
        />
      </div>

      <!-- Theme -->
      <div>
        <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">Reading theme</label>
        <div class="grid grid-cols-2 gap-2">
          <button
            v-for="t in THEMES"
            :key="t.id"
            type="button"
            class="flex items-center gap-2 px-3 py-2 text-sm rounded-md border transition-all"
            :class="settings.theme === t.id
              ? 'border-indigo-500 ring-2 ring-indigo-200 dark:ring-indigo-800'
              : 'border-slate-200 dark:border-slate-600 hover:border-indigo-400'"
            :style="{ backgroundColor: t.bg, color: t.fg }"
            @click="settings.theme = t.id"
          >
            <span class="font-medium">Aa</span>
            <span class="ml-auto text-xs opacity-80">{{ t.label }}</span>
          </button>
        </div>
      </div>

      <!-- Reading ruler -->
      <div>
        <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">Reading ruler</label>
        <div class="flex gap-2">
          <button
            v-for="opt in RULER_OPTIONS"
            :key="opt.id"
            type="button"
            class="flex-1 px-3 py-2 text-xs rounded-md border transition-colors"
            :class="settings.ruler === opt.id
              ? 'bg-indigo-600 border-indigo-600 text-white'
              : 'bg-white dark:bg-slate-700 border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-300 hover:border-indigo-400'"
            @click="settings.ruler = opt.id"
          >
            <div class="font-medium">{{ opt.label }}</div>
            <div class="text-[10px] opacity-80 mt-0.5">{{ opt.hint }}</div>
          </button>
        </div>
      </div>

      <!-- Annotations -->
      <div>
        <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">Annotations</label>
        <div class="flex gap-2">
          <button
            type="button"
            class="flex-1 px-3 py-2 text-xs rounded-md border transition-colors"
            :class="settings.annotations === 'on' && annotationsAvailable
              ? 'bg-indigo-600 border-indigo-600 text-white'
              : 'bg-white dark:bg-slate-700 border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-300'"
            :disabled="!annotationsAvailable"
            @click="settings.annotations = 'on'"
          >
            <div class="font-medium">On</div>
            <div class="text-[10px] opacity-80 mt-0.5">Capture highlights & notes</div>
          </button>
          <button
            type="button"
            class="flex-1 px-3 py-2 text-xs rounded-md border transition-colors"
            :class="settings.annotations === 'off' || !annotationsAvailable
              ? 'bg-indigo-600 border-indigo-600 text-white'
              : 'bg-white dark:bg-slate-700 border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-300 hover:border-indigo-400'"
            @click="settings.annotations = 'off'"
          >
            <div class="font-medium">Off</div>
            <div class="text-[10px] opacity-80 mt-0.5">Pure reading; no tooltip</div>
          </button>
        </div>
        <p v-if="!annotationsAvailable" class="text-xs text-slate-500 dark:text-slate-400 mt-1.5">
          Attach this PDF to a source to enable notes.
        </p>
        <label v-else class="flex items-center gap-2 text-xs text-slate-600 dark:text-slate-400 mt-2 cursor-pointer">
          <input
            v-model="settings.autoHideDrawer"
            type="checkbox"
            class="rounded border-slate-300 dark:border-slate-600"
            :disabled="settings.annotations === 'off'"
          />
          Auto-hide notes drawer after 30s of inactivity
        </label>
      </div>

      <!-- Notification suppression -->
      <label class="flex items-start gap-2.5 text-sm cursor-pointer">
        <input
          v-model="settings.suppressNotifications"
          type="checkbox"
          class="mt-0.5 rounded border-slate-300 dark:border-slate-600"
        />
        <span class="text-slate-700 dark:text-slate-300">
          Block in-app notifications
          <span class="block text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            For full peace, also enable Do Not Disturb on your system.
          </span>
        </span>
      </label>

      <!-- Resume hint -->
      <div
        v-if="resumePage"
        class="text-xs bg-indigo-50 dark:bg-indigo-950/40 text-indigo-700 dark:text-indigo-300 rounded-md px-3 py-2"
      >
        Resuming from page {{ resumePage }}.
      </div>

      <!-- Actions -->
      <div class="flex justify-end gap-2 pt-1">
        <button type="button" class="btn-ghost" @click="$emit('close')">Cancel</button>
        <button type="button" class="btn-primary" @click="onStart">
          Start Session
        </button>
      </div>
    </div>
  </BaseModal>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import BaseModal from './BaseModal.vue'
import {
  loadSettings, saveSettings, loadProgress,
  THEMES, TIMER_PRESETS,
  type DeepReadSettings, type ReadingRulerMode
} from '@/composables/useDeepRead'
import type { PDFLibraryItem } from '@/api'

const props = defineProps<{ open: boolean; item: PDFLibraryItem }>()
const emit = defineEmits<{ close: []; start: [settings: DeepReadSettings] }>()

const settings = ref<DeepReadSettings>(loadSettings())
const customMode = ref(!TIMER_PRESETS.includes(settings.value.durationMin as 10 | 15 | 25 | 50))
const resumePage = ref<number | null>(null)

const annotationsAvailable = computed(() => props.item.source_id !== null)

const RULER_OPTIONS: { id: ReadingRulerMode; label: string; hint: string }[] = [
  { id: 'off',       label: 'Off',       hint: 'No overlay' },
  { id: 'underline', label: 'Underline', hint: 'Line at cursor' },
  { id: 'masking',   label: 'Masking',   hint: '3-line window' }
]

watch(() => props.open, open => {
  if (!open) return
  settings.value = loadSettings()
  // Force annotations off when no source is attached
  if (!annotationsAvailable.value) settings.value.annotations = 'off'
  customMode.value = !TIMER_PRESETS.includes(settings.value.durationMin as 10 | 15 | 25 | 50)
  const p = loadProgress(props.item.id)
  resumePage.value = p && p.lastPage > 1 ? p.lastPage : null
})

function setDuration(m: number) {
  settings.value.durationMin = m
  customMode.value = false
}

function onStart() {
  const s: DeepReadSettings = {
    ...settings.value,
    durationMin: Math.max(1, Math.min(240, settings.value.durationMin || 25))
  }
  saveSettings(s)
  emit('start', s)
}
</script>
