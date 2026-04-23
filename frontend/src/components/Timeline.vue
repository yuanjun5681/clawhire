<script setup lang="ts">
import { computed } from 'vue'
import { formatDateTime } from '@/utils/format'
import type { AccountSummary } from '@/types'

export type TimelineEventType =
  | 'bid'
  | 'assign'
  | 'progress'
  | 'submission'
  | 'review_approved'
  | 'review_rejected'
  | 'settlement'
  | 'status'

export interface TimelineEvent {
  id: string
  type: TimelineEventType
  title: string
  summary?: string
  actor?: AccountSummary
  at: string
  meta?: Array<{ label: string; value: string }>
}

const props = defineProps<{
  events: TimelineEvent[]
  emptyText?: string
}>()

const TYPE_MAP: Record<TimelineEventType, { label: string; dot: string }> = {
  bid: { label: '报价', dot: 'bg-teal-500' },
  assign: { label: '指派', dot: 'bg-indigo-500' },
  progress: { label: '进度', dot: 'bg-amber-500' },
  submission: { label: '交付', dot: 'bg-cyan-500' },
  review_approved: { label: '验收通过', dot: 'bg-green-600' },
  review_rejected: { label: '验收驳回', dot: 'bg-red-500' },
  settlement: { label: '结算', dot: 'bg-emerald-700' },
  status: { label: '状态变更', dot: 'bg-base-content/50' },
}

const sorted = computed(() =>
  [...props.events].sort(
    (a, b) => new Date(b.at).getTime() - new Date(a.at).getTime(),
  ),
)
</script>

<template>
  <div>
    <div
      v-if="sorted.length === 0"
      class="rounded-lg border border-dashed border-base-300 bg-base-100 p-6 text-center text-sm text-base-content/50"
    >
      {{ emptyText ?? '暂无事件记录。' }}
    </div>

    <ol v-else class="relative space-y-4 pl-6">
      <span
        aria-hidden="true"
        class="pointer-events-none absolute top-2 bottom-2 left-[7px] w-px bg-base-300"
      />
      <li
        v-for="ev in sorted"
        :key="ev.id"
        class="relative"
      >
        <span
          class="absolute top-1.5 -left-6 grid h-4 w-4 place-items-center"
        >
          <span
            :class="[
              'h-2.5 w-2.5 rounded-full ring-2 ring-base-100',
              TYPE_MAP[ev.type].dot,
            ]"
          />
        </span>

        <div
          class="rounded-lg border border-base-300 bg-base-100 px-3 py-2.5"
        >
          <div class="flex items-center justify-between gap-3 text-xs">
            <div class="flex items-center gap-2 text-base-content/60">
              <span class="font-medium text-base-content/80">
                {{ TYPE_MAP[ev.type].label }}
              </span>
              <span v-if="ev.actor">· {{ ev.actor.name }}</span>
            </div>
            <time class="text-base-content/50">
              {{ formatDateTime(ev.at) }}
            </time>
          </div>

          <p class="mt-1 text-sm text-base-content">{{ ev.title }}</p>
          <p v-if="ev.summary" class="mt-1 text-xs text-base-content/60">
            {{ ev.summary }}
          </p>

          <dl
            v-if="ev.meta && ev.meta.length > 0"
            class="mt-2 grid grid-cols-[auto_1fr] gap-x-3 gap-y-0.5 text-[11px]"
          >
            <template v-for="m in ev.meta" :key="m.label">
              <dt class="text-base-content/50">{{ m.label }}</dt>
              <dd class="text-base-content/80">{{ m.value }}</dd>
            </template>
          </dl>
        </div>
      </li>
    </ol>
  </div>
</template>
