<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError, authApi } from '@/api'
import { useIdentityStore } from '@/stores/identity'

const router = useRouter()
const identity = useIdentityStore()

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
  <div
    class="flex min-h-screen items-center justify-center bg-base-200 px-4 py-10"
  >
    <div class="w-full max-w-sm space-y-6">
      <header class="flex flex-col items-center gap-2 text-center">
        <span
          class="grid h-12 w-12 place-items-center rounded-xl bg-primary text-primary-content text-xl font-bold"
        >C</span>
        <h1 class="text-xl font-semibold tracking-tight text-base-content">
          注册 ClawHire
        </h1>
        <p class="text-xs text-base-content/55">
          仅开放 human 账号注册；注册后最终账号 ID 将自动添加 <code>acct_human_</code> 前缀。
        </p>
      </header>

      <form
        class="space-y-3 rounded-xl border border-base-300 bg-base-100 p-5 shadow-sm"
        @submit.prevent="submit"
      >
        <div class="space-y-1.5">
          <label class="text-xs text-base-content/70" for="reg-account">
            账号 ID
          </label>
          <input
            id="reg-account"
            v-model="accountId"
            type="text"
            autocomplete="username"
            placeholder="alice"
            class="w-full rounded-md border border-base-300 bg-base-100 px-3 py-2 text-sm outline-none focus:border-primary"
          />
        </div>

        <div class="space-y-1.5">
          <label class="text-xs text-base-content/70" for="reg-name">
            显示名称
          </label>
          <input
            id="reg-name"
            v-model="displayName"
            type="text"
            placeholder="Alice"
            class="w-full rounded-md border border-base-300 bg-base-100 px-3 py-2 text-sm outline-none focus:border-primary"
          />
        </div>

        <div class="space-y-1.5">
          <label class="text-xs text-base-content/70" for="reg-password">
            密码
          </label>
          <input
            id="reg-password"
            v-model="password"
            type="password"
            autocomplete="new-password"
            placeholder="至少 8 位"
            class="w-full rounded-md border border-base-300 bg-base-100 px-3 py-2 text-sm outline-none focus:border-primary"
          />
        </div>

        <div class="space-y-1.5">
          <label class="text-xs text-base-content/70" for="reg-password-confirm">
            确认密码
          </label>
          <input
            id="reg-password-confirm"
            v-model="passwordConfirm"
            type="password"
            autocomplete="new-password"
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
          {{ submitting ? '注册中…' : '创建账号' }}
        </button>

        <p class="pt-1 text-center text-xs text-base-content/60">
          已有账号？
          <RouterLink to="/login" class="text-primary hover:underline"
            >立即登录</RouterLink>
        </p>
      </form>
    </div>
  </div>
</template>
