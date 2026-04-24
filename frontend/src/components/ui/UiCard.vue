<script setup lang="ts">
withDefaults(
  defineProps<{
    padding?: 'none' | 'sm' | 'md' | 'lg'
    tone?: 'base' | 'glass' | 'muted' | 'dashed'
    interactive?: boolean
  }>(),
  { padding: 'md', tone: 'base' },
)
</script>

<template>
  <div
    :class="[
      'rounded-box transition-[border-color,transform,box-shadow] duration-200',
      tone === 'base' &&
        'surface',
      tone === 'glass' &&
        'glass rounded-box shadow-[0_8px_30px_-12px_color-mix(in_oklch,var(--color-base-content)_14%,transparent)]',
      tone === 'muted' &&
        'bg-base-200/60 border border-base-300/70',
      tone === 'dashed' &&
        'bg-base-100/40 border border-dashed border-base-300',
      padding === 'sm' && 'p-3',
      padding === 'md' && 'p-5',
      padding === 'lg' && 'p-6',
      interactive &&
        'hover:-translate-y-0.5 hover:border-primary/40 hover:shadow-[0_12px_36px_-14px_color-mix(in_oklch,var(--color-primary)_30%,transparent)]',
    ]"
  >
    <slot />
  </div>
</template>
