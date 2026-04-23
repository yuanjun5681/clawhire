<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ApiRequestError, tasksApi } from '@/api'
import { useIdentityStore } from '@/stores/identity'
import TaskCard from '@/components/TaskCard.vue'
import SkeletonTaskCard from '@/components/SkeletonTaskCard.vue'
import EmptyState from '@/components/EmptyState.vue'
import ErrorState from '@/components/ErrorState.vue'
import Pagination from '@/components/Pagination.vue'
import type { Paginated, TaskListItem, TaskQuery } from '@/types'

const PAGE_SIZE = 6

type TabKey = 'published' | 'executing' | 'pending_review'

interface TabConfig {
  key: TabKey
  label: string
  emptyTitle: string
  emptyDescription: string
  buildQuery: (accountId: string) => TaskQuery
}

const identity = useIdentityStore()

const TABS: TabConfig[] = [
  {
    key: 'published',
    label: '我发布的',
    emptyTitle: '你还没有发布过任务',
    emptyDescription: '在任务大厅新建一个任务，开始协作。',
    buildQuery: (id) => ({ requesterId: id }),
  },
  {
    key: 'executing',
    label: '我执行的',
    emptyTitle: '当前没有你执行中的任务',
    emptyDescription: '在任务大厅寻找合适的任务，参与报价。',
    buildQuery: (id) => ({ executorId: id }),
  },
  {
    key: 'pending_review',
    label: '待验收',
    emptyTitle: '暂无等待你验收的任务',
    emptyDescription: '执行方提交交付后，会出现在这里等待验收。',
    buildQuery: (id) => ({ reviewerId: id, status: 'SUBMITTED' }),
  },
]

const activeTab = ref<TabKey>('published')
const page = ref(1)
const loading = ref(true)
const apiError = ref<{ message: string; code?: string } | null>(null)
const result = ref<Paginated<TaskListItem> | null>(null)

const currentTab = computed(
  () => TABS.find((t) => t.key === activeTab.value) ?? TABS[0],
)

const currentQuery = computed<TaskQuery>(() => ({
  ...currentTab.value.buildQuery(identity.currentAccountId),
  page: page.value,
  pageSize: PAGE_SIZE,
}))

async function load() {
  loading.value = true
  apiError.value = null
  try {
    result.value = await tasksApi.listTasks(currentQuery.value)
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

function selectTab(key: TabKey) {
  if (activeTab.value === key) return
  activeTab.value = key
  page.value = 1
}

function changePage(p: number) {
  page.value = p
}

onMounted(load)
watch([activeTab, page], load)
</script>

<template>
  <section class="space-y-4">
    <header class="space-y-1">
      <h1 class="text-2xl font-semibold tracking-tight">我的任务</h1>
      <p class="text-sm text-base-content/60">
        以 {{ identity.displayName }} 的身份查看与我相关的任务。
      </p>
    </header>

    <div
      role="tablist"
      class="flex items-center gap-1 border-b border-base-300"
    >
      <button
        v-for="t in TABS"
        :key="t.key"
        role="tab"
        type="button"
        :aria-selected="activeTab === t.key"
        class="-mb-px border-b-2 px-3 py-2 text-sm transition"
        :class="
          activeTab === t.key
            ? 'border-primary text-primary font-medium'
            : 'border-transparent text-base-content/60 hover:text-base-content'
        "
        @click="selectTab(t.key)"
      >
        {{ t.label }}
        <span
          v-if="activeTab === t.key && result"
          class="ml-1 text-xs text-base-content/50"
        >
          · {{ result.total }}
        </span>
      </button>
    </div>

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
      v-else-if="result && result.items.length === 0"
      :title="currentTab.emptyTitle"
      :description="currentTab.emptyDescription"
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
