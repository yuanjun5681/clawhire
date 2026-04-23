<script setup lang="ts">
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { useIdentityStore } from '@/stores/identity'

const identity = useIdentityStore()
const router = useRouter()

const navItems = [
  { to: '/tasks', label: '任务大厅' },
  { to: '/my/tasks', label: '我的任务' },
]

function goToAccount() {
  router.push(`/accounts/${identity.currentAccountId}`)
}
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
          <button
            class="flex items-center gap-2 rounded-full border border-base-300 bg-base-100 px-3 py-1.5 text-sm hover:border-primary/40"
            @click="goToAccount"
          >
            <span
              class="grid h-6 w-6 place-items-center rounded-full bg-primary/10 text-xs font-medium text-primary"
            >{{ identity.displayName.slice(0, 1) }}</span>
            <span class="text-base-content">{{ identity.displayName }}</span>
            <span
              class="rounded-sm bg-base-200 px-1.5 py-0.5 text-[10px] uppercase text-base-content/60"
            >{{ identity.accountType }}</span>
          </button>
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
