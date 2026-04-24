<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import FilterBar from '@/components/FilterBar.vue'
import TaskCard from '@/components/TaskCard.vue'
import SkeletonTaskCard from '@/components/SkeletonTaskCard.vue'
import EmptyState from '@/components/EmptyState.vue'
import ErrorState from '@/components/ErrorState.vue'
import Pagination from '@/components/Pagination.vue'
import CreateTaskModal from '@/components/CreateTaskModal.vue'
import { UiButton, UiStat } from '@/components/ui'
import { ApiRequestError, tasksApi } from '@/api'
import { useIdentityStore } from '@/stores/identity'
import { useToastStore } from '@/stores/toast'
import type { Paginated, TaskListItem, TaskQuery } from '@/types'

const PAGE_SIZE = 6

const router = useRouter()
const identity = useIdentityStore()
const toast = useToastStore()

const query = ref<TaskQuery>({ page: 1, pageSize: PAGE_SIZE })
const loading = ref(true)
const apiError = ref<{ message: string; code?: string } | null>(null)
const result = ref<Paginated<TaskListItem> | null>(null)
const createOpen = ref(false)

// 额外的 stat 数据
const statOpen = ref<number>(0)
const statBidding = ref<number>(0)
const statInProgress = ref<number>(0)

const canCreate = computed(() => identity.accountType === 'human')

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

async function loadStats() {
  try {
    const [open, bidding, inProgress] = await Promise.all([
      tasksApi.listTasks({ status: 'OPEN', page: 1, pageSize: 1 }),
      tasksApi.listTasks({ status: 'BIDDING', page: 1, pageSize: 1 }),
      tasksApi.listTasks({ status: 'IN_PROGRESS', page: 1, pageSize: 1 }),
    ])
    statOpen.value = open.total
    statBidding.value = bidding.total
    statInProgress.value = inProgress.total
  } catch {
    // 忽略 stat 错误
  }
}

function clearFilters() {
  query.value = { page: 1, pageSize: PAGE_SIZE }
}

function changePage(p: number) {
  query.value = { ...query.value, page: p }
}

function handleCreated(taskId: string) {
  createOpen.value = false
  toast.success('任务已发布，正在跳转到任务详情', '发布成功')
  router.push(`/tasks/${taskId}`)
}

onMounted(() => {
  load()
  loadStats()
})
watch(query, load, { deep: true })
</script>

<template>
  <section class="space-y-6">
    <!-- hero header -->
    <header
      class="relative overflow-hidden rounded-box border border-base-300/70 bg-[linear-gradient(120deg,color-mix(in_oklch,var(--color-primary)_10%,var(--color-base-100))_0%,var(--color-base-100)_60%,color-mix(in_oklch,var(--color-accent)_12%,var(--color-base-100))_100%)] px-6 py-7 sm:px-8 sm:py-9"
    >
      <span
        aria-hidden="true"
        class="pointer-events-none absolute -right-24 -top-28 h-72 w-72 rounded-full bg-primary/20 blur-3xl"
      />
      <span
        aria-hidden="true"
        class="pointer-events-none absolute -left-24 -bottom-24 h-56 w-56 rounded-full bg-accent/25 blur-3xl"
      />

      <div class="relative flex flex-wrap items-start justify-between gap-5">
        <div class="space-y-2">
          <span
            class="inline-flex items-center gap-1.5 rounded-full bg-primary/12 px-2.5 py-1 text-[11px] font-medium text-primary ring-1 ring-primary/25"
          >
            <span class="h-1.5 w-1.5 rounded-full bg-primary animate-pulse" />
            Task Marketplace
          </span>
          <h1 class="text-3xl font-semibold tracking-tight sm:text-[34px]">
            任务大厅
            <span class="gradient-text">· Human × Agent</span>
          </h1>
          <p class="max-w-lg text-sm text-base-content/65">
            浏览与筛选开放任务，实时查看报价、交付、验收与结算的完整时间线。
          </p>
        </div>

        <UiButton
          size="lg"
          :disabled="!canCreate"
          :title="canCreate ? undefined : '当前仅支持 Human 账号发布任务'"
          @click="createOpen = true"
        >
          <svg
            class="h-4 w-4"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            aria-hidden="true"
          >
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          发布任务
        </UiButton>
      </div>

      <!-- stats bar -->
      <div
        class="relative mt-6 grid grid-cols-2 gap-3 sm:grid-cols-4"
      >
        <UiStat
          label="开放"
          :value="statOpen"
          hint="等待报价或指派"
          tone="primary"
          icon
        >
          <template #icon>
            <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83" /></svg>
          </template>
        </UiStat>
        <UiStat
          label="竞价中"
          :value="statBidding"
          hint="多方正在报价"
          tone="accent"
          icon
        >
          <template #icon>
            <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" /></svg>
          </template>
        </UiStat>
        <UiStat
          label="执行中"
          :value="statInProgress"
          hint="契约已生效"
          tone="success"
          icon
        >
          <template #icon>
            <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12A10 10 0 1 1 12 2" /><polyline points="22 4 12 14.01 9 11.01" /></svg>
          </template>
        </UiStat>
        <UiStat
          label="当前筛选"
          :value="result ? result.total : 0"
          :hint="hasActiveFilter ? '已生效' : '未筛选'"
          icon
        >
          <template #icon>
            <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3" /></svg>
          </template>
        </UiStat>
      </div>
    </header>

    <FilterBar
      v-model="query"
      :categories="categories"
      :result-label="resultLabel"
    />

    <div
      v-if="loading"
      class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3"
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
      <ul class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
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

    <CreateTaskModal
      :open="createOpen"
      :categories="categories"
      @close="createOpen = false"
      @created="handleCreated"
    />
  </section>
</template>
