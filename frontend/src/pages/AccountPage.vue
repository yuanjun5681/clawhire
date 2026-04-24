<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { ApiRequestError, accountsApi, executorsApi, tasksApi } from '@/api'
import { useIdentityStore } from '@/stores/identity'
import TaskCard from '@/components/TaskCard.vue'
import EmptyState from '@/components/EmptyState.vue'
import ErrorState from '@/components/ErrorState.vue'
import { UiAvatar, UiBadge, UiStat } from '@/components/ui'
import { formatDate, formatDateTime } from '@/utils/format'
import type {
  AccountDetail,
  AccountListItem,
  AccountStats,
  TaskListItem,
} from '@/types'

const route = useRoute()
const identity = useIdentityStore()

const accountId = computed(() => String(route.params.accountId ?? ''))

const loading = ref(true)
const apiError = ref<{ message: string; code?: string } | null>(null)
const account = ref<AccountDetail | null>(null)
const stats = ref<AccountStats | null>(null)
const recentTasks = ref<TaskListItem[]>([])
const ownedAgents = ref<AccountListItem[]>([])

async function load() {
  if (!accountId.value) return
  loading.value = true
  apiError.value = null
  try {
    const [a, postedHead, postedSettledHead, executorHead, executorSettledHead] = await Promise.all([
      accountsApi.getAccount(accountId.value),
      tasksApi.listTasks({ requesterId: accountId.value, page: 1, pageSize: 10 }),
      tasksApi.listTasks({
        requesterId: accountId.value,
        status: 'SETTLED',
        page: 1,
        pageSize: 1,
      }),
      executorsApi.listExecutorHistory(accountId.value, { page: 1, pageSize: 10 }),
      executorsApi.listExecutorHistory(accountId.value, {
        status: 'SETTLED',
        page: 1,
        pageSize: 1,
      }),
    ])
    account.value = a
    stats.value = {
      postedCount: postedHead.total,
      executedCount: executorHead.total,
      settledCount: postedSettledHead.total + executorSettledHead.total,
    }
    recentTasks.value = [...postedHead.items, ...executorHead.items]
      .filter(
        (item, index, list) =>
          list.findIndex((candidate) => candidate.taskId === item.taskId) === index,
      )
      .slice(0, 10)
    ownedAgents.value =
      a.type === 'human'
        ? await accountsApi.listAccountAgents(accountId.value)
        : []
  } catch (e: unknown) {
    if (e instanceof ApiRequestError) {
      apiError.value = { message: e.message, code: e.code }
    } else {
      apiError.value = {
        message: e instanceof Error ? e.message : String(e),
      }
    }
    account.value = null
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(accountId, load)

const isSelf = computed(
  () => account.value?.accountId === identity.currentAccountId,
)

const isAgent = computed(() => account.value?.type === 'agent')

const STATUS_TONE: Record<string, 'success' | 'neutral' | 'warning'> = {
  active: 'success',
  disabled: 'neutral',
  pending: 'warning',
}
const STATUS_LABEL: Record<string, string> = {
  active: '活跃',
  disabled: '已禁用',
  pending: '待审核',
}

const TYPE_LABEL = { human: '人类', agent: 'Agent' } as const
</script>

<template>
  <section class="space-y-6">
    <nav class="flex items-center gap-1 text-xs text-base-content/50">
      <RouterLink to="/tasks" class="hover:text-primary">任务大厅</RouterLink>
      <svg class="h-3 w-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6" /></svg>
      <span class="text-base-content/70">账号主页</span>
    </nav>

    <div v-if="loading" class="space-y-4">
      <div class="h-40 animate-pulse rounded-box bg-base-200" />
      <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
        <div class="h-24 animate-pulse rounded-box bg-base-200" />
        <div class="h-24 animate-pulse rounded-box bg-base-200" />
        <div class="h-24 animate-pulse rounded-box bg-base-200" />
      </div>
      <div class="h-48 animate-pulse rounded-box bg-base-200" />
    </div>

    <ErrorState
      v-else-if="apiError"
      :message="apiError.message"
      :code="apiError.code"
      @retry="load"
    />

    <template v-else-if="account">
      <!-- Identity header -->
      <header
        class="relative overflow-hidden rounded-box border border-base-300/70"
      >
        <div
          aria-hidden="true"
          class="absolute inset-0 bg-[linear-gradient(120deg,color-mix(in_oklch,var(--color-primary)_22%,var(--color-base-100))_0%,color-mix(in_oklch,var(--color-base-100)_100%,transparent)_55%,color-mix(in_oklch,var(--color-accent)_18%,var(--color-base-100))_100%)]"
        />
        <span
          aria-hidden="true"
          class="pointer-events-none absolute -right-24 -top-24 h-64 w-64 rounded-full bg-primary/20 blur-3xl"
        />
        <span
          aria-hidden="true"
          class="pointer-events-none absolute -left-16 -bottom-24 h-60 w-60 rounded-full bg-accent/20 blur-3xl"
        />

        <div class="relative flex flex-col items-start gap-4 p-6 sm:flex-row sm:items-center sm:gap-6 sm:p-8">
          <UiAvatar
            :name="account.displayName"
            size="xl"
            :tone="isAgent ? 'primary' : 'brand'"
            ring
          />
          <div class="min-w-0 flex-1 space-y-2">
            <div class="flex flex-wrap items-center gap-2">
              <h1 class="text-2xl font-semibold tracking-tight sm:text-[28px]">
                {{ account.displayName }}
              </h1>
              <UiBadge :tone="isAgent ? 'primary' : 'neutral'" size="sm">
                {{ TYPE_LABEL[account.type] }}
              </UiBadge>
              <UiBadge :tone="STATUS_TONE[account.status] ?? 'neutral'" size="sm" dot>
                {{ STATUS_LABEL[account.status] }}
              </UiBadge>
              <UiBadge v-if="isSelf" tone="accent" size="sm">
                当前账号
              </UiBadge>
            </div>
            <p class="font-mono text-xs text-base-content/55">
              {{ account.accountId }}
              <span v-if="isAgent && account.nodeId">
                · node · {{ account.nodeId }}
              </span>
            </p>
            <p
              v-if="account.profile?.bio"
              class="max-w-xl text-sm text-base-content/75"
            >
              {{ account.profile.bio }}
            </p>
            <p class="text-xs text-base-content/50">
              注册于 {{ formatDate(account.createdAt) }}
              · 更新 {{ formatDateTime(account.updatedAt) }}
            </p>
          </div>
        </div>
      </header>

      <!-- Stats -->
      <div
        v-if="stats"
        class="grid grid-cols-1 gap-4 md:grid-cols-3"
      >
        <UiStat label="发布任务" :value="stats.postedCount" tone="primary" icon>
          <template #icon>
            <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" /><polyline points="14 2 14 8 20 8" /></svg>
          </template>
        </UiStat>
        <UiStat label="执行任务" :value="stats.executedCount" tone="accent" icon>
          <template #icon>
            <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10" /><polyline points="12 6 12 12 16 14" /></svg>
          </template>
        </UiStat>
        <UiStat label="已结算" :value="stats.settledCount" tone="success" icon>
          <template #icon>
            <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="6" width="20" height="13" rx="2" /><path d="M2 10h20" /></svg>
          </template>
        </UiStat>
      </div>

      <!-- Owned agents -->
      <section
        v-if="!isAgent && ownedAgents.length > 0"
        class="rounded-box border border-base-300/70 bg-base-100 p-5"
      >
        <header class="mb-4 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <svg
              class="h-4 w-4 text-primary"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <rect x="3" y="11" width="18" height="10" rx="2" />
              <circle cx="12" cy="5" r="2" />
              <path d="M12 7v4M8 16h.01M16 16h.01" />
            </svg>
            <h2 class="text-sm font-semibold tracking-tight">旗下 Agent</h2>
          </div>
          <span class="text-xs text-base-content/55">共 {{ ownedAgents.length }} 个</span>
        </header>
        <ul class="grid grid-cols-1 gap-3 md:grid-cols-2">
          <li v-for="a in ownedAgents" :key="a.accountId">
            <RouterLink
              :to="`/accounts/${a.accountId}`"
              class="group flex items-center gap-3 rounded-box border border-base-300/70 bg-base-100 p-3 transition hover:-translate-y-0.5 hover:border-primary/40 hover:shadow-[0_12px_30px_-14px_color-mix(in_oklch,var(--color-primary)_30%,transparent)]"
            >
              <UiAvatar :name="a.displayName" size="md" tone="primary" />
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm font-medium group-hover:text-primary">
                  {{ a.displayName }}
                </p>
                <p
                  v-if="a.nodeId"
                  class="truncate font-mono text-[11px] text-base-content/55"
                >
                  node · {{ a.nodeId }}
                </p>
              </div>
              <svg
                class="h-4 w-4 text-base-content/40 transition-transform group-hover:translate-x-0.5 group-hover:text-primary"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <polyline points="9 18 15 12 9 6" />
              </svg>
            </RouterLink>
          </li>
        </ul>
      </section>

      <!-- Recent tasks -->
      <section class="rounded-box border border-base-300/70 bg-base-100 p-5">
        <header class="mb-4 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <svg
              class="h-4 w-4 text-primary"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <rect x="3" y="4" width="18" height="18" rx="2" />
              <path d="M16 2v4M8 2v4M3 10h18" />
            </svg>
            <h2 class="text-sm font-semibold tracking-tight">近期任务</h2>
          </div>
          <span class="text-xs text-base-content/55">最多 10 条</span>
        </header>

        <EmptyState
          v-if="recentTasks.length === 0"
          title="暂无任务记录"
          description="该账号还没有参与任何任务。"
        />

        <ul v-else class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <li v-for="t in recentTasks" :key="t.taskId">
            <TaskCard :task="t" />
          </li>
        </ul>
      </section>
    </template>
  </section>
</template>
