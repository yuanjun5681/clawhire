<script setup lang="ts">
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
    class="flex flex-col items-center justify-center gap-3 rounded-xl border border-dashed border-base-300 bg-base-100 px-6 py-12 text-center"
  >
    <div
      class="grid h-10 w-10 place-items-center rounded-full bg-base-200 text-base-content/40"
      aria-hidden="true"
    >
      <svg
        class="h-5 w-5"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <rect x="3" y="4" width="18" height="16" rx="2" />
        <path d="M8 2v4M16 2v4M3 10h18" />
      </svg>
    </div>
    <div class="space-y-0.5">
      <p class="text-sm font-medium text-base-content">{{ title }}</p>
      <p
        v-if="description"
        class="text-xs text-base-content/50"
      >
        {{ description }}
      </p>
    </div>
    <button
      v-if="actionLabel"
      type="button"
      class="inline-flex items-center rounded-md border border-base-300 bg-base-100 px-3 py-1.5 text-xs text-base-content hover:border-primary/40 hover:text-primary"
      @click="$emit('action')"
    >
      {{ actionLabel }}
    </button>
  </div>
</template>
