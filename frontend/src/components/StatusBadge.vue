<script setup lang="ts">
import { computed } from 'vue'
import type { TaskStatus } from '@/types'

const props = withDefaults(
  defineProps<{
    status: TaskStatus
    size?: 'sm' | 'md'
  }>(),
  { size: 'md' },
)

const STATUS_MAP: Record<
  TaskStatus,
  { label: string; classes: string }
> = {
  OPEN: {
    label: '开放',
    classes: 'bg-sky-50 text-sky-700 ring-sky-200',
  },
  BIDDING: {
    label: '报价中',
    classes: 'bg-teal-50 text-teal-700 ring-teal-200',
  },
  AWARDED: {
    label: '已指派',
    classes: 'bg-indigo-50 text-indigo-700 ring-indigo-200',
  },
  IN_PROGRESS: {
    label: '执行中',
    classes: 'bg-amber-50 text-amber-800 ring-amber-200',
  },
  SUBMITTED: {
    label: '待验收',
    classes: 'bg-cyan-50 text-cyan-700 ring-cyan-200',
  },
  ACCEPTED: {
    label: '已通过',
    classes: 'bg-green-50 text-green-700 ring-green-200',
  },
  SETTLED: {
    label: '已结算',
    classes: 'bg-emerald-100 text-emerald-800 ring-emerald-200',
  },
  REJECTED: {
    label: '已驳回',
    classes: 'bg-red-50 text-red-700 ring-red-200',
  },
  CANCELLED: {
    label: '已取消',
    classes: 'bg-gray-100 text-gray-600 ring-gray-200',
  },
  EXPIRED: {
    label: '已过期',
    classes: 'bg-stone-100 text-stone-600 ring-stone-300',
  },
  DISPUTED: {
    label: '争议中',
    classes: 'bg-fuchsia-50 text-fuchsia-700 ring-fuchsia-200',
  },
}

const entry = computed(() => STATUS_MAP[props.status])

const sizeClasses = computed(() =>
  props.size === 'sm'
    ? 'px-1.5 py-0.5 text-[10px]'
    : 'px-2 py-0.5 text-xs',
)
</script>

<template>
  <span
    :class="[
      'inline-flex items-center gap-1 rounded font-medium ring-1 ring-inset whitespace-nowrap',
      entry.classes,
      sizeClasses,
    ]"
  >
    <span
      class="h-1.5 w-1.5 rounded-full bg-current opacity-60"
      aria-hidden="true"
    />
    {{ entry.label }}
  </span>
</template>
