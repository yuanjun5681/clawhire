<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { useIdentityStore } from '@/stores/identity'
import { useToastStore } from '@/stores/toast'
import { ThemeToggle, UiAvatar } from '@/components/ui'

const identity = useIdentityStore()
const route = useRoute()
const router = useRouter()
const toast = useToastStore()

const navItems = [
  {
    to: '/tasks',
    label: '任务大厅',
    icon: 'M3 3h7v7H3z M14 3h7v7h-7z M14 14h7v7h-7z M3 14h7v7H3z',
  },
  {
    to: '/my/tasks',
    label: '我的任务',
    icon: 'M20 7h-3V5a2 2 0 0 0-2-2H9a2 2 0 0 0-2 2v2H4a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2z M9 5h6v2H9z',
  },
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
  toast.info('已退出登录')
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
  <div class="relative flex min-h-screen flex-col">
    <header
      class="sticky top-0 z-30 border-b border-base-300/60 bg-base-100/70 backdrop-blur-xl supports-[backdrop-filter]:bg-base-100/50"
    >
      <div
        class="mx-auto flex h-16 w-full max-w-[1440px] items-center gap-6 px-5 sm:px-8"
      >
        <RouterLink
          to="/tasks"
          class="group flex items-center gap-2.5 font-semibold tracking-tight text-base-content"
        >
          <span
            aria-hidden="true"
            class="relative grid h-9 w-9 place-items-center rounded-xl bg-[linear-gradient(120deg,var(--color-primary),var(--color-accent))] text-white shadow-[0_6px_20px_-8px_color-mix(in_oklch,var(--color-primary)_70%,transparent)] transition-transform group-hover:rotate-[-6deg]"
          >
            <svg
              class="h-4 w-4"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M8 4L4 8l4 4" />
              <path d="M4 8h14a4 4 0 0 1 4 4v0" />
              <path d="M16 20l4-4-4-4" />
              <path d="M20 16H6a4 4 0 0 1-4-4v0" />
            </svg>
          </span>
          <span class="text-[15px]">ClawHire</span>
        </RouterLink>

        <nav class="hidden items-center gap-1 text-sm sm:flex">
          <RouterLink
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="group relative inline-flex items-center gap-2 rounded-full px-3.5 py-1.5 text-base-content/65 transition hover:text-base-content"
            active-class="!text-primary"
          >
            <span
              v-if="route.path.startsWith(item.to)"
              aria-hidden="true"
              class="absolute inset-0 -z-10 rounded-full bg-primary/10 ring-1 ring-inset ring-primary/20"
            />
            <svg
              class="h-3.5 w-3.5"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1.75"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path :d="item.icon" />
            </svg>
            {{ item.label }}
          </RouterLink>
        </nav>

        <div class="ml-auto flex items-center gap-3">
          <ThemeToggle />

          <div ref="menuRef" class="relative">
            <button
              type="button"
              class="flex items-center gap-2.5 rounded-full border border-base-300/70 bg-base-100/80 py-1 pl-1 pr-3 text-sm transition hover:border-primary/40 hover:bg-base-100"
              :aria-expanded="menuOpen"
              aria-haspopup="menu"
              @click="toggleMenu"
            >
              <UiAvatar :name="identity.displayName" size="sm" ring />
              <span class="hidden max-w-[8rem] truncate text-base-content sm:block">{{ identity.displayName || '未登录' }}</span>
              <svg
                :class="[
                  'h-3 w-3 text-base-content/50 transition-transform duration-200',
                  menuOpen ? 'rotate-180' : '',
                ]"
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
              enter-active-class="transition duration-150 ease-out"
              enter-from-class="opacity-0 -translate-y-1 scale-95"
              leave-active-class="transition duration-100 ease-in"
              leave-to-class="opacity-0 -translate-y-1 scale-95"
            >
              <div
                v-if="menuOpen"
                role="menu"
                class="absolute right-0 top-full z-40 mt-2 w-64 origin-top-right overflow-hidden rounded-box glass shadow-[0_20px_60px_-20px_color-mix(in_oklch,var(--color-base-content)_30%,transparent)]"
              >
                <div class="flex items-center gap-3 border-b border-base-300/60 px-4 py-3">
                  <UiAvatar :name="identity.displayName" size="md" ring />
                  <div class="min-w-0 flex-1">
                    <p class="truncate text-sm font-semibold text-base-content">
                      {{ identity.displayName }}
                    </p>
                    <p class="truncate font-mono text-[11px] text-base-content/55">
                      {{ identity.currentAccountId }}
                    </p>
                  </div>
                  <span
                    class="rounded-full bg-primary/12 px-2 py-0.5 text-[10px] font-medium text-primary ring-1 ring-primary/20"
                  >
                    {{ identity.accountType === 'human' ? 'Human' : 'Agent' }}
                  </span>
                </div>
                <button
                  role="menuitem"
                  type="button"
                  class="flex w-full items-center gap-3 px-4 py-2.5 text-left text-sm text-base-content transition hover:bg-base-200/70"
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
                  class="flex w-full items-center gap-3 border-t border-base-300/60 px-4 py-2.5 text-left text-sm text-error transition hover:bg-error/10"
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

      <!-- mobile nav -->
      <nav
        class="flex items-center gap-1 overflow-x-auto border-t border-base-300/50 bg-base-100/40 px-4 py-2 text-sm sm:hidden"
      >
        <RouterLink
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="inline-flex shrink-0 items-center gap-1.5 rounded-full px-3 py-1 text-base-content/65"
          active-class="!bg-primary/10 !text-primary ring-1 ring-primary/20"
        >
          {{ item.label }}
        </RouterLink>
      </nav>
    </header>

    <main class="mx-auto w-full max-w-[1440px] flex-1 px-5 py-6 sm:px-8 sm:py-8">
      <RouterView v-slot="{ Component, route: r }">
        <Transition
          mode="out-in"
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="opacity-0 translate-y-1"
          leave-active-class="transition duration-100 ease-in"
          leave-to-class="opacity-0"
        >
          <component :is="Component" :key="r.fullPath" />
        </Transition>
      </RouterView>
    </main>

    <footer
      class="mt-auto border-t border-base-300/50 bg-base-100/40 py-5 text-center text-[11px] text-base-content/45"
    >
      ClawHire · 基于 ClawSynapse 的任务合约与结算层 · © 2026
    </footer>
  </div>
</template>
