<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError, tasksApi } from '@/api'
import { useIdentityStore } from '@/stores/identity'
import TaskCard from '@/components/TaskCard.vue'
import SkeletonTaskCard from '@/components/SkeletonTaskCard.vue'
import EmptyState from '@/components/EmptyState.vue'
import ErrorState from '@/components/ErrorState.vue'
import Pagination from '@/components/Pagination.vue'
import CreateTaskModal from '@/components/CreateTaskModal.vue'
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
const router = useRouter()

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
const createOpen = ref(false)
const toast = ref<string | null>(null)

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

function flash(message: string) {
  toast.value = message
  window.setTimeout(() => {
    toast.value = null
  }, 2400)
}

function handleCreated(taskId: string) {
  createOpen.value = false
  flash('任务已发布')
  if (activeTab.value === 'published') {
    load()
  } else {
    router.push(`/tasks/${taskId}`)
  }
}

onMounted(load)
watch([activeTab, page], load)
</script>

<template>
  <section class="space-y-4">
    <header class="flex flex-wrap items-start justify-between gap-3">
      <div class="space-y-1">
        <h1 class="text-2xl font-semibold tracking-tight">我的任务</h1>
        <p class="text-sm text-base-content/60">
          以 {{ identity.displayName }} 的身份查看与我相关的任务。
        </p>
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

    <CreateTaskModal
      :open="createOpen"
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
