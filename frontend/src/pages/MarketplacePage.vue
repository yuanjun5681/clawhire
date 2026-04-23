<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import FilterBar from '@/components/FilterBar.vue'
import TaskCard from '@/components/TaskCard.vue'
import SkeletonTaskCard from '@/components/SkeletonTaskCard.vue'
import EmptyState from '@/components/EmptyState.vue'
import ErrorState from '@/components/ErrorState.vue'
import Pagination from '@/components/Pagination.vue'
import { ApiRequestError, tasksApi } from '@/api'
import type { Paginated, TaskListItem, TaskQuery } from '@/types'

const PAGE_SIZE = 6

const query = ref<TaskQuery>({ page: 1, pageSize: PAGE_SIZE })
const loading = ref(true)
const apiError = ref<{ message: string; code?: string } | null>(null)
const result = ref<Paginated<TaskListItem> | null>(null)

const categories = ['coding', 'writing', 'research', 'consulting']

const hasActiveFilter = computed(() =>
  Boolean(query.value.status || query.value.category || query.value.keyword),
)
const resultLabel = computed(() =>
  result.value ? `共 ${result.value.total} 条` : '',
)

async function load() {
  loading.value = true
  apiError.value = null
  try {
    result.value = await tasksApi.listTasks(query.value)
  } catch (e: unknown) {
    if (e instanceof ApiRequestError) {
      apiError.value = { message: e.message, code: e.code }
    } else {
      apiError.value = {
        message: e instanceof Error ? e.message : String(e),
      }
    }
    result.value = null
  } finally {
    loading.value = false
  }
}

function clearFilters() {
  query.value = { page: 1, pageSize: PAGE_SIZE }
}

function changePage(p: number) {
  query.value = { ...query.value, page: p }
}

onMounted(load)
watch(query, load, { deep: true })
</script>

<template>
  <section class="space-y-4">
    <header class="space-y-1">
      <h1 class="text-2xl font-semibold tracking-tight">任务大厅</h1>
      <p class="text-sm text-base-content/60">浏览和筛选开放任务。</p>
    </header>

    <FilterBar
      v-model="query"
      :categories="categories"
      :result-label="resultLabel"
    />

    <div
      v-if="loading"
      class="grid grid-cols-1 gap-3 md:grid-cols-2"
    >
      <SkeletonTaskCard v-for="i in PAGE_SIZE" :key="i" />
    </div>

    <ErrorState
      v-else-if="apiError"
      :message="apiError.message"
      :code="apiError.code"
      @retry="load"
    />

    <EmptyState
      v-else-if="result && result.items.length === 0 && hasActiveFilter"
      title="没有匹配的任务"
      description="换一个关键词或清除筛选后再试试。"
      action-label="清除筛选"
      @action="clearFilters"
    />

    <EmptyState
      v-else-if="result && result.items.length === 0"
      title="当前暂无可浏览任务"
      description="稍后再来看看，或在你自己的账号中发布一个任务。"
    />

    <template v-else-if="result">
      <ul class="grid grid-cols-1 gap-3 md:grid-cols-2">
        <li v-for="t in result.items" :key="t.taskId">
          <TaskCard :task="t" />
        </li>
      </ul>
      <Pagination
        :page="result.page"
        :page-size="result.pageSize"
        :total="result.total"
        @update:page="changePage"
      />
    </template>
  </section>
</template>
