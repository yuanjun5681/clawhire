<script setup lang="ts">
withDefaults(
  defineProps<{
    required?: boolean
    for?: string
  }>(),
  {},
)
</script>

<template>
  <label
    :for="$props.for"
    class="flex items-center gap-1 text-[11px] font-medium uppercase tracking-[0.08em] text-base-content/60"
  >
    <slot />
    <span v-if="required" class="text-error">*</span>
  </label>
</template>
