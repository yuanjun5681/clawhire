<script setup lang="ts">
import UiButton from './ui/UiButton.vue'

defineProps<{
  message: string
  code?: string
  retryLabel?: string
}>()

defineEmits<{
  retry: []
}>()
</script>

<template>
  <div
    class="relative overflow-hidden rounded-box border border-error/30 bg-error/5 p-5 text-sm text-error ring-1 ring-error/10"
  >
    <span
      aria-hidden="true"
      class="pointer-events-none absolute -right-16 -top-16 h-40 w-40 rounded-full bg-error/15 blur-3xl"
    />
    <div class="relative flex items-start gap-3">
      <div
        class="grid h-10 w-10 shrink-0 place-items-center rounded-full bg-error/15 text-error ring-1 ring-error/25"
        aria-hidden="true"
      >
        <svg
          class="h-4.5 w-4.5"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="M12 9v4M12 17h.01" />
          <path
            d="M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"
          />
        </svg>
      </div>
      <div class="min-w-0 flex-1 space-y-1">
        <p class="font-semibold">加载失败</p>
        <p class="text-xs text-error/85 break-words">{{ message }}</p>
        <p v-if="code" class="font-mono text-[11px] text-error/70">
          错误码 · {{ code }}
        </p>
      </div>
      <UiButton
        variant="outline"
        size="sm"
        @click="$emit('retry')"
      >
        <svg
          class="h-3.5 w-3.5"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="M3 12a9 9 0 1 0 3-6.7" />
          <polyline points="3 4 3 10 9 10" />
        </svg>
        {{ retryLabel ?? '重试' }}
      </UiButton>
    </div>
  </div>
</template>
