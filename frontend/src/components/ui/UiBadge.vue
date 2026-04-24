<script setup lang="ts">
import { computed } from 'vue'

type Tone =
  | 'neutral'
  | 'primary'
  | 'accent'
  | 'success'
  | 'warning'
  | 'error'
  | 'info'
  | 'secondary'

const props = withDefaults(
  defineProps<{
    tone?: Tone
    size?: 'xs' | 'sm' | 'md'
    dot?: boolean
    soft?: boolean
  }>(),
  { tone: 'neutral', size: 'sm', soft: true },
)

const toneClass = computed(() => {
  if (!props.soft) {
    switch (props.tone) {
      case 'primary':
        return 'bg-primary text-primary-content'
      case 'accent':
        return 'bg-accent text-accent-content'
      case 'success':
        return 'bg-success text-success-content'
      case 'warning':
        return 'bg-warning text-warning-content'
      case 'error':
        return 'bg-error text-error-content'
      case 'info':
        return 'bg-info text-info-content'
      case 'secondary':
        return 'bg-secondary text-secondary-content'
      default:
        return 'bg-base-300 text-base-content'
    }
  }
  switch (props.tone) {
    case 'primary':
      return 'bg-primary/12 text-primary ring-1 ring-primary/20'
    case 'accent':
      return 'bg-accent/15 text-[color-mix(in_oklch,var(--color-accent)_60%,var(--color-base-content))] ring-1 ring-accent/25'
    case 'success':
      return 'bg-success/12 text-success ring-1 ring-success/25'
    case 'warning':
      return 'bg-warning/20 text-[color-mix(in_oklch,var(--color-warning)_40%,var(--color-base-content))] ring-1 ring-warning/30'
    case 'error':
      return 'bg-error/12 text-error ring-1 ring-error/25'
    case 'info':
      return 'bg-info/12 text-info ring-1 ring-info/25'
    case 'secondary':
      return 'bg-secondary/12 text-secondary ring-1 ring-secondary/25'
    default:
      return 'bg-base-200 text-base-content/75 ring-1 ring-base-300'
  }
})

const sizeClass = computed(() => {
  switch (props.size) {
    case 'xs':
      return 'px-1.5 py-0.5 text-[10px] gap-1'
    case 'md':
      return 'px-2.5 py-1 text-xs gap-1.5'
    default:
      return 'px-2 py-0.5 text-[11px] gap-1.5'
  }
})
</script>

<template>
  <span
    :class="[
      'inline-flex items-center rounded-full font-medium whitespace-nowrap',
      toneClass,
      sizeClass,
    ]"
  >
    <span
      v-if="dot"
      class="h-1.5 w-1.5 shrink-0 rounded-full bg-current opacity-75"
      aria-hidden="true"
    />
    <slot />
  </span>
</template>
