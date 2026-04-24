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

const TYPE_MAP: Record<
  TimelineEventType,
  { label: string; bg: string; icon: string }
> = {
  bid: {
    label: '报价',
    bg: 'bg-secondary/15 text-secondary ring-secondary/30',
    icon: 'M12 1v22 M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6',
  },
  assign: {
    label: '指派',
    bg: 'bg-primary/15 text-primary ring-primary/30',
    icon: 'M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2 M9 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8z M22 11h-6 M19 8v6',
  },
  progress: {
    label: '进度',
    bg: 'bg-warning/20 text-[color-mix(in_oklch,var(--color-warning)_40%,var(--color-base-content))] ring-warning/40',
    icon: 'M22 12A10 10 0 0 1 2 12 M12 2v10l5 3',
  },
  submission: {
    label: '交付',
    bg: 'bg-accent/20 text-[color-mix(in_oklch,var(--color-accent)_35%,var(--color-base-content))] ring-accent/35',
    icon: 'M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4 M17 8l-5-5-5 5 M12 3v12',
  },
  review_approved: {
    label: '验收通过',
    bg: 'bg-success/15 text-success ring-success/30',
    icon: 'M20 6 9 17l-5-5',
  },
  review_rejected: {
    label: '验收驳回',
    bg: 'bg-error/15 text-error ring-error/30',
    icon: 'M18 6 6 18 M6 6l12 12',
  },
  settlement: {
    label: '结算',
    bg: 'bg-success/15 text-success ring-success/30',
    icon: 'M4 8V6a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2v2 M2 10h20v10a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V10z M6 15h.01M10 15h.01',
  },
  status: {
    label: '状态变更',
    bg: 'bg-base-300 text-base-content/75 ring-base-300',
    icon: 'M12 2a10 10 0 1 0 10 10 M22 4 12 14.01l-3-3',
  },
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
      class="rounded-box border border-dashed border-base-300/70 bg-base-100/60 p-8 text-center text-sm text-base-content/50"
    >
      {{ emptyText ?? '暂无事件记录。' }}
    </div>

    <ol v-else class="relative space-y-4 pl-8">
      <span
        aria-hidden="true"
        class="pointer-events-none absolute top-3 bottom-3 left-3 w-px bg-linear-to-b from-base-300/20 via-base-300 to-base-300/20"
      />
      <li
        v-for="ev in sorted"
        :key="ev.id"
        class="relative"
      >
        <span
          :class="[
            'absolute top-1 -left-8 grid h-6 w-6 place-items-center rounded-full ring-4 ring-base-100',
            TYPE_MAP[ev.type].bg,
          ]"
        >
          <svg
            class="h-3 w-3"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path :d="TYPE_MAP[ev.type].icon" />
          </svg>
        </span>

        <div
          class="rounded-box border border-base-300/70 bg-base-100 px-4 py-3 transition-colors hover:border-primary/30"
        >
          <div class="flex items-center justify-between gap-3 text-xs">
            <div class="flex items-center gap-2 text-base-content/60">
              <span class="font-semibold text-base-content/80">
                {{ TYPE_MAP[ev.type].label }}
              </span>
              <span v-if="ev.actor" class="text-base-content/50">· {{ ev.actor.name }}</span>
            </div>
            <time class="font-mono text-[11px] text-base-content/50">
              {{ formatDateTime(ev.at) }}
            </time>
          </div>

          <p class="mt-1.5 text-sm font-medium text-base-content">{{ ev.title }}</p>
          <p v-if="ev.summary" class="mt-1 text-xs text-base-content/60">
            {{ ev.summary }}
          </p>

          <dl
            v-if="ev.meta && ev.meta.length > 0"
            class="mt-2.5 grid grid-cols-[auto_1fr] gap-x-3 gap-y-1 text-[11px]"
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
