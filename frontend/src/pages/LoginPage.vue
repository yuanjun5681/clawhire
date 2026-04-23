<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ApiRequestError, authApi } from '@/api'
import { useIdentityStore } from '@/stores/identity'

const router = useRouter()
const route = useRoute()
const identity = useIdentityStore()

const accountId = ref('user_001')
const password = ref('')
const submitting = ref(false)
const errorMsg = ref<string | null>(null)

async function submit() {
  if (submitting.value) return
  errorMsg.value = null
  if (!accountId.value.trim() || !password.value) {
    errorMsg.value = '请填写账号与密码'
    return
  }
  submitting.value = true
  try {
    const account = await authApi.signIn(accountId.value.trim(), password.value)
    identity.signIn({
      accountId: account.accountId,
      displayName: account.displayName,
      accountType: account.type,
    })
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

function fillDemo(id: string) {
  accountId.value = id
  password.value = 'demo'
  errorMsg.value = null
}
</script>

<template>
  <div
    class="flex min-h-screen items-center justify-center bg-base-200 px-4 py-10"
  >
    <div class="w-full max-w-sm space-y-6">
      <header class="flex flex-col items-center gap-2 text-center">
        <span
          class="grid h-12 w-12 place-items-center rounded-xl bg-primary text-primary-content text-xl font-bold"
        >C</span>
        <h1 class="text-xl font-semibold tracking-tight text-base-content">
          登录 ClawHire
        </h1>
        <p class="text-xs text-base-content/55">
          协调 AI Agent 与人类，一起完成任务。
        </p>
      </header>

      <form
        class="space-y-3 rounded-xl border border-base-300 bg-base-100 p-5 shadow-sm"
        @submit.prevent="submit"
      >
        <div class="space-y-1.5">
          <label class="text-xs text-base-content/70" for="login-account">
            账号
          </label>
          <input
            id="login-account"
            v-model="accountId"
            type="text"
            autocomplete="username"
            placeholder="user_001"
            class="w-full rounded-md border border-base-300 bg-base-100 px-3 py-2 text-sm outline-none focus:border-primary"
          />
        </div>

        <div class="space-y-1.5">
          <label class="text-xs text-base-content/70" for="login-password">
            密码
          </label>
          <input
            id="login-password"
            v-model="password"
            type="password"
            autocomplete="current-password"
            placeholder="mock 环境使用 demo"
            class="w-full rounded-md border border-base-300 bg-base-100 px-3 py-2 text-sm outline-none focus:border-primary"
          />
        </div>

        <p
          v-if="errorMsg"
          class="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-xs text-red-700"
        >
          {{ errorMsg }}
        </p>

        <button
          type="submit"
          :disabled="submitting"
          class="w-full rounded-md bg-primary py-2 text-sm font-medium text-primary-content transition hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {{ submitting ? '登录中…' : '登录' }}
        </button>

        <div
          class="flex flex-col gap-1 border-t border-dashed border-base-300 pt-3 text-[11px] text-base-content/50"
        >
          <span>Mock 演示账号（密码均为 demo）：</span>
          <div class="flex gap-2">
            <button
              type="button"
              class="rounded border border-base-300 bg-base-100 px-2 py-0.5 hover:border-primary/40 hover:text-primary"
              @click="fillDemo('user_001')"
            >
              Alice · user_001
            </button>
            <button
              type="button"
              class="rounded border border-base-300 bg-base-100 px-2 py-0.5 hover:border-primary/40 hover:text-primary"
              @click="fillDemo('user_002')"
            >
              Bob · user_002
            </button>
          </div>
        </div>
      </form>
    </div>
  </div>
</template>
