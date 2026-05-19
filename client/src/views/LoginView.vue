<template>
  <div
    class="min-h-screen flex relative"
    :class="isDark ? 'bg-[#141414]' : 'bg-violet-50'"
  >
    <!-- Theme toggle -->
    <button
      class="absolute top-4 right-4 z-20 w-9 h-9 flex items-center justify-center rounded-full border transition-colors"
      :class="isDark
        ? 'bg-zinc-800 border-zinc-700 text-zinc-300 hover:bg-zinc-700'
        : 'bg-white border-violet-200 text-slate-600 hover:bg-violet-100'"
      @click="toggle"
      :title="isDark ? 'Switch to light mode' : 'Switch to dark mode'"
    >
      <!-- Sun icon (shown in dark mode) -->
      <svg v-if="isDark" xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="4"/>
        <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41"/>
      </svg>
      <!-- Moon icon (shown in light mode) -->
      <svg v-else xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
      </svg>
    </button>

    <!-- Left branding panel — hidden on mobile -->
    <div
      class="hidden lg:flex flex-col justify-between w-[52%] p-14 select-none"
      :class="isDark ? 'bg-[#1a1a1a]' : 'bg-indigo-700'"
    >
      <div>
        <!-- Logo -->
        <div class="flex items-center gap-3 mb-14">
          <div
            class="w-9 h-9 rounded-xl flex items-center justify-center shrink-0"
            :class="isDark ? 'bg-amber-500' : 'bg-white/20'"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/>
              <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/>
            </svg>
          </div>
          <span
            class="font-semibold text-lg tracking-tight"
            :class="isDark ? 'text-amber-400' : 'text-white'"
          >Note Taker</span>
        </div>

        <h2 class="text-4xl font-bold text-white leading-tight tracking-tight">
          Your life,<br>organized.
        </h2>
        <p class="mt-5 text-base leading-relaxed max-w-xs" :class="isDark ? 'text-zinc-400' : 'text-indigo-200'">
          Track classes, annotate readings, and stay on top of every deadline — all in one place.
        </p>

        <ul class="mt-10 space-y-5">
          <li v-for="f in features" :key="f.label" class="flex items-start gap-3">
            <span
              class="mt-0.5 w-5 h-5 rounded-full flex items-center justify-center shrink-0"
              :class="isDark ? 'bg-amber-500/20' : 'bg-white/20'"
            >
              <svg class="w-3 h-3" :class="isDark ? 'text-amber-400' : 'text-white'" viewBox="0 0 12 12" fill="none">
                <path d="M2 6l3 3 5-5" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </span>
            <div>
              <p class="text-sm font-medium" :class="isDark ? 'text-zinc-200' : 'text-white'">{{ f.label }}</p>
              <p class="text-xs mt-0.5" :class="isDark ? 'text-zinc-500' : 'text-indigo-300'">{{ f.desc }}</p>
            </div>
          </li>
        </ul>
      </div>

      <p class="text-xs" :class="isDark ? 'text-zinc-700' : 'text-indigo-400'">
        © {{ new Date().getFullYear() }} Note Taker
      </p>
    </div>

    <!-- Right form panel -->
    <div class="flex flex-1 items-center justify-center px-6 py-12">
      <div class="w-full max-w-sm">

        <!-- Mobile brand mark -->
        <div class="flex lg:hidden flex-col items-center mb-10">
          <div
            class="w-12 h-12 rounded-2xl flex items-center justify-center mb-3 shadow-lg"
            :class="isDark ? 'bg-amber-500' : 'bg-indigo-600'"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/>
              <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/>
            </svg>
          </div>
          <h1
            class="text-xl font-bold tracking-tight"
            :class="isDark ? 'text-white' : 'text-slate-900'"
          >Note Taker</h1>
        </div>

        <!-- Heading -->
        <div class="mb-8">
          <h2
            class="text-2xl font-bold tracking-tight"
            :class="isDark ? 'text-white' : 'text-slate-900'"
          >
            {{ mode === 'login' ? 'Welcome back' : 'Create an account' }}
          </h2>
          <p class="mt-1.5 text-sm" :class="isDark ? 'text-zinc-500' : 'text-slate-500'">
            {{ mode === 'login' ? 'Sign in to continue to Note Taker' : 'Get started — it only takes a moment' }}
          </p>
        </div>

        <!-- Form -->
        <form class="space-y-4" @submit.prevent="submit">
          <div>
            <label
              class="block text-sm font-medium mb-1.5"
              :class="isDark ? 'text-zinc-300' : 'text-slate-700'"
            >
              Email address
            </label>
            <input
              v-model="email"
              type="email"
              autocomplete="email"
              required
              class="input"
              :class="isDark ? '!bg-[#1e1e1e] !border-zinc-700 !text-zinc-100 !placeholder-zinc-600' : ''"
              placeholder="you@example.com"
            />
          </div>

          <div>
            <label
              class="block text-sm font-medium mb-1.5"
              :class="isDark ? 'text-zinc-300' : 'text-slate-700'"
            >
              Password
            </label>
            <input
              v-model="password"
              type="password"
              :autocomplete="mode === 'login' ? 'current-password' : 'new-password'"
              required
              :minlength="mode === 'register' ? 8 : undefined"
              class="input"
              :class="isDark ? '!bg-[#1e1e1e] !border-zinc-700 !text-zinc-100 !placeholder-zinc-600' : ''"
              placeholder="••••••••"
            />
            <p v-if="mode === 'register'" class="mt-1.5 text-xs" :class="isDark ? 'text-zinc-500' : 'text-slate-400'">
              At least 8 characters.
            </p>
          </div>

          <p
            v-if="error"
            class="text-sm rounded-lg px-3 py-2 border"
            :class="isDark
              ? 'text-red-400 bg-red-950/40 border-red-900'
              : 'text-red-600 bg-red-50 border-red-200'"
          >
            {{ error }}
          </p>

          <!-- Submit button — indigo in light, amber in dark -->
          <button
            type="submit"
            :disabled="loading"
            class="w-full mt-1 py-2.5 text-sm font-semibold rounded-lg transition-colors disabled:opacity-50"
            :class="isDark
              ? 'bg-amber-500 hover:bg-amber-400 text-black'
              : 'bg-indigo-600 hover:bg-indigo-700 text-white'"
          >
            {{ loading ? '…' : mode === 'login' ? 'Sign in' : 'Create account' }}
          </button>
        </form>

        <!-- <p class="mt-6 text-center text-sm" :class="isDark ? 'text-zinc-500' : 'text-slate-500'">
          {{ mode === 'login' ? "Don't have an account?" : 'Already have an account?' }}
          <button
            class="ml-1 font-medium hover:underline"
            :class="isDark ? 'text-amber-400' : 'text-indigo-600'"
            @click="toggleMode"
          >
            {{ mode === 'login' ? 'Create one' : 'Sign in' }}
          </button>
        </p> -->
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/composables/useTheme'

const router = useRouter()
const auth = useAuthStore()
const { theme, toggle } = useTheme()

const isDark = computed(() => theme.value === 'dark')

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
