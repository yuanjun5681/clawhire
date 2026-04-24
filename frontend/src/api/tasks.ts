import { httpGet, httpGetPaginated, httpPost } from './http'
import { USE_MOCK } from './config'
import * as mock from './mock/tasks'
import { normalizeTaskDetail, normalizeTaskListItem } from './normalizers'
import type {
  Paginated,
  TaskDetail,
  TaskListItem,
  TaskQuery,
} from '@/types'

export async function listTasks(
  query: TaskQuery = {},
): Promise<Paginated<TaskListItem>> {
  if (USE_MOCK) return mock.listTasks(query)
  const res = await httpGetPaginated<TaskListItem>('/tasks', { params: query })
  return {
    ...res,
    items: res.items.map(normalizeTaskListItem),
  }
}

export async function getTask(taskId: string): Promise<TaskDetail> {
  if (USE_MOCK) return mock.getTask(taskId)
  const res = await httpGet<TaskDetail>(`/tasks/${taskId}`)
  return normalizeTaskDetail(res)
}

export interface CreateBidInput {
  bidId: string
  price: number
  currency: string
  proposal?: string
}

export interface AwardTaskInput {
  contractId: string
  executorId: string
  agreedReward: {
    amount: number
    currency: string
  }
}

export interface CreateSubmissionInput {
  submissionId: string
  contractId?: string
  summary: string
  artifacts: Array<{
    type: string
    value: string
    label?: string
  }>
  evidence?: {
    type: string
    items: string[]
  }
}

export interface AcceptSubmissionInput {
  submissionId: string
  acceptedAt?: string
}

export interface RejectSubmissionInput {
  submissionId: string
  reason: string
  rejectedAt?: string
}

export async function createBid(taskId: string, payload: CreateBidInput) {
  return httpPost<CreateBidInput, { taskId: string; bidId: string; eventId?: string }>(
    `/tasks/${taskId}/bids`,
    payload,
  )
}

export async function awardTask(taskId: string, payload: AwardTaskInput) {
  return httpPost<AwardTaskInput, { taskId: string; contractId: string; eventId?: string }>(
    `/tasks/${taskId}/award`,
    payload,
  )
}

export async function createSubmission(taskId: string, payload: CreateSubmissionInput) {
  return httpPost<
    CreateSubmissionInput,
    { taskId: string; submissionId: string; eventId?: string }
  >(`/tasks/${taskId}/submissions`, payload)
}

export async function acceptSubmission(taskId: string, payload: AcceptSubmissionInput) {
  return httpPost<
    AcceptSubmissionInput,
    { taskId: string; submissionId: string; eventId?: string }
  >(`/tasks/${taskId}/accept`, payload)
}

export async function rejectSubmission(taskId: string, payload: RejectSubmissionInput) {
  return httpPost<
    RejectSubmissionInput,
    { taskId: string; submissionId: string; eventId?: string }
  >(`/tasks/${taskId}/reject`, payload)
}
