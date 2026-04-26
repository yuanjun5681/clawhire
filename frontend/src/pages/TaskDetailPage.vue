<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import {
  ApiRequestError,
  taskResourcesApi,
  tasksApi,
} from '@/api'
import { useIdentityStore } from '@/stores/identity'
import { useToastStore } from '@/stores/toast'
import StatusBadge from '@/components/StatusBadge.vue'
import RoleCard from '@/components/RoleCard.vue'
import ActionBar from '@/components/ActionBar.vue'
import type { ActionItem } from '@/components/ActionBar.vue'
import Timeline from '@/components/Timeline.vue'
import type { TimelineEvent } from '@/components/Timeline.vue'
import ErrorState from '@/components/ErrorState.vue'
import AwardModal from '@/components/AwardModal.vue'
import { UiBadge } from '@/components/ui'
import { formatDate, formatDateTime, formatReward } from '@/utils/format'
import type {
  Bid,
  Progress,
  Review,
  Settlement,
  Submission,
  TaskDetail,
  TaskStatus,
} from '@/types'

const route = useRoute()
const identity = useIdentityStore()
const toast = useToastStore()

const taskId = computed(() => String(route.params.taskId ?? ''))

const loading = ref(true)
const apiError = ref<{ message: string; code?: string } | null>(null)
const task = ref<TaskDetail | null>(null)
const bids = ref<Bid[]>([])
const progress = ref<Progress[]>([])
const submissions = ref<Submission[]>([])
const reviews = ref<Review[]>([])
const settlements = ref<Settlement[]>([])
const actionSubmitting = ref(false)
const showAwardModal = ref(false)

async function load() {
  if (!taskId.value) return
  loading.value = true
  apiError.value = null
  try {
    const [t, b, p, s, r, st] = await Promise.all([
      tasksApi.getTask(taskId.value),
      taskResourcesApi.listBids(taskId.value),
      taskResourcesApi.listProgress(taskId.value),
      taskResourcesApi.listSubmissions(taskId.value),
      taskResourcesApi.listReviews(taskId.value),
      taskResourcesApi.listSettlements(taskId.value),
    ])
    task.value = t
    bids.value = b
    progress.value = p
    submissions.value = s
    reviews.value = r
    settlements.value = st
  } catch (e: unknown) {
    if (e instanceof ApiRequestError) {
      apiError.value = { message: e.message, code: e.code }
    } else {
      apiError.value = {
        message: e instanceof Error ? e.message : String(e),
      }
    }
    task.value = null
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(taskId, load)

type ViewerRole = 'requester' | 'executor' | 'reviewer' | 'visitor'

const viewerRole = computed<ViewerRole>(() => {
  const t = task.value
  if (!t) return 'visitor'
  const me = identity.currentAccountId
  if (t.requester.id === me) return 'requester'
  if (t.assignedExecutor?.id === me) return 'executor'
  if (t.reviewer?.id === me) return 'reviewer'
  return 'visitor'
})

const STAGE_HINT: Record<TaskStatus, string> = {
  OPEN: '任务已发布，等待潜在执行者报价。',
  BIDDING: '正在征集报价，可继续报价或由需求方指派。',
  AWARDED: '任务已指派，等待执行方开始执行。',
  IN_PROGRESS: '任务执行中，可持续汇报进度直至提交交付。',
  SUBMITTED: '交付物已提交，等待验收方处理。',
  ACCEPTED: '已通过验收，等待发起结算。',
  SETTLED: '结算完成，任务关闭。',
  REJECTED: '验收未通过，执行方可根据反馈调整后重新提交。',
  CANCELLED: '任务已取消。',
  EXPIRED: '任务已过期。',
  DISPUTED: '任务存在争议，等待双方提交证据或调解。',
}

const actions = computed<ActionItem[]>(() => {
  const t = task.value
  if (!t) return []
  const role = viewerRole.value
  const list: ActionItem[] = []

  const canBid = role === 'visitor' && identity.accountType === 'human'
  const bidDisabledReason =
    role === 'visitor' && identity.accountType !== 'human'
      ? '当前前端仅支持 Human HTTP 写接口'
      : undefined

  switch (t.status) {
    case 'OPEN':
    case 'BIDDING':
      if (role === 'visitor') {
        list.push({
          key: 'bid',
          label: '提交报价',
          primary: canBid,
          disabledReason: bidDisabledReason,
        })
      }
      if (role === 'requester') {
        list.push({ key: 'assign', label: '指派执行方', primary: true })
        list.push({ key: 'cancel', label: '取消任务', danger: true })
      }
      break
    case 'AWARDED':
      break
    case 'IN_PROGRESS':
      if (role === 'executor') {
        list.push({ key: 'submit', label: '提交交付', primary: true })
      }
      break
    case 'SUBMITTED':
      if (role === 'reviewer' || role === 'requester') {
        list.push({ key: 'approve', label: '通过验收', primary: true })
        list.push({ key: 'reject', label: '驳回', danger: true })
      }
      break
    case 'ACCEPTED':
      if (role === 'reviewer' || role === 'requester') {
        list.push({ key: 'settle', label: '发起结算', primary: true })
      }
      break
    case 'SETTLED':
    case 'CANCELLED':
    case 'EXPIRED':
    case 'REJECTED':
    case 'DISPUTED':
      break
  }

  return list
})

const emptyActionHint = computed(() => {
  const t = task.value
  if (!t) return ''
  const role = viewerRole.value
  if (t.status === 'SETTLED') return '任务已结算关闭，无可执行操作。'
  if (t.status === 'CANCELLED') return '任务已取消。'
  if (t.status === 'EXPIRED') return '任务已过期。'
  if (role === 'visitor') return '你当前不是该任务的参与方。'
  return '当前阶段下你没有可执行的操作。'
})

function nextId(prefix: string) {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return `${prefix}_${crypto.randomUUID().slice(0, 8)}`
  }
  return `${prefix}_${Date.now()}`
}

async function runAction(key: string) {
  if (!task.value || actionSubmitting.value) return

  const currentTask = task.value
  const latestSubmission = submissions.value[0]

  try {
    actionSubmitting.value = true
    switch (key) {
      case 'bid': {
        const priceRaw = window.prompt('报价金额（USD）')
        if (!priceRaw) return
        const price = Number(priceRaw)
        if (!Number.isFinite(price) || price <= 0) throw new Error('报价金额无效')
        const proposal = window.prompt('报价说明（可选）')?.trim()
        await tasksApi.createBid(currentTask.taskId, {
          bidId: nextId('bid'),
          price,
          currency: currentTask.reward.currency || 'USD',
          proposal: proposal || undefined,
        })
        toast.success('报价已提交')
        break
      }
      case 'assign': {
        showAwardModal.value = true
        return
      }
      case 'submit': {
        const summary = window.prompt('交付摘要')
        if (!summary?.trim()) return
        const artifactUrl = window.prompt('交付链接（可选）')?.trim()
        const evidenceUrl = window.prompt('证据链接（可选）')?.trim()
        await tasksApi.createSubmission(currentTask.taskId, {
          submissionId: nextId('submission'),
          summary: summary.trim(),
          artifacts: artifactUrl
            ? [{ type: 'url', value: artifactUrl, label: 'Preview' }]
            : [],
          evidence:
            evidenceUrl
              ? {
                  type: 'url',
                  items: [evidenceUrl],
                }
              : undefined,
        })
        toast.success('交付已提交')
        break
      }
      case 'approve': {
        if (!latestSubmission) throw new Error('当前没有可验收的提交记录')
        if (!window.confirm('确认通过当前交付？')) return
        await tasksApi.acceptSubmission(currentTask.taskId, {
          submissionId: latestSubmission.submissionId,
          acceptedAt: new Date().toISOString(),
        })
        toast.success('已通过验收')
        break
      }
      case 'reject': {
        if (!latestSubmission) throw new Error('当前没有可驳回的提交记录')
        const reason = window.prompt('请输入驳回原因')
        if (!reason?.trim()) return
        await tasksApi.rejectSubmission(currentTask.taskId, {
          submissionId: latestSubmission.submissionId,
          reason: reason.trim(),
          rejectedAt: new Date().toISOString(),
        })
        toast.warning('已驳回交付')
        break
      }
      case 'settle': {
        if (!currentTask.assignedExecutor) throw new Error('当前任务没有可结算的执行方')
        if (!window.confirm(`确认向 ${currentTask.assignedExecutor.name} 发起结算？`)) return
        const externalRef = window.prompt('外部流水号（可选）')?.trim()
        await tasksApi.recordSettlement(currentTask.taskId, {
          settlementId: nextId('settlement'),
          channel: 'manual',
          externalRef: externalRef || undefined,
          recordedAt: new Date().toISOString(),
        })
        toast.success('已发起结算')
        break
      }
      default:
        toast.info(`当前前端尚未接入操作：${key}`)
        return
    }

    await load()
  } catch (e: unknown) {
    toast.error(
      e instanceof ApiRequestError
        ? e.message
        : e instanceof Error
          ? e.message
          : '操作失败',
      '操作失败',
    )
  } finally {
    actionSubmitting.value = false
  }
}

async function handleAward(payload: import('@/api/tasks').AwardTaskInput) {
  if (!task.value) return
  showAwardModal.value = false
  try {
    actionSubmitting.value = true
    await tasksApi.awardTask(task.value.taskId, payload)
    toast.success('已完成指派')
    await load()
  } catch (e: unknown) {
    toast.error(
      e instanceof ApiRequestError ? e.message : e instanceof Error ? e.message : '操作失败',
      '操作失败',
    )
  } finally {
    actionSubmitting.value = false
  }
}

const timelineEvents = computed<TimelineEvent[]>(() => {
  const t = task.value
  if (!t) return []
  const evs: TimelineEvent[] = []

  evs.push({
    id: `task-created-${t.taskId}`,
    type: 'status',
    title: '任务创建',
    actor: t.requester,
    at: t.createdAt,
    meta: [
      { label: '分类', value: t.category },
      {
        label: '报酬',
        value: formatReward(t.reward.amount, t.reward.currency),
      },
    ],
  })

  for (const b of bids.value) {
    evs.push({
      id: `bid-${b.bidId}`,
      type: 'bid',
      title:
        b.status === 'accepted'
          ? '报价被采纳'
          : b.status === 'rejected'
            ? '报价被驳回'
            : b.status === 'withdrawn'
              ? '报价已撤回'
              : '提交报价',
      summary: b.proposal,
      actor: b.executor,
      at: b.createdAt,
      meta: [{ label: '金额', value: formatReward(b.price, b.currency) }],
    })
  }

  if (t.assignedExecutor && t.assignedAt) {
    evs.push({
      id: `assign-${t.taskId}`,
      type: 'assign',
      title: `指派执行方 · ${t.assignedExecutor.name}`,
      actor: t.requester,
      at: t.assignedAt,
    })
  }

  for (const p of progress.value) {
    evs.push({
      id: `progress-${p.progressId}`,
      type: 'progress',
      title: p.stage ? `进度汇报 · ${p.stage}` : '进度汇报',
      summary: p.summary,
      actor: p.executor,
      at: p.reportedAt,
      meta:
        typeof p.percent === 'number'
          ? [{ label: '完成度', value: `${p.percent}%` }]
          : undefined,
    })
  }

  for (const s of submissions.value) {
    evs.push({
      id: `submission-${s.submissionId}`,
      type: 'submission',
      title:
        s.status === 'accepted'
          ? '交付已通过'
          : s.status === 'rejected'
            ? '交付已驳回'
            : '提交交付',
      summary: s.summary,
      actor: s.executor,
      at: s.submittedAt,
      meta: s.artifacts?.length
        ? [
            {
              label: '附件',
              value: s.artifacts.map((a) => a.name).join('、'),
            },
          ]
        : undefined,
    })
  }

  for (const r of reviews.value) {
    evs.push({
      id: `review-${r.reviewId}`,
      type:
        r.decision === 'approved' ? 'review_approved' : 'review_rejected',
      title: r.decision === 'approved' ? '验收通过' : '验收驳回',
      summary: r.reason,
      actor: r.reviewer,
      at: r.reviewedAt,
    })
  }

  for (const st of settlements.value) {
    evs.push({
      id: `settlement-${st.settlementId}`,
      type: 'settlement',
      title:
        st.status === 'settled'
          ? '结算完成'
          : st.status === 'failed'
            ? '结算失败'
            : '结算进行中',
      actor: st.payee,
      at: st.recordedAt,
      meta: [
        { label: '金额', value: formatReward(st.amount, st.currency) },
        ...(st.channel ? [{ label: '渠道', value: st.channel }] : []),
        ...(st.externalRef ? [{ label: '外部流水', value: st.externalRef }] : []),
      ],
    })
  }

  return evs
})

const rewardModeLabel = computed(() => {
  switch (task.value?.reward.mode) {
    case 'fixed':
      return '固定价'
    case 'bid':
      return '竞价'
    case 'milestone':
      return '按里程碑'
    default:
      return '—'
  }
})

const acceptanceModeLabel = computed(() => {
  switch (task.value?.acceptanceSpec.mode) {
    case 'manual':
      return '人工验收'
    case 'schema':
      return 'Schema 校验'
    case 'test':
      return '测试驱动'
    case 'hybrid':
      return '混合验收'
    default:
      return '—'
  }
})
</script>

<template>
  <section class="space-y-5">
    <nav class="flex items-center gap-1 text-xs text-base-content/50">
      <RouterLink to="/tasks" class="hover:text-primary">任务大厅</RouterLink>
      <svg class="h-3 w-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6" /></svg>
      <span class="text-base-content/70">任务详情</span>
    </nav>

    <AwardModal
      v-if="task"
      :open="showAwardModal"
      :task="task"
      :bids="bids"
      @close="showAwardModal = false"
      @award="handleAward"
    />

    <div v-if="loading" class="grid grid-cols-1 gap-4 md:grid-cols-12">
      <div class="space-y-4 md:col-span-8">
        <div class="h-28 animate-pulse rounded-box bg-base-200" />
        <div class="h-24 animate-pulse rounded-box bg-base-200" />
        <div class="h-48 animate-pulse rounded-box bg-base-200" />
      </div>
      <div class="space-y-3 md:col-span-4">
        <div class="h-32 animate-pulse rounded-box bg-base-200" />
        <div class="h-24 animate-pulse rounded-box bg-base-200" />
        <div class="h-24 animate-pulse rounded-box bg-base-200" />
      </div>
    </div>

    <ErrorState
      v-else-if="apiError"
      :message="apiError.message"
      :code="apiError.code"
      @retry="load"
    />

    <div
      v-else-if="task"
      class="grid grid-cols-1 gap-5 md:grid-cols-12"
    >
      <div class="space-y-5 md:col-span-8">
        <!-- Task header -->
        <header
          class="relative overflow-hidden rounded-box border border-base-300/70 bg-base-100 p-6"
        >
          <span
            aria-hidden="true"
            class="pointer-events-none absolute -right-24 -top-24 h-56 w-56 rounded-full bg-primary/10 blur-3xl"
          />
          <div class="relative space-y-3">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <h1 class="text-[26px] font-semibold leading-tight tracking-tight text-base-content">
                {{ task.title }}
              </h1>
              <StatusBadge :status="task.status" size="md" />
            </div>
            <div class="flex flex-wrap items-center gap-2 text-xs text-base-content/55">
              <UiBadge tone="neutral" size="xs">
                <svg class="h-3 w-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M4 7h16M4 12h16M4 17h10" /></svg>
                {{ task.category }}
              </UiBadge>
              <span class="font-mono">{{ task.taskId }}</span>
              <span class="text-base-content/30">·</span>
              <span>创建于 {{ formatDateTime(task.createdAt) }}</span>
              <template v-if="task.lastActivityAt">
                <span class="text-base-content/30">·</span>
                <span>最近活跃 {{ formatDateTime(task.lastActivityAt) }}</span>
              </template>
            </div>
            <p
              class="whitespace-pre-line rounded-field border border-base-300/60 bg-base-200/40 p-4 text-sm leading-relaxed text-base-content/80"
            >
              {{ task.description }}
            </p>
          </div>
        </header>

        <ActionBar
          stage-label="当前阶段"
          :stage-hint="STAGE_HINT[task.status]"
          :actions="actions"
          :empty-hint="emptyActionHint"
          @run="runAction"
        />

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
                <path d="M22 12A10 10 0 0 1 2 12" />
                <path d="M12 2v10l5 3" />
              </svg>
              <h2 class="text-sm font-semibold tracking-tight">事件时间线</h2>
            </div>
            <span class="text-xs text-base-content/55">
              共 {{ timelineEvents.length }} 条
            </span>
          </header>
          <Timeline :events="timelineEvents" empty-text="暂无活动记录。" />
        </section>
      </div>

      <aside class="space-y-4 md:col-span-4">
        <section
          class="relative overflow-hidden rounded-box border border-primary/20 bg-[linear-gradient(135deg,color-mix(in_oklch,var(--color-primary)_14%,var(--color-base-100)),var(--color-base-100)_70%)] p-5"
        >
          <span
            aria-hidden="true"
            class="pointer-events-none absolute -right-12 -top-12 h-32 w-32 rounded-full bg-primary/25 blur-3xl"
          />
          <p class="relative text-[10.5px] uppercase tracking-[0.14em] text-base-content/60 font-semibold">
            报酬
          </p>
          <p class="relative mt-2 text-[32px] font-semibold tracking-tight leading-none gradient-text">
            {{ formatReward(task.reward.amount, task.reward.currency) }}
          </p>
          <p class="relative mt-1 text-xs text-base-content/60">
            {{ rewardModeLabel }}
          </p>
        </section>

        <section class="rounded-box border border-base-300/70 bg-base-100 p-4">
          <dl class="space-y-2.5 text-xs">
            <div class="flex items-center justify-between">
              <dt class="text-base-content/50">截止时间</dt>
              <dd class="font-medium text-base-content/80">
                {{ task.deadline ? formatDate(task.deadline) : '未设置' }}
              </dd>
            </div>
            <div class="flex items-center justify-between">
              <dt class="text-base-content/50">最近活跃</dt>
              <dd class="font-medium text-base-content/80">
                {{ task.lastActivityAt ? formatDateTime(task.lastActivityAt) : '—' }}
              </dd>
            </div>
            <div class="flex items-center justify-between">
              <dt class="text-base-content/50">更新于</dt>
              <dd class="font-medium text-base-content/80">
                {{ formatDateTime(task.updatedAt) }}
              </dd>
            </div>
          </dl>
        </section>

        <section class="rounded-box border border-base-300/70 bg-base-100 p-4">
          <div class="flex items-center justify-between">
            <p class="text-[10.5px] uppercase tracking-[0.12em] text-base-content/55 font-semibold">
              验收规则
            </p>
            <UiBadge tone="neutral" size="xs">{{ acceptanceModeLabel }}</UiBadge>
          </div>
          <ul
            v-if="task.acceptanceSpec.rules.length > 0"
            class="mt-2.5 space-y-1.5 text-xs text-base-content/75"
          >
            <li
              v-for="r in task.acceptanceSpec.rules"
              :key="r"
              class="flex items-start gap-2"
            >
              <svg
                class="mt-0.5 h-3.5 w-3.5 shrink-0 text-success"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              <span>{{ r }}</span>
            </li>
          </ul>
          <p v-else class="mt-2 text-xs text-base-content/40">未配置验收规则。</p>
        </section>

        <div class="space-y-3">
          <RoleCard role="需求方" :account="task.requester" />
          <RoleCard
            role="验收方"
            :account="task.reviewer ?? null"
            hint="由需求方默认担任"
          />
          <RoleCard
            role="执行方"
            :account="task.assignedExecutor ?? null"
            hint="尚未指派"
          />
        </div>
      </aside>
    </div>
  </section>
</template>
