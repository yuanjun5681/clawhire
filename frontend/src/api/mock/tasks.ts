import { ApiRequestError } from '../http'
import { accounts, bids, tasks } from './db'
import { delay, paginate } from './util'
import type {
  AcceptSubmissionInput,
  AwardTaskInput,
  CreateBidInput,
  CreateSubmissionInput,
  CreateTaskInput,
  RecordSettlementInput,
  RejectSubmissionInput,
} from '../tasks'
import type {
  Paginated,
  TaskDetail,
  TaskListItem,
  TaskQuery,
} from '@/types'

export async function listTasks(
  query: TaskQuery = {},
): Promise<Paginated<TaskListItem>> {
  let items: TaskListItem[] = tasks.map(
    ({
      taskId,
      title,
      category,
      status,
      requester,
      reward,
      deadline,
      lastActivityAt,
    }) => ({
      taskId,
      title,
      category,
      status,
      requester,
      reward,
      deadline,
      lastActivityAt,
    }),
  )

  if (query.status) items = items.filter((t) => t.status === query.status)
  if (query.category)
    items = items.filter((t) => t.category === query.category)
  if (query.requesterId)
    items = items.filter((t) => t.requester.id === query.requesterId)
  if (query.executorId) {
    const executorTaskIds = tasks
      .filter((t) => t.assignedExecutor?.id === query.executorId)
      .map((t) => t.taskId)
    items = items.filter((t) => executorTaskIds.includes(t.taskId))
  }
  if (query.reviewerId) {
    const reviewerTaskIds = tasks
      .filter((t) => t.reviewer?.id === query.reviewerId)
      .map((t) => t.taskId)
    items = items.filter((t) => reviewerTaskIds.includes(t.taskId))
  }
  if (query.keyword) {
    const kw = query.keyword.toLowerCase()
    items = items.filter(
      (t) =>
        t.title.toLowerCase().includes(kw) ||
        t.category.toLowerCase().includes(kw),
    )
  }

  items.sort(
    (a, b) =>
      new Date(b.lastActivityAt ?? 0).getTime() -
      new Date(a.lastActivityAt ?? 0).getTime(),
  )

  return delay(paginate(items, query.page ?? 1, query.pageSize ?? 20))
}

export async function getTask(taskId: string): Promise<TaskDetail> {
  const t = tasks.find((x) => x.taskId === taskId)
  if (!t)
    throw new ApiRequestError(
      { code: 'NOT_FOUND', message: `task not found: ${taskId}` },
      404,
    )
  return delay(t)
}

export async function createTask(
  input: CreateTaskInput,
): Promise<{ taskId: string; eventId?: string }> {
  if (tasks.find((x) => x.taskId === input.taskId)) {
    throw new ApiRequestError(
      { code: 'INVALID_REQUEST', message: 'taskId already exists' },
      400,
    )
  }
  const requester = accounts[0]
  const reviewerAccount = input.reviewerId
    ? accounts.find((a) => a.accountId === input.reviewerId)
    : undefined
  const now = new Date().toISOString()
  const detail: TaskDetail = {
    taskId: input.taskId,
    title: input.title,
    description: input.description ?? '',
    category: input.category,
    status: 'OPEN',
    requester: {
      id: requester.accountId,
      kind: requester.type,
      name: requester.displayName,
      nodeId: requester.nodeId,
    },
    reviewer: reviewerAccount
      ? {
          id: reviewerAccount.accountId,
          kind: reviewerAccount.type,
          name: reviewerAccount.displayName,
          nodeId: reviewerAccount.nodeId,
        }
      : {
          id: requester.accountId,
          kind: requester.type,
          name: requester.displayName,
          nodeId: requester.nodeId,
        },
    reward: input.reward,
    acceptanceSpec: input.acceptanceSpec ?? { mode: 'manual', rules: [] },
    deadline: input.deadline,
    lastActivityAt: now,
    createdAt: now,
    updatedAt: now,
  }
  tasks.unshift(detail)
  return delay({ taskId: input.taskId })
}

export async function createBid(
  taskId: string,
  input: CreateBidInput,
): Promise<{ taskId: string; bidId: string }> {
  const t = tasks.find((x) => x.taskId === taskId)
  if (!t) throw new ApiRequestError({ code: 'NOT_FOUND', message: 'task not found' }, 404)
  const taskBids = (bids[taskId] ??= [])
  taskBids.push({
    bidId: input.bidId,
    taskId,
    executor: { id: accounts[0].accountId, kind: accounts[0].type, name: accounts[0].displayName },
    price: input.price,
    currency: input.currency,
    proposal: input.proposal,
    status: 'pending',
    createdAt: new Date().toISOString(),
  })
  if (t.status === 'OPEN') t.status = 'BIDDING'
  t.updatedAt = new Date().toISOString()
  return delay({ taskId, bidId: input.bidId })
}

export async function awardTask(
  taskId: string,
  input: AwardTaskInput,
): Promise<{ taskId: string; contractId: string }> {
  const t = tasks.find((x) => x.taskId === taskId)
  if (!t) throw new ApiRequestError({ code: 'NOT_FOUND', message: 'task not found' }, 404)
  const executorAccount = accounts.find((a) => a.accountId === input.executorId)
  if (!executorAccount)
    throw new ApiRequestError({ code: 'NOT_FOUND', message: 'executor account not found' }, 404)
  t.assignedExecutor = {
    id: executorAccount.accountId,
    kind: executorAccount.type,
    name: executorAccount.displayName,
    nodeId: executorAccount.nodeId,
  }
  t.status = 'AWARDED'
  t.reward = {
    ...t.reward,
    amount: input.agreedReward.amount,
    currency: input.agreedReward.currency,
  }
  const now = new Date().toISOString()
  t.assignedAt = now
  t.updatedAt = now
  const taskBids = bids[taskId] ?? []
  for (const b of taskBids) {
    b.status = b.executor.id === input.executorId ? 'accepted' : 'rejected'
  }
  return delay({ taskId, contractId: input.contractId })
}

export async function createSubmission(
  taskId: string,
  input: CreateSubmissionInput,
): Promise<{ taskId: string; submissionId: string }> {
  const t = tasks.find((x) => x.taskId === taskId)
  if (!t) throw new ApiRequestError({ code: 'NOT_FOUND', message: 'task not found' }, 404)
  const { submissions } = await import('./db')
  const list = (submissions[taskId] ??= [])
  list.unshift({
    submissionId: input.submissionId,
    taskId,
    executor: t.assignedExecutor ?? { id: 'unknown', kind: 'human', name: 'Unknown' },
    summary: input.summary,
    finalOutput: input.finalOutput,
    artifacts: input.artifacts.map((a) => ({ name: a.name ?? a.url, url: a.url })),
    status: 'pending_review',
    submittedAt: new Date().toISOString(),
  })
  t.status = 'SUBMITTED'
  t.updatedAt = new Date().toISOString()
  return delay({ taskId, submissionId: input.submissionId })
}

export async function acceptSubmission(
  taskId: string,
  input: AcceptSubmissionInput,
): Promise<{ taskId: string; submissionId: string }> {
  const t = tasks.find((x) => x.taskId === taskId)
  if (!t) throw new ApiRequestError({ code: 'NOT_FOUND', message: 'task not found' }, 404)
  const { submissions } = await import('./db')
  const sub = (submissions[taskId] ?? []).find((s) => s.submissionId === input.submissionId)
  if (sub) sub.status = 'accepted'
  t.status = 'ACCEPTED'
  t.updatedAt = new Date().toISOString()
  return delay({ taskId, submissionId: input.submissionId })
}

export async function rejectSubmission(
  taskId: string,
  input: RejectSubmissionInput,
): Promise<{ taskId: string; submissionId: string }> {
  const t = tasks.find((x) => x.taskId === taskId)
  if (!t) throw new ApiRequestError({ code: 'NOT_FOUND', message: 'task not found' }, 404)
  const { submissions } = await import('./db')
  const sub = (submissions[taskId] ?? []).find((s) => s.submissionId === input.submissionId)
  if (sub) sub.status = 'rejected'
  t.status = 'REJECTED'
  t.updatedAt = new Date().toISOString()
  return delay({ taskId, submissionId: input.submissionId })
}

export async function recordSettlement(
  taskId: string,
  input: RecordSettlementInput,
): Promise<{ taskId: string; settlementId: string }> {
  const t = tasks.find((x) => x.taskId === taskId)
  if (!t) throw new ApiRequestError({ code: 'NOT_FOUND', message: 'task not found' }, 404)
  if (t.status !== 'ACCEPTED') {
    throw new ApiRequestError({ code: 'INVALID_STATE', message: 'task is not accepted' }, 409)
  }
  const { settlements } = await import('./db')
  const settlementId = input.settlementId || `stl_${Date.now()}`
  const payee = t.assignedExecutor
  if (!payee) {
    throw new ApiRequestError({ code: 'INVALID_STATE', message: 'missing payee' }, 409)
  }
  settlements[taskId] = [
    ...(settlements[taskId] ?? []),
    {
      settlementId,
      taskId,
      payee,
      amount: input.amount || t.reward.amount,
      currency: input.currency || t.reward.currency,
      status: 'settled',
      channel: input.channel || 'manual',
      externalRef: input.externalRef,
      recordedAt: input.recordedAt || new Date().toISOString(),
    },
  ]
  t.status = 'SETTLED'
  t.updatedAt = new Date().toISOString()
  return delay({ taskId, settlementId })
}
