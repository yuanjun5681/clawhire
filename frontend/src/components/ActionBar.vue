<script setup lang="ts">
import UiButton from './ui/UiButton.vue'

export interface ActionItem {
  key: string
  label: string
  disabledReason?: string
  primary?: boolean
  danger?: boolean
}

const props = defineProps<{
  stageLabel?: string
  stageHint?: string
  actions: ActionItem[]
  emptyHint?: string
}>()

defineEmits<{
  run: [key: string]
}>()

function variant(a: ActionItem) {
  if (a.primary) return 'primary' as const
  if (a.danger) return 'danger' as const
  return 'outline' as const
}
</script>

<template>
  <section
    class="relative overflow-hidden rounded-box border border-base-300/70 bg-base-100 p-5"
  >
    <span
      aria-hidden="true"
      class="pointer-events-none absolute -right-20 -top-20 h-40 w-40 rounded-full bg-primary/10 blur-3xl"
    />
    <div class="relative flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <header v-if="stageLabel || stageHint" class="min-w-0 space-y-1">
        <p
          v-if="stageLabel"
          class="text-[10.5px] uppercase tracking-[0.12em] text-base-content/55 font-semibold"
        >
          {{ stageLabel }}
        </p>
        <p v-if="stageHint" class="text-sm text-base-content/75 leading-relaxed">
          {{ stageHint }}
        </p>
      </header>

      <div v-if="actions.length > 0" class="flex shrink-0 flex-wrap gap-2 sm:justify-end">
        <UiButton
          v-for="a in actions"
          :key="a.key"
          :variant="variant(a)"
          size="sm"
          :disabled="Boolean(a.disabledReason)"
          :title="a.disabledReason ?? undefined"
          @click="$emit('run', a.key)"
        >
          {{ a.label }}
        </UiButton>
      </div>
      <p
        v-else-if="props.emptyHint"
        class="inline-flex items-center gap-1.5 text-xs text-base-content/50"
      >
        <svg
          class="h-3.5 w-3.5"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.8"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <circle cx="12" cy="12" r="10" />
          <path d="M12 8v4M12 16h.01" />
        </svg>
        {{ props.emptyHint }}
      </p>
    </div>
  </section>
</template>
