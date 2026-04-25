<script setup lang="ts">
import { ref, watch } from 'vue'
import { connectionsApi, ApiRequestError } from '@/api'
import type { PlatformConnection, PlatformKind } from '@/types'
import { UiModal, UiInput, UiButton } from '@/components/ui'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{
  close: []
  created: [conn: PlatformConnection]
}>()

const PLATFORMS: { value: PlatformKind; label: string }[] = [
  { value: 'trustmesh', label: 'TrustMesh' },
]

const platform = ref<PlatformKind>('trustmesh')
const remoteUserId = ref('')
const platformNodeId = ref('')
const submitting = ref(false)
const error = ref('')

watch(
  () => props.open,
  (open) => {
    if (!open) return
    platform.value = 'trustmesh'
    remoteUserId.value = ''
    platformNodeId.value = ''
    error.value = ''
  },
)

async function submit() {
  error.value = ''
  if (!remoteUserId.value.trim()) {
    error.value = '请填写对方 User ID'
    return
  }
  submitting.value = true
  try {
    const conn = await connectionsApi.createConnection({
      platform: platform.value,
      remoteUserId: remoteUserId.value.trim(),
      platformNodeId: platformNodeId.value.trim() || undefined,
    })
    emit('created', conn)
    emit('close')
  } catch (e) {
    if (e instanceof ApiRequestError) {
      if (e.status === 409) {
        error.value = '该平台账号已绑定，请勿重复添加'
      } else {
        error.value = e.message
      }
    } else {
      error.value = '添加失败，请重试'
    }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <UiModal
    :open="open"
    title="添加平台连接"
    description="绑定外部平台账号后，ClawHire 可自动同步任务事件。"
    size="sm"
    @close="$emit('close')"
  >
    <form class="space-y-4" @submit.prevent="submit">
      <!-- Platform selector -->
      <div class="flex flex-col gap-1.5">
        <span class="text-[11px] font-medium uppercase tracking-[0.08em] text-base-content/60">
          平台 <span class="text-error">*</span>
        </span>
        <div class="flex gap-2">
          <button
            v-for="p in PLATFORMS"
            :key="p.value"
            type="button"
            :class="[
              'flex items-center gap-2 rounded-field border px-3 py-2 text-sm font-medium transition',
              platform === p.value
                ? 'border-primary/60 bg-primary/8 text-primary ring-2 ring-primary/20'
                : 'border-base-300/80 bg-base-100 text-base-content/70 hover:border-base-content/25',
            ]"
            @click="platform = p.value"
          >
            <!-- TrustMesh icon placeholder -->
            <span
              class="grid h-5 w-5 place-items-center rounded bg-[linear-gradient(120deg,#6366f1,#8b5cf6)] text-[9px] font-bold text-white"
            >
              TM
            </span>
            {{ p.label }}
          </button>
        </div>
      </div>

      <!-- Remote user ID -->
      <UiInput
        v-model="remoteUserId"
        label="对方 User ID"
        placeholder="usr_xxxx"
        required
        :error="error && !platformNodeId ? error : ''"
      />

      <!-- Platform node ID (optional) -->
      <UiInput
        v-model="platformNodeId"
        label="节点 ID（可选）"
        placeholder="留空则使用默认节点"
        hint="仅在需要连接非默认节点实例时填写"
      />

      <p v-if="error" class="rounded-field bg-error/10 px-3 py-2 text-xs text-error">
        {{ error }}
      </p>
    </form>

    <template #footer>
      <UiButton variant="ghost" size="sm" @click="$emit('close')">取消</UiButton>
      <UiButton size="sm" :loading="submitting" @click="submit">确认绑定</UiButton>
    </template>
  </UiModal>
</template>
