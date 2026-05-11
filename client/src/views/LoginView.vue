<template>
  <div class="min-h-screen flex items-center justify-center bg-slate-50 dark:bg-slate-900 px-4">
    <div class="w-full max-w-sm">
      <div class="mb-8 text-center">
        <h1 class="text-2xl font-bold text-slate-900 dark:text-white tracking-tight">Class Notes</h1>
        <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">
          {{ mode === 'login' ? 'Sign in to your account' : 'Create an account' }}
        </p>
      </div>

      <form class="bg-white dark:bg-slate-800 rounded-xl shadow-sm border border-slate-200 dark:border-slate-700 p-6 space-y-4" @submit.prevent="submit">
        <div>
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Email</label>
          <input
            v-model="email"
            type="email"
            autocomplete="email"
            required
            class="input"
            placeholder="you@example.com"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">Password</label>
          <input
            v-model="password"
            type="password"
            :autocomplete="mode === 'login' ? 'current-password' : 'new-password'"
            required
            :minlength="mode === 'register' ? 8 : undefined"
            class="input"
            placeholder="••••••••"
          />
          <p v-if="mode === 'register'" class="mt-1 text-xs text-slate-400">At least 8 characters.</p>
        </div>

        <p v-if="error" class="text-sm text-red-600 dark:text-red-400">{{ error }}</p>

        <button
          type="submit"
          :disabled="loading"
          class="btn-primary w-full"
        >
          {{ loading ? '…' : mode === 'login' ? 'Sign in' : 'Create account' }}
        </button>
      </form>

      <p class="mt-4 text-center text-sm text-slate-500 dark:text-slate-400">
        {{ mode === 'login' ? "Don't have an account?" : 'Already have an account?' }}
        <button
          class="ml-1 text-indigo-600 dark:text-indigo-400 hover:underline font-medium"
          @click="toggleMode"
        >
          {{ mode === 'login' ? 'Create one' : 'Sign in' }}
        </button>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const mode = ref<'login' | 'register'>('login')
const email = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

function toggleMode() {
  mode.value = mode.value === 'login' ? 'register' : 'login'
  error.value = ''
}

async function submit() {
  error.value = ''
  loading.value = true
  try {
    if (mode.value === 'login') {
      await auth.login(email.value, password.value)
    } else {
      await auth.register(email.value, password.value)
    }
    router.push('/')
  } catch (e: any) {
    error.value = e.message ?? 'Something went wrong'
  } finally {
    loading.value = false
  }
}
</script>
