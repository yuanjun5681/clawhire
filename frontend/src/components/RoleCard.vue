<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import UiAvatar from './ui/UiAvatar.vue'
import UiBadge from './ui/UiBadge.vue'
import type { AccountSummary } from '@/types'

const props = defineProps<{
  role: '需求方' | '执行方' | '验收方' | string
  account?: AccountSummary | null
  hint?: string
}>()

const isAgent = computed(() => props.account?.kind === 'agent')
</script>

<template>
  <div
    :class="[
      'relative overflow-hidden rounded-box border border-base-300/70 bg-base-100 p-4 transition hover:border-primary/30',
      isAgent
        ? 'ring-1 ring-primary/20 shadow-[inset_0_0_0_1px_color-mix(in_oklch,var(--color-primary)_20%,transparent)]'
        : '',
    ]"
  >
    <span
      v-if="isAgent"
      aria-hidden="true"
      class="pointer-events-none absolute -right-10 -top-12 h-28 w-28 rounded-full bg-primary/12 blur-3xl"
    />
    <div class="relative flex items-center justify-between text-[10.5px] uppercase tracking-[0.12em] text-base-content/55">
      <span class="font-semibold">{{ role }}</span>
      <UiBadge
        v-if="account"
        :tone="isAgent ? 'primary' : 'neutral'"
        size="xs"
      >
        {{ account.kind }}
      </UiBadge>
    </div>

    <div v-if="account" class="relative mt-2.5 flex items-center gap-3">
      <UiAvatar
        :name="account.name"
        size="md"
        :tone="isAgent ? 'primary' : 'neutral'"
      />
      <div class="min-w-0 flex-1">
        <RouterLink
          :to="`/accounts/${account.id}`"
          class="block truncate text-sm font-medium text-base-content hover:text-primary"
        >
          {{ account.name }}
        </RouterLink>
        <div
          v-if="isAgent && account.nodeId"
          class="truncate font-mono text-[11px] text-base-content/55"
        >
          node · {{ account.nodeId }}
        </div>
        <div
          v-else-if="hint"
          class="truncate text-[11px] text-base-content/55"
        >
          {{ hint }}
        </div>
      </div>
    </div>

    <p
      v-else
      class="relative mt-2 rounded-field border border-dashed border-base-300/80 px-3 py-2 text-xs text-base-content/45"
    >
      {{ hint ?? '未分配' }}
    </p>
  </div>
</template>
