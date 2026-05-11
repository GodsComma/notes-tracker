import { ref } from 'vue'

type Theme = 'light' | 'dark'

const theme = ref<Theme>('light')

function apply(t: Theme) {
  theme.value = t
  document.documentElement.classList.toggle('dark', t === 'dark')
}

export function useTheme() {
  function setTheme(t: Theme) {
    apply(t)
    localStorage.setItem('theme', t)
  }

  function toggle() {
    setTheme(theme.value === 'dark' ? 'light' : 'dark')
  }

  return { theme, toggle, setTheme }
}

export function initTheme() {
  const mq = window.matchMedia('(prefers-color-scheme: dark)')

  // Apply on load — saved preference wins, otherwise follow browser
  const saved = localStorage.getItem('theme') as Theme | null
  apply(saved ?? (mq.matches ? 'dark' : 'light'))

  // React to OS preference changes in real-time when the user
  // hasn't pinned a manual preference
  mq.addEventListener('change', (e) => {
    if (!localStorage.getItem('theme')) {
      apply(e.matches ? 'dark' : 'light')
    }
  })
}
