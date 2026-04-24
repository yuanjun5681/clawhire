<script setup lang="ts">
import { computed } from 'vue'
import UiBadge from './ui/UiBadge.vue'
import type { TaskStatus } from '@/types'

const props = withDefaults(
  defineProps<{
    status: TaskStatus
    size?: 'xs' | 'sm' | 'md'
  }>(),
  { size: 'sm' },
)

type Tone =
  | 'primary'
  | 'accent'
  | 'info'
  | 'success'
  | 'warning'
  | 'error'
  | 'neutral'
  | 'secondary'

const STATUS_MAP: Record<TaskStatus, { label: string; tone: Tone }> = {
  OPEN: { label: '开放', tone: 'info' },
  BIDDING: { label: '报价中', tone: 'secondary' },
  AWARDED: { label: '已指派', tone: 'primary' },
  IN_PROGRESS: { label: '执行中', tone: 'warning' },
  SUBMITTED: { label: '待验收', tone: 'accent' },
  ACCEPTED: { label: '已通过', tone: 'success' },
  SETTLED: { label: '已结算', tone: 'success' },
  REJECTED: { label: '已驳回', tone: 'error' },
  CANCELLED: { label: '已取消', tone: 'neutral' },
  EXPIRED: { label: '已过期', tone: 'neutral' },
  DISPUTED: { label: '争议中', tone: 'error' },
}

const entry = computed(() => STATUS_MAP[props.status])
</script>

<template>
  <UiBadge :tone="entry.tone" :size="size" dot>
    {{ entry.label }}
  </UiBadge>
</template>
