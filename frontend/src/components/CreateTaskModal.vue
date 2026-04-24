<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ApiRequestError, tasksApi } from '@/api'
import type { CreateTaskInput } from '@/api/tasks'
import { useToastStore } from '@/stores/toast'
import {
  UiBadge,
  UiButton,
  UiInput,
  UiLabel,
  UiModal,
  UiSelect,
  UiTextarea,
} from './ui'

const props = defineProps<{
  open: boolean
  defaultCategory?: string
  categories?: string[]
}>()

const emit = defineEmits<{
  close: []
  created: [taskId: string]
}>()

const toast = useToastStore()

type RewardMode = 'fixed' | 'bid' | 'milestone'
type AcceptanceMode = 'manual' | 'schema' | 'test' | 'hybrid'

interface FormState {
  title: string
  category: string
  description: string
  reviewerId: string
  rewardMode: RewardMode
  rewardAmount: string
  rewardCurrency: string
  acceptanceMode: AcceptanceMode
  acceptanceRules: string
  deadline: string
}

function emptyForm(): FormState {
  return {
    title: '',
    category: props.defaultCategory ?? 'coding',
    description: '',
    reviewerId: '',
    rewardMode: 'fixed',
    rewardAmount: '',
    rewardCurrency: 'USD',
    acceptanceMode: 'manual',
    acceptanceRules: '',
    deadline: '',
  }
}

const form = reactive<FormState>(emptyForm())
const submitting = ref(false)
const errorMessage = ref<string | null>(null)

const REWARD_OPTIONS: Array<{ label: string; value: RewardMode; hint: string }> = [
  { label: '固定价', value: 'fixed', hint: '一次性支付' },
  { label: '竞价', value: 'bid', hint: '由执行方报价' },
  { label: '里程碑', value: 'milestone', hint: '按阶段结算' },
]

const ACCEPTANCE_OPTIONS: Array<{ label: string; value: AcceptanceMode; hint: string }> = [
  { label: '人工验收', value: 'manual', hint: '由验收方主观判定' },
  { label: 'Schema 校验', value: 'schema', hint: '按契约字段核对' },
  { label: '测试驱动', value: 'test', hint: '跑测试用例' },
  { label: '混合验收', value: 'hybrid', hint: '人工 + 自动' },
]

watch(
  () => props.open,
  (open) => {
    if (open) {
      Object.assign(form, emptyForm())
      errorMessage.value = null
    }
  },
)

const categoryOptions = computed(() =>
  (props.categories ?? []).map((c) => ({ label: c, value: c })),
)

function nextTaskId() {
  const rand =
    typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
      ? crypto.randomUUID().slice(0, 8)
      : Math.random().toString(36).slice(2, 10)
  return `task_${rand}`
}

async function handleSubmit(e: Event) {
  e.preventDefault()
  if (submitting.value) return
  errorMessage.value = null

  const title = form.title.trim()
  const category = form.category.trim()
  if (!title) {
    errorMessage.value = '请填写任务标题'
    return
  }
  if (!category) {
    errorMessage.value = '请填写任务分类'
    return
  }

  const amount = Number(form.rewardAmount)
  if (form.rewardMode !== 'milestone') {
    if (!form.rewardAmount || !Number.isFinite(amount) || amount < 0) {
      errorMessage.value = '请填写正确的报酬金额'
      return
    }
  }

  const rules = form.acceptanceRules
    .split('\n')
    .map((r) => r.trim())
    .filter(Boolean)

  const payload: CreateTaskInput = {
    taskId: nextTaskId(),
    title,
    category,
    description: form.description.trim() || undefined,
    reviewerId: form.reviewerId.trim() || undefined,
    reward: {
      mode: form.rewardMode,
      amount: Number.isFinite(amount) ? amount : 0,
      currency: form.rewardCurrency.trim() || 'USD',
    },
    acceptanceSpec: {
      mode: form.acceptanceMode,
      rules,
    },
    deadline: form.deadline ? new Date(form.deadline).toISOString() : undefined,
  }

  try {
    submitting.value = true
    const res = await tasksApi.createTask(payload)
    emit('created', res.taskId)
  } catch (err: unknown) {
    const msg =
      err instanceof ApiRequestError
        ? err.message
        : err instanceof Error
          ? err.message
          : '发布失败'
    errorMessage.value = msg
    toast.error(msg, '发布失败')
  } finally {
    submitting.value = false
  }
}

function handleClose() {
  if (submitting.value) return
  emit('close')
}
</script>

<template>
  <UiModal
    :open="open"
    size="lg"
    :close-on-overlay="!submitting"
    @close="handleClose"
  >
    <template #header>
      <div class="flex items-start gap-3">
        <span
          class="mt-0.5 inline-flex h-9 w-9 items-center justify-center rounded-[12px] bg-[linear-gradient(135deg,color-mix(in_oklch,var(--color-primary)_30%,var(--color-base-100)),color-mix(in_oklch,var(--color-accent)_28%,var(--color-base-100)))] text-primary shadow-[0_8px_24px_-12px_color-mix(in_oklch,var(--color-primary)_55%,transparent)]"
        >
          <svg
            class="h-4 w-4"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
            <polyline points="14 2 14 8 20 8" />
            <line x1="12" y1="18" x2="12" y2="12" />
            <line x1="9" y1="15" x2="15" y2="15" />
          </svg>
        </span>
        <div class="min-w-0 flex-1">
          <h2 class="text-[15px] font-semibold tracking-tight">
            发布新任务
            <span class="gradient-text">· 契约驱动</span>
          </h2>
          <p class="mt-0.5 text-xs text-base-content/60">
            填写任务信息，提交后会立即在任务大厅可见。
          </p>
        </div>
      </div>
    </template>

    <form id="create-task-form" class="space-y-5" @submit="handleSubmit">
      <UiInput
        v-model="form.title"
        label="任务标题"
        required
        maxlength="120"
        placeholder="例如：构建品牌落地页"
      />

      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <UiSelect
          v-if="categoryOptions.length > 0"
          v-model="form.category"
          label="分类"
          required
          :options="categoryOptions"
        />
        <UiInput
          v-else
          v-model="form.category"
          label="分类"
          required
          placeholder="如 coding / writing"
        />
        <UiInput
          v-model="form.deadline"
          label="截止时间（可选）"
          type="datetime-local"
        />
      </div>

      <UiTextarea
        v-model="form.description"
        label="任务描述（可选）"
        :rows="4"
        placeholder="说明背景、范围和交付物。"
      />

      <!-- Reward group -->
      <section
        class="relative overflow-hidden rounded-box border border-base-300/70 bg-base-200/40 p-4"
      >
        <span
          aria-hidden="true"
          class="pointer-events-none absolute -right-16 -top-16 h-36 w-36 rounded-full bg-primary/15 blur-3xl"
        />
        <header class="relative mb-3 flex items-center gap-2">
          <svg
            class="h-4 w-4 text-primary"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <circle cx="12" cy="12" r="10" />
            <path d="M12 6v12M9 9h5a2 2 0 0 1 0 4H9m0 0h6" />
          </svg>
          <UiLabel>报酬</UiLabel>
        </header>
        <div class="relative space-y-3">
          <div
            role="radiogroup"
            aria-label="报酬模式"
            class="grid grid-cols-3 gap-2"
          >
            <button
              v-for="opt in REWARD_OPTIONS"
              :key="opt.value"
              type="button"
              role="radio"
              :aria-checked="form.rewardMode === opt.value"
              :class="[
                'group flex flex-col items-start gap-1 rounded-box border px-3 py-2.5 text-left transition',
                form.rewardMode === opt.value
                  ? 'border-primary/60 bg-[linear-gradient(135deg,color-mix(in_oklch,var(--color-primary)_12%,var(--color-base-100)),var(--color-base-100))] shadow-[0_10px_24px_-14px_color-mix(in_oklch,var(--color-primary)_55%,transparent)]'
                  : 'border-base-300/70 bg-base-100 hover:border-primary/30',
              ]"
              @click="form.rewardMode = opt.value"
            >
              <span
                :class="[
                  'text-sm font-medium',
                  form.rewardMode === opt.value ? 'text-primary' : 'text-base-content',
                ]"
              >{{ opt.label }}</span>
              <span class="text-[11px] text-base-content/55">{{ opt.hint }}</span>
            </button>
          </div>

          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <UiInput
              v-model="form.rewardAmount"
              label="金额"
              type="number"
              min="0"
              step="any"
              :placeholder="form.rewardMode === 'milestone' ? '可留空' : '例如 300'"
              prefix-icon
            >
              <template #prefix>
                <span class="font-mono text-xs">$</span>
              </template>
            </UiInput>
            <UiInput
              v-model="form.rewardCurrency"
              label="币种"
              maxlength="6"
              placeholder="USD"
            />
          </div>
        </div>
      </section>

      <!-- Acceptance group -->
      <section
        class="relative overflow-hidden rounded-box border border-base-300/70 bg-base-200/40 p-4"
      >
        <span
          aria-hidden="true"
          class="pointer-events-none absolute -left-16 -bottom-16 h-36 w-36 rounded-full bg-accent/15 blur-3xl"
        />
        <header class="relative mb-3 flex items-center gap-2">
          <svg
            class="h-4 w-4 text-accent"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
            <polyline points="22 4 12 14.01 9 11.01" />
          </svg>
          <UiLabel>验收</UiLabel>
        </header>
        <div class="relative grid grid-cols-1 gap-3 sm:grid-cols-5">
          <div class="sm:col-span-2">
            <UiSelect
              v-model="form.acceptanceMode"
              label="验收方式"
              :options="ACCEPTANCE_OPTIONS.map((o) => ({ label: o.label, value: o.value }))"
            />
            <p class="mt-1.5 flex items-center gap-1 text-[11px] text-base-content/55">
              <UiBadge tone="accent" size="xs" soft>hint</UiBadge>
              {{
                ACCEPTANCE_OPTIONS.find((o) => o.value === form.acceptanceMode)?.hint
              }}
            </p>
          </div>
          <div class="sm:col-span-3">
            <UiTextarea
              v-model="form.acceptanceRules"
              label="验收规则（每行一条，可选）"
              :rows="3"
              placeholder="例如：UI 走查通过&#10;接口响应 &lt; 200ms"
            />
          </div>
        </div>
      </section>

      <UiInput
        v-model="form.reviewerId"
        label="指定验收方账号 ID（可选）"
        placeholder="留空则由你自己担任验收方"
        prefix-icon
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

      <Transition
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="opacity-0 -translate-y-1"
        leave-active-class="transition duration-150 ease-in"
        leave-to-class="opacity-0"
      >
        <p
          v-if="errorMessage"
          class="flex items-start gap-2 rounded-box border border-error/40 bg-error/10 px-3.5 py-2.5 text-xs text-error"
        >
          <svg
            class="mt-0.5 h-4 w-4 shrink-0"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <circle cx="12" cy="12" r="10" />
            <line x1="12" y1="8" x2="12" y2="12" />
            <line x1="12" y1="16" x2="12.01" y2="16" />
          </svg>
          {{ errorMessage }}
        </p>
      </Transition>
    </form>

    <template #footer>
      <UiButton variant="ghost" :disabled="submitting" @click="handleClose">
        取消
      </UiButton>
      <UiButton
        type="submit"
        form="create-task-form"
        variant="primary"
        :loading="submitting"
      >
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
          <path d="m5 12 5 5L20 7" />
        </svg>
        {{ submitting ? '发布中…' : '发布任务' }}
      </UiButton>
    </template>
  </UiModal>
</template>
