import type { AccountSummary } from './account'

export const TASK_STATUSES = [
  'OPEN',
  'BIDDING',
  'AWARDED',
  'IN_PROGRESS',
  'SUBMITTED',
  'ACCEPTED',
  'SETTLED',
  'REJECTED',
  'CANCELLED',
  'EXPIRED',
  'DISPUTED',
] as const

export type TaskStatus = (typeof TASK_STATUSES)[number]

export type TaskCategory = string

export type RewardMode = 'fixed' | 'bid' | 'milestone'

export interface Reward {
  mode: RewardMode
  amount: number
  currency: string
}

export type AcceptanceMode = 'manual' | 'schema' | 'test' | 'hybrid'

export interface AcceptanceSpec {
  mode: AcceptanceMode
  rules: string[]
}

export interface TaskListItem {
  taskId: string
  title: string
  category: TaskCategory
  status: TaskStatus
  requester: AccountSummary
  reward: Reward
  deadline?: string
  lastActivityAt?: string
}

export interface TaskDetail extends TaskListItem {
  description: string
  reviewer?: AccountSummary
  assignedExecutor?: AccountSummary
  assignedAt?: string
  acceptanceSpec: AcceptanceSpec
  createdAt: string
  updatedAt: string
}

export interface TaskQuery {
  status?: TaskStatus
  category?: TaskCategory
  requesterId?: string
  executorId?: string
  reviewerId?: string
  keyword?: string
  page?: number
  pageSize?: number
}
