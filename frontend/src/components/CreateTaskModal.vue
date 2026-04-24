<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { ApiRequestError, tasksApi } from '@/api'
import type { CreateTaskInput } from '@/api/tasks'

const props = defineProps<{
  open: boolean
  defaultCategory?: string
  categories?: string[]
}>()

const emit = defineEmits<{
  close: []
  created: [taskId: string]
}>()

interface FormState {
  title: string
  category: string
  description: string
  reviewerId: string
  rewardMode: 'fixed' | 'bid' | 'milestone'
  rewardAmount: string
  rewardCurrency: string
  acceptanceMode: 'manual' | 'schema' | 'test' | 'hybrid'
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

watch(
  () => props.open,
  (open) => {
    if (open) {
      Object.assign(form, emptyForm())
      errorMessage.value = null
    }
  },
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
    errorMessage.value =
      err instanceof ApiRequestError
        ? err.message
        : err instanceof Error
          ? err.message
          : '发布失败'
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
  <Transition
    enter-active-class="transition duration-150"
    enter-from-class="opacity-0"
    leave-active-class="transition duration-100"
    leave-to-class="opacity-0"
  >
    <div
      v-if="open"
      class="fixed inset-0 z-40 flex items-start justify-center overflow-y-auto bg-black/40 p-4 sm:p-6"
      role="dialog"
      aria-modal="true"
      @click.self="handleClose"
    >
      <div
        class="my-8 w-full max-w-2xl rounded-xl border border-base-300 bg-base-100 shadow-xl"
      >
        <header
          class="flex items-center justify-between border-b border-base-200 px-5 py-3"
        >
          <div>
            <h2 class="text-base font-semibold text-base-content">发布新任务</h2>
            <p class="mt-0.5 text-xs text-base-content/55">
              填写任务信息，提交后会立即在任务大厅可见。
            </p>
          </div>
          <button
            type="button"
            class="rounded-md p-1 text-base-content/50 hover:bg-base-200 hover:text-base-content"
            aria-label="关闭"
            @click="handleClose"
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
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </header>

        <form class="space-y-4 px-5 py-4" @submit="handleSubmit">
          <div class="space-y-1">
            <label class="text-xs font-medium text-base-content/70">
              任务标题 <span class="text-red-500">*</span>
            </label>
            <input
              v-model="form.title"
              type="text"
              maxlength="120"
              required
              placeholder="例如：构建品牌落地页"
              class="w-full rounded-md border border-base-300 bg-base-100 px-3 py-1.5 text-sm outline-none focus:border-primary/60 focus:ring-1 focus:ring-primary/30"
            />
          </div>

          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div class="space-y-1">
              <label class="text-xs font-medium text-base-content/70">
                分类 <span class="text-red-500">*</span>
              </label>
              <input
                v-model="form.category"
                type="text"
                required
                :list="categories?.length ? 'create-task-categories' : undefined"
                placeholder="如 coding / writing"
                class="w-full rounded-md border border-base-300 bg-base-100 px-3 py-1.5 text-sm outline-none focus:border-primary/60 focus:ring-1 focus:ring-primary/30"
              />
              <datalist
                v-if="categories?.length"
                id="create-task-categories"
              >
                <option v-for="c in categories" :key="c" :value="c" />
              </datalist>
            </div>
            <div class="space-y-1">
              <label class="text-xs font-medium text-base-content/70">
                截止时间（可选）
              </label>
              <input
                v-model="form.deadline"
                type="datetime-local"
                class="w-full rounded-md border border-base-300 bg-base-100 px-3 py-1.5 text-sm outline-none focus:border-primary/60 focus:ring-1 focus:ring-primary/30"
              />
            </div>
          </div>

          <div class="space-y-1">
            <label class="text-xs font-medium text-base-content/70">
              任务描述（可选）
            </label>
            <textarea
              v-model="form.description"
              rows="4"
              placeholder="说明背景、范围和交付物。"
              class="w-full rounded-md border border-base-300 bg-base-100 px-3 py-2 text-sm leading-relaxed outline-none focus:border-primary/60 focus:ring-1 focus:ring-primary/30"
            />
          </div>

          <fieldset class="rounded-lg border border-base-200 px-3 py-3">
            <legend class="px-1 text-xs font-medium text-base-content/70">
              报酬
            </legend>
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
              <div class="space-y-1">
                <label class="text-[11px] text-base-content/55">模式</label>
                <select
                  v-model="form.rewardMode"
                  class="w-full rounded-md border border-base-300 bg-base-100 px-3 py-1.5 text-sm outline-none focus:border-primary/60"
                >
                  <option value="fixed">固定价</option>
                  <option value="bid">竞价</option>
                  <option value="milestone">按里程碑</option>
                </select>
              </div>
              <div class="space-y-1">
                <label class="text-[11px] text-base-content/55">金额</label>
                <input
                  v-model="form.rewardAmount"
                  type="number"
                  min="0"
                  step="any"
                  :placeholder="form.rewardMode === 'milestone' ? '可留空' : '例如 300'"
                  class="w-full rounded-md border border-base-300 bg-base-100 px-3 py-1.5 text-sm outline-none focus:border-primary/60 focus:ring-1 focus:ring-primary/30"
                />
              </div>
              <div class="space-y-1">
                <label class="text-[11px] text-base-content/55">币种</label>
                <input
                  v-model="form.rewardCurrency"
                  type="text"
                  maxlength="6"
                  placeholder="USD"
                  class="w-full rounded-md border border-base-300 bg-base-100 px-3 py-1.5 text-sm outline-none focus:border-primary/60 focus:ring-1 focus:ring-primary/30"
                />
              </div>
            </div>
          </fieldset>

          <fieldset class="rounded-lg border border-base-200 px-3 py-3">
            <legend class="px-1 text-xs font-medium text-base-content/70">
              验收
            </legend>
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
              <div class="space-y-1">
                <label class="text-[11px] text-base-content/55">验收方式</label>
                <select
                  v-model="form.acceptanceMode"
                  class="w-full rounded-md border border-base-300 bg-base-100 px-3 py-1.5 text-sm outline-none focus:border-primary/60"
                >
                  <option value="manual">人工验收</option>
                  <option value="schema">Schema 校验</option>
                  <option value="test">测试驱动</option>
                  <option value="hybrid">混合验收</option>
                </select>
              </div>
              <div class="sm:col-span-2 space-y-1">
                <label class="text-[11px] text-base-content/55">
                  验收规则（每行一条，可选）
                </label>
                <textarea
                  v-model="form.acceptanceRules"
                  rows="2"
                  placeholder="例如：UI 走查通过"
                  class="w-full rounded-md border border-base-300 bg-base-100 px-3 py-2 text-sm leading-relaxed outline-none focus:border-primary/60 focus:ring-1 focus:ring-primary/30"
                />
              </div>
            </div>
          </fieldset>

          <div class="space-y-1">
            <label class="text-xs font-medium text-base-content/70">
              指定验收方账号 ID（可选）
            </label>
            <input
              v-model="form.reviewerId"
              type="text"
              placeholder="留空则由你自己担任验收方"
              class="w-full rounded-md border border-base-300 bg-base-100 px-3 py-1.5 text-sm outline-none focus:border-primary/60 focus:ring-1 focus:ring-primary/30"
            />
          </div>

          <p
            v-if="errorMessage"
            class="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-xs text-red-700"
          >
            {{ errorMessage }}
          </p>

          <footer
            class="flex items-center justify-end gap-2 border-t border-base-200 pt-3"
          >
            <button
              type="button"
              class="rounded-md border border-base-300 bg-base-100 px-3 py-1.5 text-sm text-base-content hover:border-primary/40 hover:text-primary disabled:opacity-50"
              :disabled="submitting"
              @click="handleClose"
            >
              取消
            </button>
            <button
              type="submit"
              class="rounded-md bg-primary px-4 py-1.5 text-sm text-primary-content hover:bg-primary/90 disabled:bg-primary/60"
              :disabled="submitting"
            >
              {{ submitting ? '发布中…' : '发布任务' }}
            </button>
          </footer>
        </form>
      </div>
    </div>
  </Transition>
</template>
