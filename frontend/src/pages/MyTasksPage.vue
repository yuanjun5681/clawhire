<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError, tasksApi } from '@/api'
import { useIdentityStore } from '@/stores/identity'
import { useToastStore } from '@/stores/toast'
import TaskCard from '@/components/TaskCard.vue'
import SkeletonTaskCard from '@/components/SkeletonTaskCard.vue'
import EmptyState from '@/components/EmptyState.vue'
import ErrorState from '@/components/ErrorState.vue'
import Pagination from '@/components/Pagination.vue'
import CreateTaskModal from '@/components/CreateTaskModal.vue'
import { UiButton } from '@/components/ui'
import type { Paginated, TaskListItem, TaskQuery } from '@/types'

const PAGE_SIZE = 6

type TabKey = 'published' | 'executing' | 'pending_review'

interface TabConfig {
  key: TabKey
  label: string
  emptyTitle: string
  emptyDescription: string
  icon: string
  buildQuery: (accountId: string) => TaskQuery
}

const identity = useIdentityStore()
const router = useRouter()
const toast = useToastStore()

const TABS: TabConfig[] = [
  {
    key: 'published',
    label: '我发布的',
    emptyTitle: '你还没有发布过任务',
    emptyDescription: '在任务大厅新建一个任务，开始协作。',
    icon: 'M12 20h9 M16.5 3.5a2.12 2.12 0 1 1 3 3L7 19l-4 1 1-4L16.5 3.5z',
    buildQuery: (id) => ({ requesterId: id }),
  },
  {
    key: 'executing',
    label: '我执行的',
    emptyTitle: '当前没有你执行中的任务',
    emptyDescription: '在任务大厅寻找合适的任务，参与报价。',
    icon: 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z',
    buildQuery: (id) => ({ executorId: id }),
  },
  {
    key: 'pending_review',
    label: '待验收',
    emptyTitle: '暂无等待你验收的任务',
    emptyDescription: '执行方提交交付后，会出现在这里等待验收。',
    icon: 'M22 11.08V12a10 10 0 1 1-5.93-9.14 M22 4 12 14.01l-3-3',
    buildQuery: (id) => ({ reviewerId: id, status: 'SUBMITTED' }),
  },
]

const activeTab = ref<TabKey>('published')
const page = ref(1)
const loading = ref(true)
const apiError = ref<{ message: string; code?: string } | null>(null)
const result = ref<Paginated<TaskListItem> | null>(null)
const createOpen = ref(false)

// 每个 tab 的总数缓存，用来在 tab 上显示数字
const counts = ref<Record<TabKey, number | null>>({
  published: null,
  executing: null,
  pending_review: null,
})

const canCreate = computed(() => identity.accountType === 'human')

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
    counts.value[activeTab.value] = result.value.total
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

async function loadCounts() {
  try {
    const [a, b, c] = await Promise.all(
      TABS.map((t) =>
        tasksApi.listTasks({
          ...t.buildQuery(identity.currentAccountId),
          page: 1,
          pageSize: 1,
        }),
      ),
    )
    counts.value.published = a.total
    counts.value.executing = b.total
    counts.value.pending_review = c.total
  } catch {
    // 忽略
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

function handleCreated(taskId: string) {
  createOpen.value = false
  toast.success('任务已发布')
  if (activeTab.value === 'published') {
    load()
  } else {
    router.push(`/tasks/${taskId}`)
  }
}

onMounted(() => {
  load()
  loadCounts()
})
watch([activeTab, page], load)
</script>

<template>
  <section class="space-y-6">
    <header
      class="relative overflow-hidden rounded-box border border-base-300/70 bg-[linear-gradient(130deg,color-mix(in_oklch,var(--color-secondary)_10%,var(--color-base-100))_0%,var(--color-base-100)_55%,color-mix(in_oklch,var(--color-primary)_10%,var(--color-base-100))_100%)] px-6 py-7 sm:px-8"
    >
      <span
        aria-hidden="true"
        class="pointer-events-none absolute -right-20 -top-24 h-56 w-56 rounded-full bg-primary/20 blur-3xl"
      />
      <div class="relative flex flex-wrap items-start justify-between gap-4">
        <div class="space-y-2">
          <span
            class="inline-flex items-center gap-1.5 rounded-full bg-secondary/15 px-2.5 py-1 text-[11px] font-medium text-secondary ring-1 ring-secondary/25"
          >
            <span class="h-1.5 w-1.5 rounded-full bg-secondary animate-pulse" />
            我的工作台
          </span>
          <h1 class="text-3xl font-semibold tracking-tight sm:text-[32px]">
            我的任务
            <span class="gradient-text">· {{ identity.displayName }}</span>
          </h1>
          <p class="max-w-lg text-sm text-base-content/65">
            以 <span class="font-mono text-base-content/80">{{ identity.currentAccountId }}</span> 的身份查看发布、执行和待验收的任务。
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
    </header>

    <!-- Tabs: segmented pills -->
    <div
      role="tablist"
      class="inline-flex flex-wrap items-center gap-1.5 rounded-full border border-base-300/70 bg-base-100/80 p-1 backdrop-blur"
    >
      <button
        v-for="t in TABS"
        :key="t.key"
        role="tab"
        type="button"
        :aria-selected="activeTab === t.key"
        :class="[
          'group relative inline-flex items-center gap-2 rounded-full px-4 py-2 text-sm transition',
          activeTab === t.key
            ? 'bg-[linear-gradient(120deg,var(--color-primary),var(--color-accent))] text-primary-content shadow-[0_6px_20px_-8px_color-mix(in_oklch,var(--color-primary)_60%,transparent)]'
            : 'text-base-content/60 hover:text-base-content',
        ]"
        @click="selectTab(t.key)"
      >
        <svg
          class="h-3.5 w-3.5"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="1.8"
          stroke-linecap="round"
          stroke-linejoin="round"
          aria-hidden="true"
        >
          <path :d="t.icon" />
        </svg>
        {{ t.label }}
        <span
          v-if="counts[t.key] !== null"
          :class="[
            'rounded-full px-1.5 py-0.5 text-[10.5px] font-semibold',
            activeTab === t.key
              ? 'bg-white/25 text-primary-content'
              : 'bg-base-200 text-base-content/70',
          ]"
        >
          {{ counts[t.key] }}
        </span>
      </button>
    </div>

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
      v-else-if="result && result.items.length === 0"
      :title="currentTab.emptyTitle"
      :description="currentTab.emptyDescription"
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
      @close="createOpen = false"
      @created="handleCreated"
    />
  </section>
</template>
