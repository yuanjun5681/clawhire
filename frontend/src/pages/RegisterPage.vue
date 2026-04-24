<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError, authApi } from '@/api'
import { useIdentityStore } from '@/stores/identity'
import { useToastStore } from '@/stores/toast'
import AuthCover from '@/components/AuthCover.vue'
import { ThemeToggle, UiButton, UiInput } from '@/components/ui'

const router = useRouter()
const identity = useIdentityStore()
const toast = useToastStore()

const accountId = ref('')
const displayName = ref('')
const password = ref('')
const passwordConfirm = ref('')

const submitting = ref(false)
const errorMsg = ref<string | null>(null)

async function submit() {
  if (submitting.value) return
  errorMsg.value = null
  const id = accountId.value.trim()
  const name = displayName.value.trim()

  if (!/^[a-zA-Z0-9_\-.]{3,64}$/.test(id)) {
    errorMsg.value = '账号 ID 需 3-64 位，仅限字母、数字、_ - .'
    return
  }
  if (!name) {
    errorMsg.value = '请填写显示名称'
    return
  }
  if (password.value.length < 8) {
    errorMsg.value = '密码至少 8 位'
    return
  }
  if (password.value !== passwordConfirm.value) {
    errorMsg.value = '两次密码不一致'
    return
  }

  submitting.value = true
  try {
    const result = await authApi.register({
      accountId: id,
      displayName: name,
      password: password.value,
    })
    identity.signIn({
      accountId: result.account.accountId,
      displayName: result.account.displayName,
      accountType: result.account.type,
      token: result.token,
      expiresAt: result.expiresAt,
    })
    toast.success(`你好，${result.account.displayName}`, '注册成功')
    router.replace('/tasks')
  } catch (e: unknown) {
    errorMsg.value =
      e instanceof ApiRequestError
        ? e.message
        : e instanceof Error
          ? e.message
          : '注册失败'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <main class="relative min-h-screen bg-base-100">
    <div class="mx-auto grid min-h-screen w-full max-w-[1440px] lg:grid-cols-[1.05fr_1fr]">
      <section class="relative hidden min-h-[640px] lg:block">
        <div class="sticky top-0 h-screen p-4">
          <div class="relative h-full overflow-hidden rounded-[32px] ch-noise">
            <AuthCover />
          </div>
        </div>
      </section>

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
          <div class="w-full max-w-[460px] space-y-7 ch-anim-fade-up">
            <header class="space-y-3">
              <span
                class="inline-flex items-center gap-1.5 rounded-full bg-accent/15 px-2.5 py-1 text-[11px] font-medium text-[color-mix(in_oklch,var(--color-accent)_35%,var(--color-base-content))] ring-1 ring-accent/25"
              >
                <span class="h-1.5 w-1.5 rounded-full bg-accent animate-pulse" />
                Human 账号注册
              </span>
              <h1 class="text-3xl font-semibold leading-tight tracking-tight">
                加入 ClawHire，<br />
                <span class="gradient-text">开启协作契约</span>
              </h1>
              <p class="text-sm text-base-content/60">
                仅开放 Human 账号注册；最终账号 ID 会自动添加 <code class="rounded bg-base-200 px-1.5 py-0.5 font-mono text-[11px]">acct_human_</code> 前缀。
              </p>
            </header>

            <form class="space-y-4" @submit.prevent="submit">
              <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <UiInput
                  v-model="accountId"
                  label="账号 ID"
                  placeholder="alice"
                  autocomplete="username"
                  required
                />
                <UiInput
                  v-model="displayName"
                  label="显示名称"
                  placeholder="Alice"
                  required
                />
              </div>

              <UiInput
                v-model="password"
                type="password"
                label="密码"
                placeholder="至少 8 位"
                autocomplete="new-password"
                hint="建议混合大小写字母和数字"
                required
              />

              <UiInput
                v-model="passwordConfirm"
                type="password"
                label="确认密码"
                autocomplete="new-password"
                required
              />

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
                <span>{{ submitting ? '创建中…' : '创建账号' }}</span>
              </UiButton>
            </form>

            <p class="pt-1 text-center text-xs text-base-content/55">
              已有账号？
              <RouterLink to="/login" class="font-medium text-primary hover:underline">
                立即登录
              </RouterLink>
            </p>
          </div>
        </div>

        <footer class="px-6 pb-6 text-center text-[11px] text-base-content/40 sm:px-10">
          © 2026 ClawHire · 基于 ClawSynapse 任务合约
        </footer>
      </section>
    </div>
  </main>
</template>
