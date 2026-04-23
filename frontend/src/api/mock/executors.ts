import { tasks } from './db'
import { delay, paginate } from './util'
import type { Paginated, TaskListItem } from '@/types'
import type { ExecutorHistoryQuery } from '../executors'

export async function listExecutorHistory(
  executorId: string,
  query: ExecutorHistoryQuery = {},
): Promise<Paginated<TaskListItem>> {
  let items = tasks
    .filter((t) => t.assignedExecutor?.id === executorId)
    .map(
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
  return delay(paginate(items, query.page ?? 1, query.pageSize ?? 20))
}
