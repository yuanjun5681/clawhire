<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import type { AccountSummary } from '@/types'

const props = defineProps<{
  role: '需求方' | '执行方' | '验收方' | string
  account?: AccountSummary | null
  hint?: string
}>()

const initial = computed(() =>
  props.account?.name ? props.account.name.slice(0, 1) : '—',
)

const isAgent = computed(() => props.account?.kind === 'agent')
</script>

<template>
  <div
    class="rounded-lg border border-base-300 bg-base-100 p-3"
    :class="isAgent ? 'border-l-2 border-l-primary/60' : ''"
  >
    <div
      class="flex items-center justify-between text-[11px] uppercase tracking-wider text-base-content/50"
    >
      <span>{{ role }}</span>
      <span
        v-if="account"
        class="rounded bg-base-200 px-1.5 py-0.5 text-[10px] font-medium text-base-content/70"
      >{{ account.kind }}</span>
    </div>

    <div v-if="account" class="mt-2 flex items-center gap-2">
      <div
        class="grid h-8 w-8 place-items-center rounded-full text-sm font-medium"
        :class="
          isAgent
            ? 'bg-primary/10 text-primary'
            : 'bg-base-200 text-base-content/70'
        "
      >
        {{ initial }}
      </div>
      <div class="min-w-0 flex-1">
        <RouterLink
          :to="`/accounts/${account.id}`"
          class="block truncate text-sm font-medium text-base-content hover:text-primary"
        >
          {{ account.name }}
        </RouterLink>
        <div
          v-if="isAgent && account.nodeId"
          class="truncate font-mono text-[11px] text-base-content/50"
        >
          node · {{ account.nodeId }}
        </div>
        <div v-else-if="hint" class="truncate text-[11px] text-base-content/50">
          {{ hint }}
        </div>
      </div>
    </div>

    <p v-else class="mt-2 text-xs text-base-content/40">未分配</p>
  </div>
</template>
