import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

export const SESSION_KEY = 'clawhire.session'

export interface SessionSnapshot {
  accountId: string
  displayName: string
  accountType: 'human' | 'agent'
  token: string
  expiresAt?: string
}

export function loadSessionSnapshot(): SessionSnapshot | null {
  try {
    const raw = localStorage.getItem(SESSION_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as SessionSnapshot
    if (!parsed?.accountId || !parsed.displayName || !parsed.token) return null
    if (parsed.expiresAt && Date.parse(parsed.expiresAt) <= Date.now()) {
      localStorage.removeItem(SESSION_KEY)
      return null
    }
    return parsed
  } catch {
    return null
  }
}

export function saveSessionSnapshot(s: SessionSnapshot | null) {
  if (s) localStorage.setItem(SESSION_KEY, JSON.stringify(s))
  else localStorage.removeItem(SESSION_KEY)
}

export const useIdentityStore = defineStore('identity', () => {
  const initial = loadSessionSnapshot()
  const currentAccountId = ref<string>(initial?.accountId ?? '')
  const displayName = ref<string>(initial?.displayName ?? '')
  const accountType = ref<'human' | 'agent'>(initial?.accountType ?? 'human')
  const token = ref<string>(initial?.token ?? '')
  const expiresAt = ref<string>(initial?.expiresAt ?? '')

  const isLoggedIn = computed(() => Boolean(currentAccountId.value && token.value))

  function signIn(snapshot: SessionSnapshot) {
    currentAccountId.value = snapshot.accountId
    displayName.value = snapshot.displayName
    accountType.value = snapshot.accountType
    token.value = snapshot.token
    expiresAt.value = snapshot.expiresAt ?? ''
    saveSessionSnapshot(snapshot)
  }

  function signOut() {
    currentAccountId.value = ''
    displayName.value = ''
    accountType.value = 'human'
    token.value = ''
    expiresAt.value = ''
    saveSessionSnapshot(null)
  }

  return {
    currentAccountId,
    displayName,
    accountType,
    token,
    expiresAt,
    isLoggedIn,
    signIn,
    signOut,
  }
})
