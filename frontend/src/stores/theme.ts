import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'

export type ThemeMode = 'light' | 'dark' | 'auto'

const THEME_KEY = 'clawhire.theme'
const LIGHT = 'clawhire'
const DARK = 'clawhire-dark'

function loadInitial(): ThemeMode {
  try {
    const v = localStorage.getItem(THEME_KEY) as ThemeMode | null
    if (v === 'light' || v === 'dark' || v === 'auto') return v
  } catch { /* ignore */ }
  return 'auto'
}

function systemPrefersDark(): boolean {
  return typeof window !== 'undefined' &&
    window.matchMedia?.('(prefers-color-scheme: dark)').matches === true
}

export const useThemeStore = defineStore('theme', () => {
  const mode = ref<ThemeMode>(loadInitial())
  const systemDark = ref(systemPrefersDark())

  const resolved = computed<'light' | 'dark'>(() => {
    if (mode.value === 'auto') return systemDark.value ? 'dark' : 'light'
    return mode.value
  })

  const daisyTheme = computed(() => (resolved.value === 'dark' ? DARK : LIGHT))

  function apply() {
    if (typeof document === 'undefined') return
    document.documentElement.setAttribute('data-theme', daisyTheme.value)
    document.documentElement.style.colorScheme = resolved.value
  }

  function setMode(next: ThemeMode) {
    mode.value = next
    try {
      localStorage.setItem(THEME_KEY, next)
    } catch { /* ignore */ }
  }

  function cycle() {
    const next: ThemeMode =
      mode.value === 'light' ? 'dark' : mode.value === 'dark' ? 'auto' : 'light'
    setMode(next)
  }

  function bindSystemListener() {
    if (typeof window === 'undefined') return () => {}
    const mql = window.matchMedia('(prefers-color-scheme: dark)')
    const handler = (e: MediaQueryListEvent) => {
      systemDark.value = e.matches
    }
    mql.addEventListener?.('change', handler)
    return () => mql.removeEventListener?.('change', handler)
  }

  watch(resolved, apply, { immediate: true })

  return { mode, resolved, daisyTheme, setMode, cycle, bindSystemListener }
})
