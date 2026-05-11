import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useRouter } from 'vue-router'

export interface AuthUser {
  id: number
  email: string
  created_at: string
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AuthUser | null>(null)
  const status = ref<'idle' | 'loading' | 'error'>('idle')
  const router = useRouter()

  async function init() {
    status.value = 'loading'
    try {
      const res = await fetch('/api/v1/auth/me', { credentials: 'include' })
      if (res.ok) {
        user.value = await res.json()
      } else {
        user.value = null
      }
    } catch {
      user.value = null
    } finally {
      status.value = 'idle'
    }
  }

  async function login(email: string, password: string) {
    const res = await fetch('/api/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ email, password })
    })
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: res.statusText }))
      throw new Error(err.error ?? res.statusText)
    }
    user.value = await res.json()
  }

  async function register(email: string, password: string) {
    const res = await fetch('/api/v1/auth/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ email, password })
    })
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: res.statusText }))
      throw new Error(err.error ?? res.statusText)
    }
    user.value = await res.json()
  }

  async function logout() {
    await fetch('/api/v1/auth/logout', { method: 'POST', credentials: 'include' }).catch(() => {})
    user.value = null
    router.push('/login')
  }

  return { user, status, init, login, register, logout }
})
