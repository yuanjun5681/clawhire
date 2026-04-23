import { ApiRequestError } from '../http'
import { accounts, tasks } from './db'
import { delay, paginate } from './util'
import type {
  AccountDetail,
  AccountListItem,
  AccountQuery,
  AccountStats,
  Paginated,
  TaskListItem,
} from '@/types'

function toListItem(a: AccountDetail): AccountListItem {
  const { updatedAt: _u, profile: _p, ...rest } = a
  return rest
}

export async function listAccounts(
  query: AccountQuery = {},
): Promise<Paginated<AccountListItem>> {
  let items = accounts.map(toListItem)
  if (query.type) items = items.filter((a) => a.type === query.type)
  if (query.status) items = items.filter((a) => a.status === query.status)
  if (query.ownerAccountId)
    items = items.filter((a) => a.ownerAccountId === query.ownerAccountId)
  if (query.nodeId) items = items.filter((a) => a.nodeId === query.nodeId)
  if (query.keyword) {
    const kw = query.keyword.toLowerCase()
    items = items.filter(
      (a) =>
        a.displayName.toLowerCase().includes(kw) ||
        a.accountId.toLowerCase().includes(kw),
    )
  }
  return delay(paginate(items, query.page ?? 1, query.pageSize ?? 20))
}

export async function getAccount(accountId: string): Promise<AccountDetail> {
  const a = accounts.find((x) => x.accountId === accountId)
  if (!a)
    throw new ApiRequestError(
      { code: 'NOT_FOUND', message: `account not found: ${accountId}` },
      404,
    )
  return delay(a)
}

export async function listAccountAgents(
  accountId: string,
): Promise<AccountListItem[]> {
  return delay(
    accounts
      .filter((a) => a.type === 'agent' && a.ownerAccountId === accountId)
      .map(toListItem),
  )
}

export async function getAccountStats(
  accountId: string,
): Promise<AccountStats> {
  const posted = tasks.filter((t) => t.requester.id === accountId).length
  const executed = tasks.filter(
    (t) => t.assignedExecutor?.id === accountId,
  ).length
  const settled = tasks.filter(
    (t) =>
      (t.requester.id === accountId || t.assignedExecutor?.id === accountId) &&
      t.status === 'SETTLED',
  ).length
  return delay({
    postedCount: posted,
    executedCount: executed,
    settledCount: settled,
  })
}

export async function getAccountRecentTasks(
  accountId: string,
): Promise<TaskListItem[]> {
  const related = tasks
    .filter(
      (t) =>
        t.requester.id === accountId || t.assignedExecutor?.id === accountId,
    )
    .sort(
      (a, b) =>
        new Date(b.lastActivityAt ?? 0).getTime() -
        new Date(a.lastActivityAt ?? 0).getTime(),
    )
    .slice(0, 10)
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
  return delay(related)
}
