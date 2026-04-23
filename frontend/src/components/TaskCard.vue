<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import StatusBadge from './StatusBadge.vue'
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
    class="group block rounded-xl border border-base-300 bg-base-100 p-4 transition hover:border-primary/40 hover:shadow-sm"
  >
    <div class="flex items-start justify-between gap-3">
      <h3
        class="min-w-0 flex-1 truncate text-[15px] font-medium text-base-content group-hover:text-primary"
      >
        {{ task.title }}
      </h3>
      <StatusBadge :status="task.status" />
    </div>

    <div
      v-if="!compact"
      class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-base-content/60"
    >
      <span class="rounded bg-base-200 px-1.5 py-0.5">{{ task.category }}</span>
      <span>需求方 · {{ task.requester.name }}</span>
    </div>

    <div
      class="mt-3 flex flex-wrap items-center justify-between gap-2 border-t border-base-200 pt-3 text-xs"
    >
      <div class="flex items-center gap-4">
        <span class="text-base-content">
          <span class="font-medium">
            {{ formatReward(task.reward.amount, task.reward.currency) }}
          </span>
        </span>
        <span v-if="deadlineText" class="text-base-content/60">
          截止 {{ deadlineText }}
        </span>
      </div>
      <span v-if="activityText" class="text-base-content/40">
        {{ activityText }}活跃
      </span>
    </div>
  </RouterLink>
</template>
