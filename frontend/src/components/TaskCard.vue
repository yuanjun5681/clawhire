<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import StatusBadge from './StatusBadge.vue'
import UiBadge from './ui/UiBadge.vue'
import UiAvatar from './ui/UiAvatar.vue'
import { formatDate, formatRelative, formatReward } from '@/utils/format'
import type { TaskListItem } from '@/types'

const props = defineProps<{
  task: TaskListItem
  compact?: boolean
}>()

const deadlineText = computed(() =>
  props.task.deadline ? formatDate(props.task.deadline) : null,
)
const activityText = computed(() =>
  props.task.lastActivityAt ? formatRelative(props.task.lastActivityAt) : null,
)
</script>

<template>
  <RouterLink
    :to="`/tasks/${task.taskId}`"
    class="group relative block overflow-hidden rounded-box border border-base-300/70 bg-base-100 p-5 surface-hover hover:-translate-y-0.5 hover:border-primary/40 hover:shadow-[0_16px_40px_-18px_color-mix(in_oklch,var(--color-primary)_35%,transparent)]"
  >
    <!-- decorative corner glow -->
    <span
      aria-hidden="true"
      class="pointer-events-none absolute -right-16 -top-20 h-40 w-40 rounded-full bg-primary/12 blur-3xl opacity-0 transition-opacity duration-300 group-hover:opacity-100"
    />

    <div class="flex items-start justify-between gap-3">
      <h3
        class="line-clamp-2 min-w-0 flex-1 text-[15.5px] font-semibold leading-snug tracking-tight text-base-content group-hover:text-primary"
      >
        {{ task.title }}
      </h3>
      <StatusBadge :status="task.status" />
    </div>

    <div
      v-if="!compact"
      class="mt-3 flex flex-wrap items-center gap-2 text-xs text-base-content/60"
    >
      <UiBadge tone="neutral" size="xs">
        <svg
          class="h-3 w-3"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.8"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="M4 7h16M4 12h16M4 17h10" />
        </svg>
        {{ task.category }}
      </UiBadge>
      <span class="inline-flex items-center gap-1.5">
        <UiAvatar :name="task.requester.name" size="xs" tone="neutral" />
        <span class="truncate">{{ task.requester.name }}</span>
      </span>
    </div>

    <div
      class="mt-4 flex flex-wrap items-center justify-between gap-2 border-t border-base-300/60 pt-3 text-xs"
    >
      <div class="flex items-center gap-4">
        <span class="inline-flex items-baseline gap-1 text-base-content">
          <span class="text-[10px] uppercase tracking-[0.12em] text-base-content/45">
            REWARD
          </span>
          <span class="text-[15px] font-semibold tracking-tight gradient-text">
            {{ formatReward(task.reward.amount, task.reward.currency) }}
          </span>
        </span>
        <span v-if="deadlineText" class="inline-flex items-center gap-1 text-base-content/55">
          <svg
            class="h-3 w-3"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="1.8"
            stroke-linecap="round"
            stroke-linejoin="round"
            aria-hidden="true"
          >
            <rect x="3" y="4" width="18" height="18" rx="2" />
            <path d="M16 2v4M8 2v4M3 10h18" />
          </svg>
          {{ deadlineText }}
        </span>
      </div>
      <span v-if="activityText" class="text-base-content/45">
        {{ activityText }}活跃
      </span>
    </div>
  </RouterLink>
</template>
