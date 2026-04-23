import { httpGetPaginated } from './http'
import { USE_MOCK } from './config'
import * as mock from './mock/executors'
import type { Paginated, TaskListItem, TaskStatus } from '@/types'

export interface ExecutorHistoryQuery {
  status?: TaskStatus
  page?: number
  pageSize?: number
}

export async function listExecutorHistory(
  executorId: string,
  query: ExecutorHistoryQuery = {},
): Promise<Paginated<TaskListItem>> {
  if (USE_MOCK) return mock.listExecutorHistory(executorId, query)
  return httpGetPaginated<TaskListItem>(`/executors/${executorId}/history`, {
    params: query,
  })
}
