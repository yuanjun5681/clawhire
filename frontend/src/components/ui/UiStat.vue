<script setup lang="ts">
withDefaults(
  defineProps<{
    label: string
    value: string | number
    suffix?: string
    hint?: string
    tone?: 'default' | 'primary' | 'accent' | 'success'
    icon?: boolean
  }>(),
  { tone: 'default' },
)
</script>

<template>
  <div
    :class="[
      'relative overflow-hidden rounded-box border p-4 surface-hover',
      'border-base-300/80 bg-base-100',
    ]"
  >
    <span
      aria-hidden="true"
      :class="[
        'pointer-events-none absolute -right-10 -top-10 h-32 w-32 rounded-full blur-2xl opacity-40',
        tone === 'primary' && 'bg-primary/40',
        tone === 'accent' && 'bg-accent/40',
        tone === 'success' && 'bg-success/40',
        tone === 'default' && 'bg-base-content/10',
      ]"
    />

    <div class="relative flex items-start justify-between gap-3">
      <div class="min-w-0">
        <p
          class="text-[10.5px] uppercase tracking-[0.12em] font-semibold text-base-content/55"
        >
          {{ label }}
        </p>
        <p
          class="mt-2 text-3xl font-semibold tracking-tight text-base-content leading-none"
        >
          {{ value }}
          <span v-if="suffix" class="ml-1 text-sm font-medium text-base-content/55">{{ suffix }}</span>
        </p>
        <p v-if="hint" class="mt-1.5 text-xs text-base-content/55">{{ hint }}</p>
      </div>
      <div
        v-if="icon"
        :class="[
          'grid h-10 w-10 shrink-0 place-items-center rounded-xl',
          tone === 'primary' && 'bg-primary/15 text-primary',
          tone === 'accent' && 'bg-accent/20 text-accent-content',
          tone === 'success' && 'bg-success/15 text-success',
          tone === 'default' && 'bg-base-200 text-base-content/70',
        ]"
      >
        <slot name="icon" />
      </div>
    </div>
  </div>
</template>
