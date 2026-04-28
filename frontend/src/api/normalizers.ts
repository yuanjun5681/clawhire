import type {
  AccountDetail,
  AccountListItem,
  AccountSummary,
  Bid,
  Progress,
  Review,
  Settlement,
  Submission,
  TaskDetail,
  TaskListItem,
} from '@/types'

type RawActor = {
  id: string
  kind?: string
  name?: string
  nodeId?: string
}

type RawArtifact = {
  type?: string
  url?: string
  name?: string
}

type RawEvidence = {
  type?: string
  items?: string[]
}

function normalizeKind(kind?: string): AccountSummary['kind'] {
  return kind === 'agent' ? 'agent' : 'human'
}

export function normalizeActor(actor?: RawActor | null): AccountSummary | undefined {
  if (!actor?.id) return undefined
  return {
    id: actor.id,
    kind: normalizeKind(actor.kind),
    name: actor.name?.trim() || actor.id,
    nodeId: actor.nodeId,
  }
}

function normalizeArtifacts(artifacts?: RawArtifact[]): Progress['artifacts'] {
  if (!artifacts?.length) return undefined
  return artifacts.map((item) => {
    const url = item.url?.trim()
    return {
      name: item.name?.trim() || url || item.type || '附件',
      url,
      type: item.type,
    }
  })
}

function normalizeEvidence(evidence?: RawEvidence | null): Submission['evidence'] {
  if (!evidence?.items?.length) return undefined
  return evidence.items.map((item, index) => ({
    name: evidence.type ? `${evidence.type} ${index + 1}` : `证据 ${index + 1}`,
    url: item,
  }))
}

function normalizeBidStatus(status?: string): Bid['status'] {
  switch (status) {
    case 'awarded':
      return 'accepted'
    case 'rejected':
      return 'rejected'
    case 'withdrawn':
      return 'withdrawn'
    default:
      return 'pending'
  }
}

function normalizeSubmissionStatus(status?: string): Submission['status'] {
  switch (status) {
    case 'accepted':
      return 'accepted'
    case 'rejected':
      return 'rejected'
    default:
      return 'pending_review'
  }
}

function normalizeReviewDecision(decision?: string): Review['decision'] {
  return decision === 'rejected' ? 'rejected' : 'approved'
}

function normalizeSettlementStatus(status?: string): Settlement['status'] {
  switch (status) {
    case 'recorded':
    case 'paid':
      return 'settled'
    case 'failed':
      return 'failed'
    case 'refunded':
      return 'cancelled'
    case 'pending_payment':
      return 'processing'
    default:
      return 'pending'
  }
}

export function normalizeAccountListItem(item: AccountListItem): AccountListItem {
  return {
    ...item,
    nodeId: item.nodeId || undefined,
    ownerAccountId: item.ownerAccountId || undefined,
  }
}

export function normalizeAccountDetail(item: AccountDetail): AccountDetail {
  return {
    ...normalizeAccountListItem(item),
    updatedAt: item.updatedAt,
  }
}

export function normalizeTaskListItem(item: TaskListItem): TaskListItem {
  return {
    ...item,
    requester:
      normalizeActor(item.requester as unknown as RawActor) ?? {
        id: 'unknown',
        kind: 'human',
        name: '未知需求方',
      },
  }
}

export function normalizeTaskDetail(item: TaskDetail): TaskDetail {
  const spec = item.acceptanceSpec ?? { mode: 'manual', rules: [] }
  return {
    ...item,
    requester: normalizeActor(item.requester as unknown as RawActor)!,
    reviewer: normalizeActor(item.reviewer as unknown as RawActor),
    assignedExecutor: normalizeActor(item.assignedExecutor as unknown as RawActor),
    acceptanceSpec: { ...spec, rules: spec.rules ?? [] },
  }
}

export function normalizeBid(item: Bid): Bid {
  return {
    ...item,
    executor: normalizeActor(item.executor as unknown as RawActor)!,
    status: normalizeBidStatus(item.status),
  }
}

export function normalizeProgress(item: Progress): Progress {
  return {
    ...item,
    executor: normalizeActor(item.executor as unknown as RawActor)!,
    artifacts: normalizeArtifacts(item.artifacts as unknown as RawArtifact[]),
  }
}

export function normalizeSubmission(item: Submission): Submission {
  return {
    ...item,
    executor: normalizeActor(item.executor as unknown as RawActor) ?? {
      id: 'unknown',
      kind: 'human',
      name: '未知执行方',
    },
    artifacts: normalizeArtifacts(item.artifacts as unknown as RawArtifact[]),
    evidence: normalizeEvidence(item.evidence as unknown as RawEvidence),
    status: normalizeSubmissionStatus(item.status),
  }
}

export function normalizeReview(item: Review): Review {
  return {
    ...item,
    reviewer: normalizeActor(item.reviewer as unknown as RawActor)!,
    decision: normalizeReviewDecision(item.decision),
  }
}

export function normalizeSettlement(item: Settlement): Settlement {
  return {
    ...item,
    payee: normalizeActor(item.payee as unknown as RawActor)!,
    status: normalizeSettlementStatus(item.status),
  }
}
