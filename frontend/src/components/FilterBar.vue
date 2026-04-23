<script setup lang="ts">
import { computed } from 'vue'
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
    chips.push({ key: 'category', label: `分类 · ${props.modelValue.category}` })
  if (props.modelValue.keyword)
    chips.push({ key: 'keyword', label: `关键词 · ${props.modelValue.keyword}` })
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
</script>

<template>
  <div class="rounded-xl border border-base-300 bg-base-100 p-3">
    <div class="flex flex-wrap items-center gap-2">
      <div class="relative min-w-[220px] flex-1">
        <input
          :value="modelValue.keyword ?? ''"
          type="text"
          placeholder="按标题或分类搜索"
          class="w-full rounded-md border border-base-300 bg-base-100 py-1.5 pr-3 pl-8 text-sm text-base-content outline-none focus:border-primary/60 focus:ring-1 focus:ring-primary/30"
          @input="
            patch({
              keyword:
                ($event.target as HTMLInputElement).value.trim() || undefined,
            })
          "
        />
        <svg
          class="pointer-events-none absolute top-2 left-2.5 h-4 w-4 text-base-content/40"
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
      </div>

      <select
        :value="modelValue.status ?? ''"
        class="rounded-md border border-base-300 bg-base-100 py-1.5 pr-8 pl-3 text-sm outline-none focus:border-primary/60"
        @change="
          patch({
            status:
              (($event.target as HTMLSelectElement).value as TaskStatus) ||
              undefined,
          })
        "
      >
        <option value="">全部状态</option>
        <option v-for="s in TASK_STATUSES" :key="s" :value="s">
          {{ STATUS_LABELS[s] }}
        </option>
      </select>

      <select
        v-if="categories && categories.length"
        :value="modelValue.category ?? ''"
        class="rounded-md border border-base-300 bg-base-100 py-1.5 pr-8 pl-3 text-sm outline-none focus:border-primary/60"
        @change="
          patch({
            category:
              ($event.target as HTMLSelectElement).value || undefined,
          })
        "
      >
        <option value="">全部分类</option>
        <option v-for="c in categories" :key="c" :value="c">{{ c }}</option>
      </select>

      <span
        v-if="resultLabel"
        class="ml-auto text-xs text-base-content/50"
      >{{ resultLabel }}</span>
    </div>

    <div
      v-if="activeChips.length > 0"
      class="mt-2 flex flex-wrap items-center gap-1.5 border-t border-base-200 pt-2 text-xs"
    >
      <span class="text-base-content/40">当前筛选：</span>
      <button
        v-for="chip in activeChips"
        :key="chip.key"
        type="button"
        class="inline-flex items-center gap-1 rounded-full border border-base-300 bg-base-200 px-2 py-0.5 text-base-content/70 hover:border-primary/40 hover:text-base-content"
        @click="clear(chip.key)"
      >
        {{ chip.label }}
        <span aria-hidden="true" class="text-base-content/40">×</span>
      </button>
      <button
        type="button"
        class="ml-1 text-base-content/50 underline-offset-2 hover:underline"
        @click="clearAll"
      >
        全部清除
      </button>
    </div>
  </div>
</template>
