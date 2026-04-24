<script setup lang="ts">
import UiButton from './ui/UiButton.vue'

defineProps<{
  title: string
  description?: string
  actionLabel?: string
}>()

defineEmits<{
  action: []
}>()
</script>

<template>
  <div
    class="relative flex flex-col items-center justify-center gap-3 overflow-hidden rounded-box border border-dashed border-base-300/70 bg-base-100/60 px-6 py-14 text-center"
  >
    <span
      aria-hidden="true"
      class="pointer-events-none absolute inset-0 bg-[radial-gradient(500px_260px_at_50%_-20%,color-mix(in_oklch,var(--color-primary)_14%,transparent),transparent_60%)]"
    />
    <div
      class="relative grid h-14 w-14 place-items-center rounded-full bg-[linear-gradient(135deg,color-mix(in_oklch,var(--color-primary)_14%,transparent),color-mix(in_oklch,var(--color-accent)_18%,transparent))] text-primary"
      aria-hidden="true"
    >
      <svg
        class="h-6 w-6"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="1.6"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <rect x="3" y="4" width="18" height="16" rx="3" />
        <path d="M8 2v4M16 2v4M3 10h18" />
        <circle cx="12" cy="15" r="2" />
      </svg>
    </div>
    <div class="relative space-y-1">
      <p class="text-sm font-semibold text-base-content">{{ title }}</p>
      <p
        v-if="description"
        class="max-w-sm text-xs text-base-content/55"
      >
        {{ description }}
      </p>
    </div>
    <UiButton
      v-if="actionLabel"
      variant="outline"
      size="sm"
      @click="$emit('action')"
    >
      {{ actionLabel }}
    </UiButton>
  </div>
</template>
