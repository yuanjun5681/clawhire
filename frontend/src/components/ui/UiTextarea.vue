<script setup lang="ts">
import { useAttrs } from 'vue'

defineOptions({ inheritAttrs: false })

const props = defineProps<{
  modelValue?: string | null
  label?: string
  hint?: string
  error?: string
  placeholder?: string
  rows?: number
  required?: boolean
}>()

const emit = defineEmits<{ 'update:modelValue': [v: string] }>()

const attrs = useAttrs()

function onInput(e: Event) {
  emit('update:modelValue', (e.target as HTMLTextAreaElement).value)
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
    <textarea
      :value="modelValue ?? ''"
      :placeholder="placeholder"
      :rows="rows ?? 4"
      :class="[
        'w-full rounded-field bg-base-100 text-base-content placeholder:text-base-content/40',
        'border border-base-300/80 px-3.5 py-2.5 text-sm leading-relaxed',
        'transition-[border-color,box-shadow] duration-200',
        'focus:outline-none focus:border-primary/60 focus:ring-[3px] focus:ring-primary/15',
        'hover:border-base-content/20 resize-y min-h-[96px]',
        error ? '!border-error/70 focus:!ring-error/20' : '',
      ]"
      v-bind="attrs"
      @input="onInput"
    />
    <p v-if="error" class="text-xs text-error">{{ error }}</p>
    <p v-else-if="hint" class="text-xs text-base-content/50">{{ hint }}</p>
  </label>
</template>
