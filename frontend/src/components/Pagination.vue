<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{
  'update:page': [page: number]
}>()

const totalPages = computed(() =>
  Math.max(1, Math.ceil(props.total / Math.max(1, props.pageSize))),
)

function range(start: number, end: number): number[] {
  return Array.from({ length: end - start + 1 }, (_, i) => start + i)
}

const pageList = computed<(number | '...')[]>(() => {
  const current = props.page
  const total = totalPages.value
  if (total <= 7) return range(1, total)
  if (current <= 4) return [...range(1, 5), '...', total]
  if (current >= total - 3) return [1, '...', ...range(total - 4, total)]
  return [1, '...', current - 1, current, current + 1, '...', total]
})

function go(p: number) {
  if (p < 1 || p > totalPages.value || p === props.page) return
  emit('update:page', p)
}
</script>

<template>
  <nav
    v-if="total > 0"
    class="flex flex-wrap items-center justify-between gap-3 rounded-box border border-base-300/60 bg-base-100/60 px-3 py-2 text-xs backdrop-blur"
    aria-label="分页导航"
  >
    <span class="text-base-content/55">
      第 <span class="font-semibold text-base-content">{{ page }}</span> /
      {{ totalPages }} 页 · 共 {{ total }} 条
    </span>

    <div class="flex items-center gap-1">
      <button
        type="button"
        class="inline-flex h-8 items-center gap-1 rounded-full border border-base-300/70 bg-base-100 px-3 text-base-content transition hover:border-primary/40 hover:text-primary disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:border-base-300"
        :disabled="page <= 1"
        @click="go(page - 1)"
      >
        <svg class="h-3 w-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6" /></svg>
        上一页
      </button>

      <template v-for="(p, i) in pageList" :key="i">
        <span
          v-if="p === '...'"
          class="px-1.5 text-base-content/40"
          aria-hidden="true"
        >…</span>
        <button
          v-else
          type="button"
          :class="[
            'grid h-8 min-w-[32px] place-items-center rounded-full px-2 text-center transition',
            p === page
              ? 'bg-[linear-gradient(120deg,var(--color-primary),var(--color-accent))] text-primary-content shadow-[0_4px_14px_-4px_color-mix(in_oklch,var(--color-primary)_70%,transparent)]'
              : 'border border-base-300/70 bg-base-100 text-base-content hover:border-primary/40 hover:text-primary',
          ]"
          @click="go(p)"
        >
          {{ p }}
        </button>
      </template>

      <button
        type="button"
        class="inline-flex h-8 items-center gap-1 rounded-full border border-base-300/70 bg-base-100 px-3 text-base-content transition hover:border-primary/40 hover:text-primary disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:border-base-300"
        :disabled="page >= totalPages"
        @click="go(page + 1)"
      >
        下一页
        <svg class="h-3 w-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6" /></svg>
      </button>
    </div>
  </nav>
</template>
