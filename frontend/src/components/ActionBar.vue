<script setup lang="ts">
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

const emit = defineEmits<{
  run: [key: string]
}>()

function classFor(a: ActionItem): string {
  const base =
    'inline-flex items-center justify-center rounded-md px-3 py-1.5 text-sm transition disabled:cursor-not-allowed disabled:opacity-50'
  if (a.primary) {
    return `${base} bg-primary text-primary-content hover:bg-primary/90 disabled:bg-primary/60`
  }
  if (a.danger) {
    return `${base} border border-red-200 bg-red-50 text-red-700 hover:bg-red-100`
  }
  return `${base} border border-base-300 bg-base-100 text-base-content hover:border-primary/40 hover:text-primary`
}
</script>

<template>
  <section
    class="rounded-xl border border-base-300 bg-base-100 p-4"
  >
    <header v-if="stageLabel || stageHint" class="mb-3 space-y-0.5">
      <p
        v-if="stageLabel"
        class="text-[11px] uppercase tracking-wider text-base-content/50"
      >
        {{ stageLabel }}
      </p>
      <p v-if="stageHint" class="text-sm text-base-content/70">
        {{ stageHint }}
      </p>
    </header>

    <div v-if="actions.length > 0" class="flex flex-wrap gap-2">
      <button
        v-for="a in actions"
        :key="a.key"
        type="button"
        :class="classFor(a)"
        :disabled="Boolean(a.disabledReason)"
        :title="a.disabledReason ?? undefined"
        @click="emit('run', a.key)"
      >
        {{ a.label }}
      </button>
    </div>
    <p v-else-if="props.emptyHint" class="text-sm text-base-content/50">
      {{ props.emptyHint }}
    </p>
  </section>
</template>
