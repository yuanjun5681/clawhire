<script setup lang="ts">
const props = defineProps<{
  modelValue: boolean
  label?: string
  description?: string
  disabled?: boolean
}>()

const emit = defineEmits<{ 'update:modelValue': [v: boolean] }>()

function toggle() {
  if (props.disabled) return
  emit('update:modelValue', !props.modelValue)
}
</script>

<template>
  <label
    :class="[
      'group inline-flex cursor-pointer items-center gap-3',
      disabled ? 'cursor-not-allowed opacity-60' : '',
    ]"
  >
    <button
      type="button"
      role="switch"
      :aria-checked="modelValue"
      :disabled="disabled"
      :class="[
        'relative inline-flex h-6 w-11 shrink-0 items-center rounded-full transition-colors duration-300',
        modelValue
          ? 'bg-[linear-gradient(120deg,var(--color-primary),var(--color-accent))]'
          : 'bg-base-300',
      ]"
      @click="toggle"
    >
      <span
        :class="[
          'inline-block h-5 w-5 transform rounded-full bg-white shadow-md transition-transform duration-300',
          modelValue ? 'translate-x-5' : 'translate-x-0.5',
        ]"
      />
    </button>
    <span v-if="label || description" class="flex flex-col">
      <span v-if="label" class="text-sm font-medium text-base-content">{{ label }}</span>
      <span v-if="description" class="text-xs text-base-content/60">{{ description }}</span>
    </span>
  </label>
</template>
