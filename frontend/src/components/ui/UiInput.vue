<script setup lang="ts">
import { computed, useAttrs } from 'vue'

defineOptions({ inheritAttrs: false })

const props = withDefaults(
  defineProps<{
    modelValue?: string | number | null
    label?: string
    hint?: string
    error?: string
    size?: 'sm' | 'md' | 'lg'
    type?: string
    placeholder?: string
    prefixIcon?: boolean
    suffixIcon?: boolean
    required?: boolean
  }>(),
  { size: 'md', type: 'text' },
)

const emit = defineEmits<{ 'update:modelValue': [v: string] }>()

const attrs = useAttrs()

const sizeClass = computed(() => {
  switch (props.size) {
    case 'sm':
      return 'h-9 text-sm px-3'
    case 'lg':
      return 'h-12 text-[15px] px-4'
    default:
      return 'h-11 text-sm px-3.5'
  }
})

function onInput(e: Event) {
  emit('update:modelValue', (e.target as HTMLInputElement).value)
}
</script>

<template>
  <label class="flex flex-col gap-1.5">
    <span
      v-if="label"
      class="text-[11px] font-medium uppercase tracking-[0.08em] text-base-content/60"
    >
      {{ label }}
      <span v-if="required" class="text-error">*</span>
    </span>

    <span class="relative block">
      <span
        v-if="prefixIcon"
        class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-base-content/40"
      >
        <slot name="prefix" />
      </span>

      <input
        :value="modelValue ?? ''"
        :type="type"
        :placeholder="placeholder"
        :class="[
          'peer w-full rounded-field bg-base-100 text-base-content placeholder:text-base-content/40',
          'border border-base-300/80',
          'transition-[border-color,box-shadow,background-color] duration-200',
          'focus:outline-none focus:border-primary/60',
          'focus:ring-[3px] focus:ring-primary/15',
          'hover:border-base-content/20',
          sizeClass,
          prefixIcon ? 'pl-10' : '',
          suffixIcon ? 'pr-10' : '',
          error ? '!border-error/70 focus:!ring-error/20' : '',
        ]"
        v-bind="attrs"
        @input="onInput"
      />

      <span
        v-if="suffixIcon"
        class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-base-content/40"
      >
        <slot name="suffix" />
      </span>
    </span>

    <p v-if="error" class="text-xs text-error">{{ error }}</p>
    <p v-else-if="hint" class="text-xs text-base-content/50">{{ hint }}</p>
  </label>
</template>
