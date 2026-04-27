<script setup lang="ts">
import { ref, watch } from 'vue'
import { ApiRequestError, connectionsApi } from '@/api'
import { UiModal, UiButton } from '@/components/ui'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{
  close: []
  opened: []
}>()

const submitting = ref(false)
const error = ref('')

watch(
  () => props.open,
  (open) => {
    if (!open) return
    error.value = ''
  },
)

async function connectTrustMesh() {
  error.value = ''
  const popup = window.open('', '_blank')
  if (!popup) {
    error.value = '浏览器阻止了新窗口，请允许弹窗后重试'
    return
  }
  popup.opener = null
  submitting.value = true
  try {
    const { url } = await connectionsApi.getTrustMeshConnectURL()
    popup.location.href = url
    emit('opened')
    emit('close')
  } catch (e) {
    popup.close()
    if (e instanceof ApiRequestError) {
      error.value = e.message
    } else {
      error.value = '打开 TrustMesh 授权页失败，请重试'
    }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <UiModal
    :open="open"
    title="连接 TrustMesh"
    description="前往 TrustMesh 授权页完成绑定。"
    size="sm"
    @close="$emit('close')"
  >
    <div class="space-y-4">
      <button
        type="button"
        class="flex w-full items-center gap-3 rounded-field border border-base-300/80 bg-base-100 px-4 py-3 text-left transition hover:border-primary/45 hover:bg-primary/5"
        @click="connectTrustMesh"
      >
        <span
          class="grid h-9 w-9 shrink-0 place-items-center rounded-lg bg-[linear-gradient(120deg,#6366f1,#8b5cf6)] text-xs font-bold text-white"
        >
          TM
        </span>
        <span class="min-w-0 flex-1">
          <span class="block text-sm font-semibold text-base-content">TrustMesh</span>
          <span class="block text-xs text-base-content/55">打开授权页面并选择 PM Agent</span>
        </span>
        <svg
          class="h-4 w-4 text-base-content/45"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          aria-hidden="true"
        >
          <path d="M7 17L17 7" />
          <path d="M7 7h10v10" />
        </svg>
      </button>

      <p v-if="error" class="rounded-field bg-error/10 px-3 py-2 text-xs text-error">
        {{ error }}
      </p>
    </div>

    <template #footer>
      <UiButton variant="ghost" size="sm" @click="$emit('close')">取消</UiButton>
      <UiButton size="sm" :loading="submitting" @click="connectTrustMesh">打开授权页</UiButton>
    </template>
  </UiModal>
</template>
