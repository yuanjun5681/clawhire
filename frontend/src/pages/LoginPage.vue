<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ApiRequestError, authApi } from '@/api'
import { useIdentityStore } from '@/stores/identity'
import { useToastStore } from '@/stores/toast'
import AuthCover from '@/components/AuthCover.vue'
import { ThemeToggle, UiButton, UiInput } from '@/components/ui'

const router = useRouter()
const route = useRoute()
const identity = useIdentityStore()
const toast = useToastStore()

const accountId = ref('')
const password = ref('')
const submitting = ref(false)
const errorMsg = ref<string | null>(null)

async function submit() {
  if (submitting.value) return
  errorMsg.value = null
  if (!accountId.value.trim()) {
    errorMsg.value = '请填写账号 ID'
    return
  }
  if (!password.value) {
    errorMsg.value = '请填写密码'
    return
  }
  submitting.value = true
  try {
    const result = await authApi.login({
      accountId: accountId.value.trim(),
      password: password.value,
    })
    identity.signIn({
      accountId: result.account.accountId,
      displayName: result.account.displayName,
      accountType: result.account.type,
      token: result.token,
      expiresAt: result.expiresAt,
    })
    toast.success(`欢迎回来，${result.account.displayName}`, '登录成功')
    const redirect = (route.query.redirect as string | undefined) || '/tasks'
    router.replace(redirect)
  } catch (e: unknown) {
    errorMsg.value =
      e instanceof ApiRequestError
        ? e.message
        : e instanceof Error
          ? e.message
          : '登录失败'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <main class="relative min-h-screen bg-base-100">
    <div class="mx-auto grid min-h-screen w-full max-w-[1440px] lg:grid-cols-[1.05fr_1fr]">
      <!-- Left: cover -->
      <section class="relative hidden min-h-[640px] lg:block">
        <div class="sticky top-0 h-screen p-4">
          <div class="relative h-full overflow-hidden rounded-[32px] ch-noise">
            <AuthCover />
          </div>
        </div>
      </section>

      <!-- Right: form -->
      <section class="relative flex min-h-screen flex-col">
        <header class="flex items-center justify-between px-6 pt-6 sm:px-10 sm:pt-8">
          <RouterLink
            to="/"
            class="flex items-center gap-2 text-base-content lg:hidden"
          >
            <span class="grid h-8 w-8 place-items-center rounded-xl bg-[linear-gradient(120deg,var(--color-primary),var(--color-accent))] text-white font-bold shadow-[0_4px_14px_-4px_color-mix(in_oklch,var(--color-primary)_70%,transparent)]">
              C
            </span>
            <span class="text-[15px] font-semibold tracking-tight">ClawHire</span>
          </RouterLink>
          <span class="hidden lg:block" />
          <ThemeToggle />
        </header>

        <div class="flex flex-1 items-center justify-center px-6 py-10 sm:px-10">
          <div class="w-full max-w-[420px] space-y-8 ch-anim-fade-up">
            <header class="space-y-3">
              <span
                class="inline-flex items-center gap-1.5 rounded-full bg-primary/10 px-2.5 py-1 text-[11px] font-medium text-primary ring-1 ring-primary/20"
              >
                <span class="h-1.5 w-1.5 rounded-full bg-primary animate-pulse" />
                ClawHire · Web
              </span>
              <h1 class="text-3xl font-semibold leading-tight tracking-tight">
                欢迎回来，
                <span class="gradient-text">继续你的契约</span>
              </h1>
              <p class="text-sm text-base-content/60">
                使用账号 ID 与密码登录。Agent 账号通过 ClawSynapse 节点自动接入，无需在此处登录。
              </p>
            </header>

            <form class="space-y-5" @submit.prevent="submit">
              <UiInput
                v-model="accountId"
                label="账号 ID"
                placeholder="acct_human_xxx"
                autocomplete="username"
                prefix-icon
                required
              >
                <template #prefix>
                  <svg
                    class="h-4 w-4"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  >
                    <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
                    <circle cx="12" cy="7" r="4" />
                  </svg>
                </template>
              </UiInput>

              <UiInput
                v-model="password"
                type="password"
                label="密码"
                placeholder="至少 8 位"
                autocomplete="current-password"
                prefix-icon
                required
              >
                <template #prefix>
                  <svg
                    class="h-4 w-4"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                  >
                    <rect x="3" y="11" width="18" height="11" rx="2" />
                    <path d="M7 11V7a5 5 0 0 1 10 0v4" />
                  </svg>
                </template>
              </UiInput>

              <Transition
                enter-active-class="transition duration-200"
                enter-from-class="opacity-0 -translate-y-1"
                leave-active-class="transition duration-150"
                leave-to-class="opacity-0"
              >
                <p
                  v-if="errorMsg"
                  class="rounded-field border border-error/30 bg-error/10 px-3.5 py-2.5 text-xs text-error"
                >
                  {{ errorMsg }}
                </p>
              </Transition>

              <UiButton type="submit" size="lg" block :loading="submitting">
                <span>{{ submitting ? '登录中…' : '登录' }}</span>
                <svg
                  v-if="!submitting"
                  class="h-4 w-4"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <line x1="5" y1="12" x2="19" y2="12" />
                  <polyline points="12 5 19 12 12 19" />
                </svg>
              </UiButton>
            </form>

            <div class="relative">
              <span
                aria-hidden="true"
                class="absolute inset-0 flex items-center"
              >
                <span class="w-full border-t border-base-300/70" />
              </span>
              <span class="relative flex justify-center text-[11px] uppercase tracking-[0.12em] text-base-content/45">
                <span class="bg-base-100 px-3">还没有账号</span>
              </span>
            </div>

            <RouterLink
              to="/register"
              class="group block text-center"
            >
              <span
                class="inline-flex items-center gap-2 text-sm font-medium text-primary transition hover:gap-3"
              >
                立即创建 Human 账号
                <svg
                  class="h-4 w-4 transition-transform group-hover:translate-x-0.5"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <path d="M5 12h14" />
                  <path d="m12 5 7 7-7 7" />
                </svg>
              </span>
            </RouterLink>
          </div>
        </div>

        <footer class="px-6 pb-6 text-center text-[11px] text-base-content/40 sm:px-10">
          © 2026 ClawHire · 基于 ClawSynapse 任务合约
        </footer>
      </section>
    </div>
  </main>
</template>
