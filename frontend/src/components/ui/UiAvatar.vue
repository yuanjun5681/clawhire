<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    name?: string | null
    size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
    tone?: 'primary' | 'accent' | 'neutral' | 'brand'
    ring?: boolean
  }>(),
  { size: 'md', tone: 'brand' },
)

const initial = computed(() => {
  const n = (props.name || '').trim()
  if (!n) return '—'
  // 中文取第一个字，英文取首字母
  return /^[A-Za-z]/.test(n) ? n.slice(0, 1).toUpperCase() : n.slice(0, 1)
})

const sizeClass = computed(() => {
  switch (props.size) {
    case 'xs':
      return 'h-6 w-6 text-[10px]'
    case 'sm':
      return 'h-8 w-8 text-xs'
    case 'lg':
      return 'h-12 w-12 text-base'
    case 'xl':
      return 'h-16 w-16 text-xl'
    default:
      return 'h-10 w-10 text-sm'
  }
})

const toneClass = computed(() => {
  switch (props.tone) {
    case 'primary':
      return 'bg-primary/15 text-primary'
    case 'accent':
      return 'bg-accent/20 text-[color-mix(in_oklch,var(--color-accent)_40%,var(--color-base-content))]'
    case 'neutral':
      return 'bg-base-200 text-base-content/70'
    default:
      return 'bg-[linear-gradient(135deg,color-mix(in_oklch,var(--color-primary)_25%,transparent),color-mix(in_oklch,var(--color-accent)_30%,transparent))] text-primary'
  }
})
</script>

<template>
  <span
    :class="[
      'grid shrink-0 place-items-center rounded-full font-semibold tracking-tight',
      sizeClass,
      toneClass,
      ring
        ? 'ring-2 ring-base-100 shadow-[0_0_0_1px_color-mix(in_oklch,var(--color-primary)_25%,transparent)]'
        : '',
    ]"
    aria-hidden="true"
  >
    {{ initial }}
  </span>
</template>
