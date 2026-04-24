<script setup lang="ts">
import { computed } from 'vue'
import UiInput from './ui/UiInput.vue'
import UiSelect from './ui/UiSelect.vue'
import UiBadge from './ui/UiBadge.vue'
import { TASK_STATUSES, type TaskQuery, type TaskStatus } from '@/types'

const props = defineProps<{
  modelValue: TaskQuery
  categories?: string[]
  resultLabel?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: TaskQuery]
}>()

const STATUS_LABELS: Record<TaskStatus, string> = {
  OPEN: '开放',
  BIDDING: '报价中',
  AWARDED: '已指派',
  IN_PROGRESS: '执行中',
  SUBMITTED: '待验收',
  ACCEPTED: '已通过',
  SETTLED: '已结算',
  REJECTED: '已驳回',
  CANCELLED: '已取消',
  EXPIRED: '已过期',
  DISPUTED: '争议中',
}

function patch(partial: Partial<TaskQuery>) {
  emit('update:modelValue', { ...props.modelValue, page: 1, ...partial })
}

const activeChips = computed(() => {
  const chips: { key: keyof TaskQuery; label: string }[] = []
  if (props.modelValue.status)
    chips.push({
      key: 'status',
      label: `状态 · ${STATUS_LABELS[props.modelValue.status]}`,
    })
  if (props.modelValue.category)
    chips.push({
      key: 'category',
      label: `分类 · ${props.modelValue.category}`,
    })
  if (props.modelValue.keyword)
    chips.push({
      key: 'keyword',
      label: `关键词 · ${props.modelValue.keyword}`,
    })
  return chips
})

function clear(key: keyof TaskQuery) {
  const next = { ...props.modelValue, page: 1 }
  delete next[key]
  emit('update:modelValue', next)
}

function clearAll() {
  emit('update:modelValue', { page: 1, pageSize: props.modelValue.pageSize })
}

const keyword = computed({
  get: () => props.modelValue.keyword ?? '',
  set: (v: string) => patch({ keyword: v.trim() || undefined }),
})
</script>

<template>
  <div
    class="rounded-box border border-base-300/70 bg-base-100/80 p-3 shadow-[0_2px_0_color-mix(in_oklch,var(--color-base-content)_3%,transparent)] backdrop-blur"
  >
    <div class="grid grid-cols-1 gap-2.5 sm:grid-cols-[1fr_auto_auto_auto] sm:items-end">
      <UiInput
        v-model="keyword"
        size="md"
        placeholder="搜索标题或分类…"
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
            <circle cx="11" cy="11" r="7" />
            <path d="m21 21-4.3-4.3" />
          </svg>
        </template>
      </UiInput>

      <UiSelect
        :model-value="modelValue.status ?? ''"
        size="md"
        @update:model-value="
          (v: string) => patch({ status: (v as TaskStatus) || undefined })
        "
      >
        <option value="">全部状态</option>
        <option v-for="s in TASK_STATUSES" :key="s" :value="s">
          {{ STATUS_LABELS[s] }}
        </option>
      </UiSelect>

      <UiSelect
        v-if="categories && categories.length"
        :model-value="modelValue.category ?? ''"
        size="md"
        @update:model-value="
          (v: string) => patch({ category: v || undefined })
        "
      >
        <option value="">全部分类</option>
        <option v-for="c in categories" :key="c" :value="c">{{ c }}</option>
      </UiSelect>

      <span
        v-if="resultLabel"
        class="hidden text-xs text-base-content/55 sm:inline-flex sm:items-center sm:h-11 sm:px-3 sm:rounded-field sm:bg-base-200/60"
      >
        {{ resultLabel }}
      </span>
    </div>

    <div
      v-if="activeChips.length > 0"
      class="mt-3 flex flex-wrap items-center gap-1.5 border-t border-base-300/60 pt-2.5 text-xs"
    >
      <span class="text-base-content/45">当前筛选：</span>
      <button
        v-for="chip in activeChips"
        :key="chip.key"
        type="button"
        class="group"
        @click="clear(chip.key)"
      >
        <UiBadge tone="primary" size="sm">
          {{ chip.label }}
          <span aria-hidden="true" class="ml-0.5 opacity-60 group-hover:opacity-100">×</span>
        </UiBadge>
      </button>
      <button
        type="button"
        class="ml-1 text-base-content/50 underline-offset-4 hover:underline hover:text-base-content"
        @click="clearAll"
      >
        全部清除
      </button>
    </div>
  </div>
</template>
