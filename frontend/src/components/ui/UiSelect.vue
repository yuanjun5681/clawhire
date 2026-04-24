<script setup lang="ts">
import { useAttrs } from 'vue'

defineOptions({ inheritAttrs: false })

const props = defineProps<{
  modelValue?: string | number | null
  label?: string
  hint?: string
  size?: 'sm' | 'md' | 'lg'
  placeholder?: string
  required?: boolean
  options?: Array<{ label: string; value: string | number } | string>
}>()

const emit = defineEmits<{ 'update:modelValue': [v: string] }>()
const attrs = useAttrs()

function onChange(e: Event) {
  emit('update:modelValue', (e.target as HTMLSelectElement).value)
}

function sizeClass() {
  switch (props.size) {
    case 'sm':
      return 'h-9 text-sm pl-3 pr-9'
    case 'lg':
      return 'h-12 text-[15px] pl-4 pr-10'
    default:
      return 'h-11 text-sm pl-3.5 pr-9'
  }
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
      <select
        :value="modelValue ?? ''"
        :class="[
          'w-full appearance-none rounded-field border border-base-300/80 bg-base-100 text-base-content',
          'transition-[border-color,box-shadow] duration-200',
          'focus:outline-none focus:border-primary/60 focus:ring-[3px] focus:ring-primary/15',
          'hover:border-base-content/20',
          sizeClass(),
        ]"
        v-bind="attrs"
        @change="onChange"
      >
        <option v-if="placeholder" value="" disabled>{{ placeholder }}</option>
        <slot>
          <template v-if="options">
            <template v-for="opt in options" :key="typeof opt === 'string' ? opt : opt.value">
              <option
                :value="typeof opt === 'string' ? opt : opt.value"
              >
                {{ typeof opt === 'string' ? opt : opt.label }}
              </option>
            </template>
          </template>
        </slot>
      </select>
      <svg
        class="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/40"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <polyline points="6 9 12 15 18 9" />
      </svg>
    </span>
    <p v-if="hint" class="text-xs text-base-content/50">{{ hint }}</p>
  </label>
</template>
