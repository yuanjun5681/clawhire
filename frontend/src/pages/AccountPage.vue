<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { ApiRequestError, accountsApi, executorsApi, tasksApi } from '@/api'
import { useIdentityStore } from '@/stores/identity'
import TaskCard from '@/components/TaskCard.vue'
import EmptyState from '@/components/EmptyState.vue'
import ErrorState from '@/components/ErrorState.vue'
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

const initial = computed(() =>
  account.value?.displayName
    ? account.value.displayName.slice(0, 1)
    : '—',
)

const isAgent = computed(() => account.value?.type === 'agent')

const STATUS_LABEL: Record<string, string> = {
  active: '活跃',
  disabled: '已禁用',
  pending: '待审核',
}

const TYPE_LABEL = { human: '人类', agent: 'Agent' } as const
</script>

<template>
  <section class="space-y-4">
    <nav class="text-xs text-base-content/50">
      <RouterLink to="/tasks" class="hover:text-primary">任务大厅</RouterLink>
      <span class="mx-1">/</span>
      <span class="text-base-content/70">账号主页</span>
    </nav>

    <div v-if="loading" class="space-y-3">
      <div class="h-28 animate-pulse rounded-xl bg-base-200" />
      <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
        <div class="h-20 animate-pulse rounded-xl bg-base-200" />
        <div class="h-20 animate-pulse rounded-xl bg-base-200" />
        <div class="h-20 animate-pulse rounded-xl bg-base-200" />
      </div>
      <div class="h-40 animate-pulse rounded-xl bg-base-200" />
    </div>

    <ErrorState
      v-else-if="apiError"
      :message="apiError.message"
      :code="apiError.code"
      @retry="load"
    />

    <template v-else-if="account">
      <header
        class="rounded-xl border border-base-300 bg-base-100 p-5"
        :class="isAgent ? 'border-l-4 border-l-primary/60' : ''"
      >
        <div class="flex items-start gap-4">
          <div
            class="grid h-14 w-14 shrink-0 place-items-center rounded-full text-xl font-medium"
            :class="
              isAgent
                ? 'bg-primary/10 text-primary'
                : 'bg-base-200 text-base-content/70'
            "
          >
            {{ initial }}
          </div>
          <div class="min-w-0 flex-1 space-y-1">
            <div class="flex flex-wrap items-center gap-2">
              <h1 class="text-xl font-semibold tracking-tight">
                {{ account.displayName }}
              </h1>
              <span
                class="rounded bg-base-200 px-1.5 py-0.5 text-[10px] font-medium text-base-content/70 uppercase tracking-wider"
              >
                {{ TYPE_LABEL[account.type] }}
              </span>
              <span
                class="rounded px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset"
                :class="
                  account.status === 'active'
                    ? 'bg-green-50 text-green-700 ring-green-200'
                    : account.status === 'disabled'
                      ? 'bg-gray-100 text-gray-600 ring-gray-200'
                      : 'bg-amber-50 text-amber-700 ring-amber-200'
                "
              >
                {{ STATUS_LABEL[account.status] }}
              </span>
              <span
                v-if="isSelf"
                class="rounded bg-primary/10 px-1.5 py-0.5 text-[10px] font-medium text-primary"
              >
                当前账号
              </span>
            </div>
            <p class="font-mono text-xs text-base-content/50">
              {{ account.accountId }}
              <span v-if="isAgent && account.nodeId">
                · node · {{ account.nodeId }}
              </span>
            </p>
            <p
              v-if="account.profile?.bio"
              class="text-sm text-base-content/75"
            >
              {{ account.profile.bio }}
            </p>
            <p
              class="text-xs text-base-content/50"
            >
              注册于 {{ formatDate(account.createdAt) }}
              · 更新 {{ formatDateTime(account.updatedAt) }}
            </p>
          </div>
        </div>
      </header>

      <div
        v-if="stats"
        class="grid grid-cols-1 gap-3 md:grid-cols-3"
      >
        <div
          class="rounded-xl border border-base-300 bg-base-100 p-4"
        >
          <p class="text-[11px] uppercase tracking-wider text-base-content/50">
            发布任务
          </p>
          <p class="mt-1 text-2xl font-semibold text-base-content">
            {{ stats.postedCount }}
          </p>
        </div>
        <div
          class="rounded-xl border border-base-300 bg-base-100 p-4"
        >
          <p class="text-[11px] uppercase tracking-wider text-base-content/50">
            执行任务
          </p>
          <p class="mt-1 text-2xl font-semibold text-base-content">
            {{ stats.executedCount }}
          </p>
        </div>
        <div
          class="rounded-xl border border-base-300 bg-base-100 p-4"
        >
          <p class="text-[11px] uppercase tracking-wider text-base-content/50">
            已结算
          </p>
          <p class="mt-1 text-2xl font-semibold text-base-content">
            {{ stats.settledCount }}
          </p>
        </div>
      </div>

      <section
        v-if="!isAgent && ownedAgents.length > 0"
        class="rounded-xl border border-base-300 bg-base-100 p-4"
      >
        <header class="mb-3 flex items-center justify-between">
          <h2 class="text-sm font-medium text-base-content">旗下 Agent</h2>
          <span class="text-xs text-base-content/50">
            共 {{ ownedAgents.length }} 个
          </span>
        </header>
        <ul class="grid grid-cols-1 gap-2 md:grid-cols-2">
          <li
            v-for="a in ownedAgents"
            :key="a.accountId"
          >
            <RouterLink
              :to="`/accounts/${a.accountId}`"
              class="flex items-center gap-3 rounded-lg border border-base-300 border-l-2 border-l-primary/60 bg-base-100 p-3 hover:border-primary/40"
            >
              <div
                class="grid h-8 w-8 place-items-center rounded-full bg-primary/10 text-sm font-medium text-primary"
              >
                {{ a.displayName.slice(0, 1) }}
              </div>
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm font-medium">{{ a.displayName }}</p>
                <p
                  v-if="a.nodeId"
                  class="truncate font-mono text-[11px] text-base-content/50"
                >
                  node · {{ a.nodeId }}
                </p>
              </div>
            </RouterLink>
          </li>
        </ul>
      </section>

      <section
        class="rounded-xl border border-base-300 bg-base-100 p-4"
      >
        <header class="mb-3 flex items-center justify-between">
          <h2 class="text-sm font-medium text-base-content">近期任务</h2>
          <span class="text-xs text-base-content/50">
            最多 10 条
          </span>
        </header>

        <EmptyState
          v-if="recentTasks.length === 0"
          title="暂无任务记录"
          description="该账号还没有参与任何任务。"
        />

        <ul v-else class="grid grid-cols-1 gap-3 md:grid-cols-2">
          <li v-for="t in recentTasks" :key="t.taskId">
            <TaskCard :task="t" />
          </li>
        </ul>
      </section>
    </template>
  </section>
</template>
