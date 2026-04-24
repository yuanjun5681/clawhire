<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import {
  ApiRequestError,
  taskResourcesApi,
  tasksApi,
} from '@/api'
import { useIdentityStore } from '@/stores/identity'
import StatusBadge from '@/components/StatusBadge.vue'
import RoleCard from '@/components/RoleCard.vue'
import ActionBar from '@/components/ActionBar.vue'
import type { ActionItem } from '@/components/ActionBar.vue'
import Timeline from '@/components/Timeline.vue'
import type { TimelineEvent } from '@/components/Timeline.vue'
import ErrorState from '@/components/ErrorState.vue'
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

const taskId = computed(() => String(route.params.taskId ?? ''))

const loading = ref(true)
const apiError = ref<{ message: string; code?: string } | null>(null)
const task = ref<TaskDetail | null>(null)
const bids = ref<Bid[]>([])
const progress = ref<Progress[]>([])
const submissions = ref<Submission[]>([])
const reviews = ref<Review[]>([])
const settlements = ref<Settlement[]>([])
const toast = ref<string | null>(null)
const actionSubmitting = ref(false)

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
    case 'SETTLED':
    case 'CANCELLED':
    case 'EXPIRED':
    case 'ACCEPTED':
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

function flash(message: string) {
  toast.value = message
  window.setTimeout(() => {
    toast.value = null
  }, 2400)
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
        flash('报价已提交')
        break
      }
      case 'assign': {
        const executorId = window.prompt('执行方账号 ID')
        if (!executorId?.trim()) return
        const amountRaw = window.prompt(
          '约定金额',
          String(currentTask.reward.amount),
        )
        if (!amountRaw) return
        const amount = Number(amountRaw)
        if (!Number.isFinite(amount) || amount <= 0) throw new Error('约定金额无效')
        await tasksApi.awardTask(currentTask.taskId, {
          contractId: nextId('contract'),
          executorId: executorId.trim(),
          agreedReward: {
            amount,
            currency: currentTask.reward.currency || 'USD',
          },
        })
        flash('已完成指派')
        break
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
        flash('交付已提交')
        break
      }
      case 'approve': {
        if (!latestSubmission) throw new Error('当前没有可验收的提交记录')
        if (!window.confirm('确认通过当前交付？')) return
        await tasksApi.acceptSubmission(currentTask.taskId, {
          submissionId: latestSubmission.submissionId,
          acceptedAt: new Date().toISOString(),
        })
        flash('已通过验收')
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
        flash('已驳回交付')
        break
      }
      default:
        flash(`当前前端尚未接入操作：${key}`)
        return
    }

    await load()
  } catch (e: unknown) {
    flash(
      e instanceof ApiRequestError
        ? e.message
        : e instanceof Error
          ? e.message
          : '操作失败',
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

  if (t.assignedExecutor) {
    evs.push({
      id: `assign-${t.taskId}`,
      type: 'assign',
      title: `指派执行方 · ${t.assignedExecutor.name}`,
      actor: t.requester,
      at: t.updatedAt,
    })
  }

  for (const p of progress.value) {
    evs.push({
      id: `progress-${p.progressId}`,
      type: 'progress',
      title: p.stage
        ? `进度汇报 · ${p.stage}`
        : '进度汇报',
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
        ...(st.channel
          ? [{ label: '渠道', value: st.channel }]
          : []),
        ...(st.externalRef
          ? [{ label: '外部流水', value: st.externalRef }]
          : []),
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
  <section class="space-y-4">
    <nav class="text-xs text-base-content/50">
      <RouterLink to="/tasks" class="hover:text-primary">任务大厅</RouterLink>
      <span class="mx-1">/</span>
      <span class="text-base-content/70">任务详情</span>
    </nav>

    <div v-if="loading" class="grid grid-cols-1 gap-4 md:grid-cols-12">
      <div class="space-y-3 md:col-span-8">
        <div class="h-7 w-2/3 animate-pulse rounded bg-base-200" />
        <div class="h-4 w-1/3 animate-pulse rounded bg-base-200" />
        <div class="h-24 animate-pulse rounded-xl bg-base-200" />
        <div class="h-16 animate-pulse rounded-xl bg-base-200" />
        <div class="h-40 animate-pulse rounded-xl bg-base-200" />
      </div>
      <div class="space-y-3 md:col-span-4">
        <div class="h-24 animate-pulse rounded-xl bg-base-200" />
        <div class="h-24 animate-pulse rounded-xl bg-base-200" />
        <div class="h-24 animate-pulse rounded-xl bg-base-200" />
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
      class="grid grid-cols-1 gap-4 md:grid-cols-12"
    >
      <div class="space-y-4 md:col-span-8">
        <header
          class="rounded-xl border border-base-300 bg-base-100 p-4 space-y-2"
        >
          <div class="flex items-start justify-between gap-3">
            <h1 class="text-xl font-semibold tracking-tight text-base-content">
              {{ task.title }}
            </h1>
            <StatusBadge :status="task.status" />
          </div>
          <div
            class="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-base-content/55"
          >
            <span class="font-mono">{{ task.taskId }}</span>
            <span class="text-base-content/30">·</span>
            <span>分类 · {{ task.category }}</span>
            <span class="text-base-content/30">·</span>
            <span>创建于 {{ formatDateTime(task.createdAt) }}</span>
            <span
              v-if="task.lastActivityAt"
              class="text-base-content/30"
            >·</span>
            <span v-if="task.lastActivityAt">
              最近活跃 {{ formatDateTime(task.lastActivityAt) }}
            </span>
          </div>
          <p
            class="whitespace-pre-line text-sm leading-relaxed text-base-content/80"
          >
            {{ task.description }}
          </p>
        </header>

        <ActionBar
          stage-label="当前阶段"
          :stage-hint="STAGE_HINT[task.status]"
          :actions="actions"
          :empty-hint="emptyActionHint"
          @run="runAction"
        />

        <section
          class="rounded-xl border border-base-300 bg-base-100 p-4"
        >
          <header class="mb-3 flex items-center justify-between">
            <h2 class="text-sm font-medium text-base-content">事件时间线</h2>
            <span class="text-xs text-base-content/50">
              共 {{ timelineEvents.length }} 条
            </span>
          </header>
          <Timeline
            :events="timelineEvents"
            empty-text="暂无活动记录。"
          />
        </section>
      </div>

      <aside class="space-y-3 md:col-span-4">
        <section
          class="rounded-xl border border-base-300 bg-base-100 p-4 space-y-2"
        >
          <p
            class="text-[11px] uppercase tracking-wider text-base-content/50"
          >
            报酬
          </p>
          <p class="text-2xl font-semibold text-base-content">
            {{ formatReward(task.reward.amount, task.reward.currency) }}
          </p>
          <p class="text-xs text-base-content/55">
            {{ rewardModeLabel }}
          </p>
        </section>

        <section
          class="rounded-xl border border-base-300 bg-base-100 p-4 space-y-1.5 text-xs"
        >
          <div class="flex items-center justify-between">
            <span class="text-base-content/50">截止时间</span>
            <span class="text-base-content/80">
              {{ task.deadline ? formatDate(task.deadline) : '未设置' }}
            </span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-base-content/50">最近活跃</span>
            <span class="text-base-content/80">
              {{
                task.lastActivityAt
                  ? formatDateTime(task.lastActivityAt)
                  : '—'
              }}
            </span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-base-content/50">更新于</span>
            <span class="text-base-content/80">
              {{ formatDateTime(task.updatedAt) }}
            </span>
          </div>
        </section>

        <section
          class="rounded-xl border border-base-300 bg-base-100 p-4 space-y-2"
        >
          <div class="flex items-center justify-between">
            <p
              class="text-[11px] uppercase tracking-wider text-base-content/50"
            >
              验收规则
            </p>
            <span
              class="rounded bg-base-200 px-1.5 py-0.5 text-[10px] font-medium text-base-content/70"
            >{{ acceptanceModeLabel }}</span>
          </div>
          <ul
            v-if="task.acceptanceSpec.rules.length > 0"
            class="list-inside list-disc space-y-0.5 text-xs text-base-content/75"
          >
            <li v-for="r in task.acceptanceSpec.rules" :key="r">{{ r }}</li>
          </ul>
          <p v-else class="text-xs text-base-content/40">未配置验收规则。</p>
        </section>

        <div class="space-y-2">
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
