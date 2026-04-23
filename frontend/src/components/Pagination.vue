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
    class="flex items-center justify-between gap-3 text-xs"
    aria-label="分页导航"
  >
    <span class="text-base-content/50">
      第 {{ page }} / {{ totalPages }} 页 · 共 {{ total }} 条
    </span>

    <div class="flex items-center gap-1">
      <button
        type="button"
        class="rounded-md border border-base-300 bg-base-100 px-2.5 py-1 text-base-content disabled:cursor-not-allowed disabled:opacity-40"
        :disabled="page <= 1"
        @click="go(page - 1)"
      >
        上一页
      </button>

      <template v-for="(p, i) in pageList" :key="i">
        <span
          v-if="p === '...'"
          class="px-1 text-base-content/40"
          aria-hidden="true"
        >…</span>
        <button
          v-else
          type="button"
          class="min-w-[32px] rounded-md border px-2 py-1 text-center"
          :class="
            p === page
              ? 'border-primary/60 bg-primary/10 text-primary font-medium'
              : 'border-base-300 bg-base-100 text-base-content hover:border-primary/40 hover:text-primary'
          "
          @click="go(p)"
        >
          {{ p }}
        </button>
      </template>

      <button
        type="button"
        class="rounded-md border border-base-300 bg-base-100 px-2.5 py-1 text-base-content disabled:cursor-not-allowed disabled:opacity-40"
        :disabled="page >= totalPages"
        @click="go(page + 1)"
      >
        下一页
      </button>
    </div>
  </nav>
</template>
