<script setup lang="ts">
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
    class="rounded-xl border border-red-200 bg-red-50 p-6 text-sm text-red-800"
  >
    <div class="flex items-start gap-3">
      <div
        class="grid h-8 w-8 shrink-0 place-items-center rounded-full bg-red-100 text-red-600"
        aria-hidden="true"
      >
        <svg
          class="h-4 w-4"
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
        <p class="font-medium text-red-800">加载失败</p>
        <p class="text-xs text-red-700/90 break-all">{{ message }}</p>
        <p v-if="code" class="font-mono text-[11px] text-red-600/70">
          错误码 · {{ code }}
        </p>
      </div>
      <button
        type="button"
        class="shrink-0 rounded-md border border-red-200 bg-white px-3 py-1.5 text-xs text-red-700 hover:bg-red-100"
        @click="$emit('retry')"
      >
        {{ retryLabel ?? '重试' }}
      </button>
    </div>
  </div>
</template>
