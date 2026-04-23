import { httpGet, httpGetPaginated } from './http'
import { USE_MOCK } from './config'
import * as mock from './mock/tasks'
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
  return httpGetPaginated<TaskListItem>('/tasks', { params: query })
}

export async function getTask(taskId: string): Promise<TaskDetail> {
  if (USE_MOCK) return mock.getTask(taskId)
  return httpGet<TaskDetail>(`/tasks/${taskId}`)
}
