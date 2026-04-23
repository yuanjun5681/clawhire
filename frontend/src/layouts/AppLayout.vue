<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { useIdentityStore } from '@/stores/identity'

const identity = useIdentityStore()
const router = useRouter()

const navItems = [
  { to: '/tasks', label: '任务大厅' },
  { to: '/my/tasks', label: '我的任务' },
]

const menuOpen = ref(false)
const menuRef = ref<HTMLElement | null>(null)

function toggleMenu() {
  menuOpen.value = !menuOpen.value
}

function closeMenu() {
  menuOpen.value = false
}

function goToAccount() {
  closeMenu()
  router.push(`/accounts/${identity.currentAccountId}`)
}

function signOut() {
  closeMenu()
  identity.signOut()
  router.push('/login')
}

function handleClickOutside(e: MouseEvent) {
  if (!menuOpen.value) return
  const el = menuRef.value
  if (el && !el.contains(e.target as Node)) closeMenu()
}

function handleEscape(e: KeyboardEvent) {
  if (e.key === 'Escape') closeMenu()
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleEscape)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleEscape)
})
</script>

<template>
  <div class="flex min-h-screen flex-col bg-base-200">
    <header
      class="sticky top-0 z-20 border-b border-base-300 bg-base-100/90 backdrop-blur"
    >
      <div
        class="mx-auto flex h-14 w-full max-w-[1440px] items-center gap-6 px-6"
      >
        <RouterLink
          to="/tasks"
          class="flex items-center gap-2 font-semibold tracking-tight text-base-content"
        >
          <span
            class="grid h-7 w-7 place-items-center rounded-md bg-primary text-primary-content text-sm font-bold"
          >C</span>
          <span class="text-[15px]">ClawHire</span>
        </RouterLink>

        <nav class="flex items-center gap-1 text-sm">
          <RouterLink
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="rounded-md px-3 py-1.5 text-base-content/70 transition hover:bg-base-200 hover:text-base-content"
            active-class="!bg-base-200 !text-base-content font-medium"
          >
            {{ item.label }}
          </RouterLink>
        </nav>

        <div class="ml-auto flex items-center gap-3">
          <div ref="menuRef" class="relative">
            <button
              type="button"
              class="flex items-center gap-2 rounded-full border border-base-300 bg-base-100 px-3 py-1.5 text-sm hover:border-primary/40"
              :aria-expanded="menuOpen"
              aria-haspopup="menu"
              @click="toggleMenu"
            >
              <span
                class="grid h-6 w-6 place-items-center rounded-full bg-primary/10 text-xs font-medium text-primary"
              >{{ identity.displayName.slice(0, 1) || '—' }}</span>
              <span class="text-base-content">{{ identity.displayName }}</span>
              <svg
                class="h-3 w-3 text-base-content/50"
                :class="menuOpen ? 'rotate-180' : ''"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                aria-hidden="true"
              >
                <polyline points="6 9 12 15 18 9" />
              </svg>
            </button>

            <Transition
              enter-active-class="transition duration-100 ease-out"
              enter-from-class="opacity-0 -translate-y-1"
              leave-active-class="transition duration-75 ease-in"
              leave-to-class="opacity-0 -translate-y-1"
            >
              <div
                v-if="menuOpen"
                role="menu"
                class="absolute right-0 top-full z-30 mt-2 w-52 overflow-hidden rounded-lg border border-base-300 bg-base-100 shadow-lg"
              >
                <div
                  class="border-b border-base-200 px-3 py-2 text-xs text-base-content/60"
                >
                  <p class="truncate font-medium text-base-content">
                    {{ identity.displayName }}
                  </p>
                  <p class="truncate font-mono text-[11px]">
                    {{ identity.currentAccountId }}
                  </p>
                </div>
                <button
                  role="menuitem"
                  type="button"
                  class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-base-content hover:bg-base-200"
                  @click="goToAccount"
                >
                  <svg
                    class="h-4 w-4 text-base-content/60"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    aria-hidden="true"
                  >
                    <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
                    <circle cx="12" cy="7" r="4" />
                  </svg>
                  账号主页
                </button>
                <button
                  role="menuitem"
                  type="button"
                  class="flex w-full items-center gap-2 border-t border-base-200 px-3 py-2 text-left text-sm text-red-600 hover:bg-red-50"
                  @click="signOut"
                >
                  <svg
                    class="h-4 w-4"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    aria-hidden="true"
                  >
                    <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                    <polyline points="16 17 21 12 16 7" />
                    <line x1="21" y1="12" x2="9" y2="12" />
                  </svg>
                  退出登录
                </button>
              </div>
            </Transition>
          </div>
        </div>
      </div>
    </header>

    <main class="mx-auto w-full max-w-[1440px] flex-1 px-6 py-6">
      <RouterView />
    </main>

    <footer
      class="border-t border-base-300 bg-base-100 py-4 text-center text-xs text-base-content/50"
    >
      ClawHire · Web MVP
    </footer>
  </div>
</template>
