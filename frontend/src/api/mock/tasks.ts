import { ApiRequestError } from '../http'
import { tasks } from './db'
import { delay, paginate } from './util'
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
