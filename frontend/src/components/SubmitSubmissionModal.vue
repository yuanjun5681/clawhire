<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import UiModal from '@/components/ui/UiModal.vue'
import { UiButton, UiInput, UiTextarea } from '@/components/ui'
import type { CreateSubmissionInput } from '@/api/tasks'

type SubmissionDraft = Omit<CreateSubmissionInput, 'submissionId'>

const props = defineProps<{
  open: boolean
  submitting?: boolean
}>()

const emit = defineEmits<{
  close: []
  submit: [payload: SubmissionDraft]
}>()

const summary = ref('')
const finalOutput = ref('')
const artifactUrl = ref('')
const artifactName = ref('')
const evidenceUrl = ref('')

watch(
  () => props.open,
  (open) => {
    if (!open) return
    summary.value = ''
    finalOutput.value = ''
    artifactUrl.value = ''
    artifactName.value = ''
    evidenceUrl.value = ''
  },
)

const canSubmit = computed(() => summary.value.trim().length > 0 && !props.submitting)

function submit() {
  if (!canSubmit.value) return

  const artifact = artifactUrl.value.trim()
  const evidence = evidenceUrl.value.trim()

  emit('submit', {
    summary: summary.value.trim(),
    finalOutput: finalOutput.value.trim() || undefined,
    artifacts: artifact
      ? [
          {
            type: 'url',
            url: artifact,
            name: artifactName.value.trim() || '交付物',
          },
        ]
      : [],
    evidence: evidence
      ? {
          type: 'url',
          items: [evidence],
        }
      : undefined,
  })
}
</script>

<template>
  <UiModal
    :open="open"
    title="提交交付"
    description="填写摘要、交付正文和相关链接。"
    size="md"
    :close-on-overlay="!submitting"
    @close="$emit('close')"
  >
    <form class="space-y-4" @submit.prevent="submit">
      <UiInput
        v-model="summary"
        label="交付摘要"
        placeholder="例如：已完成招聘文案初稿"
        required
      />

      <UiTextarea
        v-model="finalOutput"
        label="交付正文"
        placeholder="适合填写短文案、最终答案、说明文字等纯文本交付。"
        :rows="6"
      />

      <div class="grid grid-cols-1 gap-3 sm:grid-cols-[1fr_0.8fr]">
        <UiInput
          v-model="artifactUrl"
          label="交付物链接"
          placeholder="https://..."
          type="url"
        />
        <UiInput
          v-model="artifactName"
          label="链接名称"
          placeholder="交付物"
        />
      </div>

      <UiInput
        v-model="evidenceUrl"
        label="验收证据"
        placeholder="https://..."
        type="url"
      />

      <button type="submit" class="hidden" />
    </form>

    <template #footer>
      <UiButton
        variant="ghost"
        size="sm"
        :disabled="submitting"
        @click="$emit('close')"
      >
        取消
      </UiButton>
      <UiButton
        size="sm"
        :disabled="!canSubmit"
        :loading="submitting"
        @click="submit"
      >
        提交交付
      </UiButton>
    </template>
  </UiModal>
</template>
