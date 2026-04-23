import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

const SESSION_KEY = 'clawhire.session'

export interface SessionSnapshot {
  accountId: string
  displayName: string
  accountType: 'human' | 'agent'
}

function loadSession(): SessionSnapshot | null {
  try {
    const raw = localStorage.getItem(SESSION_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as SessionSnapshot
    if (!parsed?.accountId || !parsed.displayName) return null
    return parsed
  } catch {
    return null
  }
}

function saveSession(s: SessionSnapshot | null) {
  if (s) localStorage.setItem(SESSION_KEY, JSON.stringify(s))
  else localStorage.removeItem(SESSION_KEY)
}

export const useIdentityStore = defineStore('identity', () => {
  const initial = loadSession()
  const currentAccountId = ref<string>(initial?.accountId ?? '')
  const displayName = ref<string>(initial?.displayName ?? '')
  const accountType = ref<'human' | 'agent'>(initial?.accountType ?? 'human')

  const isLoggedIn = computed(() => Boolean(currentAccountId.value))

  function signIn(snapshot: SessionSnapshot) {
    currentAccountId.value = snapshot.accountId
    displayName.value = snapshot.displayName
    accountType.value = snapshot.accountType
    saveSession(snapshot)
  }

  function signOut() {
    currentAccountId.value = ''
    displayName.value = ''
    accountType.value = 'human'
    saveSession(null)
  }

  return {
    currentAccountId,
    displayName,
    accountType,
    isLoggedIn,
    signIn,
    signOut,
  }
})
