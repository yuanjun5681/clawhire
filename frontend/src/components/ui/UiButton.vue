<script setup lang="ts">
import { computed } from 'vue'

type Variant = 'primary' | 'secondary' | 'ghost' | 'outline' | 'danger' | 'accent'
type Size = 'xs' | 'sm' | 'md' | 'lg'

const props = withDefaults(
  defineProps<{
    variant?: Variant
    size?: Size
    block?: boolean
    loading?: boolean
    disabled?: boolean
    type?: 'button' | 'submit' | 'reset'
    icon?: boolean
  }>(),
  { variant: 'primary', size: 'md', type: 'button' },
)

defineEmits<{ click: [e: MouseEvent] }>()

const sizeClass = computed(() => {
  if (props.icon) {
    return {
      xs: 'h-7 w-7',
      sm: 'h-8 w-8',
      md: 'h-10 w-10',
      lg: 'h-12 w-12',
    }[props.size]
  }
  return {
    xs: 'h-7 px-2.5 text-xs gap-1',
    sm: 'h-8 px-3 text-xs gap-1.5',
    md: 'h-10 px-4 text-sm gap-2',
    lg: 'h-12 px-6 text-[15px] gap-2',
  }[props.size]
})

const variantClass = computed(() => {
  switch (props.variant) {
    case 'primary':
      return [
        'text-primary-content',
        'bg-[linear-gradient(120deg,var(--color-primary)_0%,color-mix(in_oklch,var(--color-primary)_80%,var(--color-accent))_100%)]',
        'shadow-[0_6px_20px_-8px_color-mix(in_oklch,var(--color-primary)_60%,transparent)]',
        'hover:brightness-110 active:brightness-95',
        'ring-1 ring-inset ring-white/10',
      ].join(' ')
    case 'accent':
      return 'bg-accent text-accent-content hover:brightness-105 shadow-[0_4px_14px_-6px_color-mix(in_oklch,var(--color-accent)_60%,transparent)]'
    case 'secondary':
      return 'bg-base-200 text-base-content hover:bg-base-300'
    case 'ghost':
      return 'text-base-content hover:bg-base-200'
    case 'outline':
      return 'border border-base-300 text-base-content hover:border-primary/40 hover:text-primary bg-base-100/60'
    case 'danger':
      return 'bg-error text-error-content hover:brightness-110 shadow-[0_4px_14px_-6px_color-mix(in_oklch,var(--color-error)_60%,transparent)]'
  }
  return ''
})
</script>

<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    :class="[
      'relative inline-flex items-center justify-center rounded-field font-medium tracking-tight transition-[transform,filter,background-color,color,border-color,box-shadow] duration-200',
      'select-none whitespace-nowrap',
      'disabled:cursor-not-allowed disabled:opacity-55 disabled:shadow-none',
      'active:scale-[0.98]',
      block ? 'w-full' : '',
      sizeClass,
      variantClass,
    ]"
    @click="$emit('click', $event)"
  >
    <span
      v-if="loading"
      class="absolute inset-0 grid place-items-center"
      aria-hidden="true"
    >
      <span
        class="h-4 w-4 animate-spin rounded-full border-2 border-current border-b-transparent"
      />
    </span>
    <span
      :class="[
        'inline-flex items-center justify-center',
        icon ? '' : 'gap-[inherit]',
        loading ? 'opacity-0' : 'opacity-100',
      ]"
    >
      <slot />
    </span>
  </button>
</template>
