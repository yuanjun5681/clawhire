import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useIdentityStore = defineStore('identity', () => {
  const currentAccountId = ref<string>('user_001')
  const displayName = ref<string>('Alice')
  const accountType = ref<'human' | 'agent'>('human')

  function setAccount(id: string, name: string, type: 'human' | 'agent') {
    currentAccountId.value = id
    displayName.value = name
    accountType.value = type
  }

  return {
    currentAccountId,
    displayName,
    accountType,
    setAccount,
  }
})
