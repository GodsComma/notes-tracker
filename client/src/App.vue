<template>
  <AppShell v-if="authStore.user" />
  <RouterView v-else />
</template>

<script setup lang="ts">
import AppShell from '@/components/AppShell.vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
</script>
