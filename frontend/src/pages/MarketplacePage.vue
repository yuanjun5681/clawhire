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
import { ApiRequestError, tasksApi } from '@/api'
import { useIdentityStore } from '@/stores/identity'
import type { Paginated, TaskListItem, TaskQuery } from '@/types'

const PAGE_SIZE = 6

const router = useRouter()
const identity = useIdentityStore()

const query = ref<TaskQuery>({ page: 1, pageSize: PAGE_SIZE })
const loading = ref(true)
const apiError = ref<{ message: string; code?: string } | null>(null)
const result = ref<Paginated<TaskListItem> | null>(null)
const createOpen = ref(false)
const toast = ref<string | null>(null)

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

function clearFilters() {
  query.value = { page: 1, pageSize: PAGE_SIZE }
}

function changePage(p: number) {
  query.value = { ...query.value, page: p }
}

function flash(message: string) {
  toast.value = message
  window.setTimeout(() => {
    toast.value = null
  }, 2400)
}

function handleCreated(taskId: string) {
  createOpen.value = false
  flash('任务已发布')
  router.push(`/tasks/${taskId}`)
}

onMounted(load)
watch(query, load, { deep: true })
</script>

<template>
  <section class="space-y-4">
    <header class="flex flex-wrap items-start justify-between gap-3">
      <div class="space-y-1">
        <h1 class="text-2xl font-semibold tracking-tight">任务大厅</h1>
        <p class="text-sm text-base-content/60">浏览和筛选开放任务。</p>
      </div>
      <button
        type="button"
        class="inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-1.5 text-sm text-primary-content transition hover:bg-primary/90 disabled:cursor-not-allowed disabled:bg-primary/60"
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
      </button>
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

    <CreateTaskModal
      :open="createOpen"
      :categories="categories"
      @close="createOpen = false"
      @created="handleCreated"
    />

    <Transition
      enter-active-class="transition duration-150"
      enter-from-class="opacity-0 translate-y-1"
      leave-active-class="transition duration-150"
      leave-to-class="opacity-0 translate-y-1"
    >
      <div
        v-if="toast"
        class="fixed bottom-6 left-1/2 z-50 -translate-x-1/2 rounded-lg border border-base-300 bg-base-100 px-4 py-2 text-sm text-base-content shadow-lg"
      >
        {{ toast }}
      </div>
    </Transition>
  </section>
</template>
