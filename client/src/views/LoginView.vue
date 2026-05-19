<template>
  <div class="min-h-screen flex bg-white dark:bg-slate-950">

    <!-- Left branding panel — hidden on mobile -->
    <div class="hidden lg:flex flex-col justify-between w-[52%] bg-slate-900 p-14 select-none">
      <div>
        <!-- Logo -->
        <div class="flex items-center gap-3 mb-14">
          <div class="w-9 h-9 rounded-xl bg-indigo-500 flex items-center justify-center shrink-0">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/>
              <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/>
            </svg>
          </div>
          <span class="text-white font-semibold text-lg tracking-tight">Note Taker</span>
        </div>

        <h2 class="text-4xl font-bold text-white leading-tight tracking-tight">
          Your academic life,<br>organized.
        </h2>
        <p class="mt-5 text-slate-400 text-base leading-relaxed max-w-xs">
          Track classes, annotate readings, and stay on top of every deadline — all in one place.
        </p>

        <ul class="mt-10 space-y-4">
          <li v-for="f in features" :key="f.label" class="flex items-start gap-3">
            <span class="mt-0.5 w-5 h-5 rounded-full bg-indigo-500/20 flex items-center justify-center shrink-0">
              <svg class="w-3 h-3 text-indigo-400" viewBox="0 0 12 12" fill="none">
                <path d="M2 6l3 3 5-5" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </span>
            <div>
              <p class="text-sm font-medium text-slate-200">{{ f.label }}</p>
              <p class="text-xs text-slate-500 mt-0.5">{{ f.desc }}</p>
            </div>
          </li>
        </ul>
      </div>

      <p class="text-slate-600 text-xs">© {{ new Date().getFullYear() }} Note Taker</p>
    </div>

    <!-- Right form panel -->
    <div class="flex flex-1 items-center justify-center px-6 py-12">
      <div class="w-full max-w-sm">

        <!-- Mobile brand mark -->
        <div class="flex lg:hidden flex-col items-center mb-10">
          <div class="w-12 h-12 rounded-2xl bg-indigo-600 flex items-center justify-center mb-3 shadow-lg">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/>
              <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/>
            </svg>
          </div>
          <h1 class="text-xl font-bold text-slate-900 dark:text-white tracking-tight">Note Taker</h1>
        </div>

        <!-- Heading -->
        <div class="mb-8">
          <h2 class="text-2xl font-bold text-slate-900 dark:text-white tracking-tight">
            {{ mode === 'login' ? 'Welcome back' : 'Create an account' }}
          </h2>
          <p class="mt-1.5 text-sm text-slate-500 dark:text-slate-400">
            {{ mode === 'login' ? 'Sign in to continue to Note Taker' : 'Get started — it only takes a moment' }}
          </p>
        </div>

        <!-- Form -->
        <form class="space-y-4" @submit.prevent="submit">
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">
              Email address
            </label>
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
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">
              Password
            </label>
            <input
              v-model="password"
              type="password"
              :autocomplete="mode === 'login' ? 'current-password' : 'new-password'"
              required
              :minlength="mode === 'register' ? 8 : undefined"
              class="input"
              placeholder="••••••••"
            />
            <p v-if="mode === 'register'" class="mt-1.5 text-xs text-slate-400">
              At least 8 characters.
            </p>
          </div>

          <p v-if="error" class="text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-950/40 border border-red-200 dark:border-red-900 rounded-lg px-3 py-2">
            {{ error }}
          </p>

          <button
            type="submit"
            :disabled="loading"
            class="btn-primary w-full py-2.5 mt-1"
          >
            {{ loading ? '…' : mode === 'login' ? 'Sign in' : 'Create account' }}
          </button>
        </form>

        <p class="mt-6 text-center text-sm text-slate-500 dark:text-slate-400">
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

const features = [
  { label: 'Semester & class management', desc: 'Organise every course by term with a shared calendar.' },
  { label: 'PDF deep-read sessions', desc: 'Distraction-free reading with built-in focus tracking.' },
  { label: 'Deadline tracking', desc: 'Due dates surfaced in your weekly and upcoming panels.' },
]

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
