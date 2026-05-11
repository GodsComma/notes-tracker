<template>
  <BaseModal :open="open" :title="title" @close="$emit('cancel')">
    <div class="space-y-4">
      <p class="text-sm text-slate-600 dark:text-slate-300">{{ message }}</p>
      <div class="flex justify-end gap-2 pt-1">
        <button class="btn-ghost" @click="$emit('cancel')">Cancel</button>
        <button
          class="px-4 py-2 rounded-lg text-sm font-medium transition-colors text-white"
          :class="danger
            ? 'bg-red-600 hover:bg-red-700'
            : 'bg-indigo-600 hover:bg-indigo-700'"
          @click="$emit('confirm')"
        >{{ confirmLabel }}</button>
      </div>
    </div>
  </BaseModal>
</template>

<script setup lang="ts">
import BaseModal from './BaseModal.vue'

withDefaults(defineProps<{
  open: boolean
  title: string
  message: string
  confirmLabel?: string
  danger?: boolean
}>(), {
  confirmLabel: 'Delete',
  danger: true
})

defineEmits<{ confirm: []; cancel: [] }>()
</script>
