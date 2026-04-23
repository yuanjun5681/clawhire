import { httpGet, httpGetPaginated } from './http'
import { USE_MOCK } from './config'
import * as mock from './mock/accounts'
import type {
  AccountDetail,
  AccountListItem,
  AccountQuery,
  AccountStats,
  Paginated,
  TaskListItem,
} from '@/types'

export async function listAccounts(
  query: AccountQuery = {},
): Promise<Paginated<AccountListItem>> {
  if (USE_MOCK) return mock.listAccounts(query)
  return httpGetPaginated<AccountListItem>('/accounts', { params: query })
}

export async function getAccount(accountId: string): Promise<AccountDetail> {
  if (USE_MOCK) return mock.getAccount(accountId)
  return httpGet<AccountDetail>(`/accounts/${accountId}`)
}

export async function listAccountAgents(
  accountId: string,
): Promise<AccountListItem[]> {
  if (USE_MOCK) return mock.listAccountAgents(accountId)
  return httpGet<AccountListItem[]>(`/accounts/${accountId}/agents`)
}

export async function getAccountStats(
  accountId: string,
): Promise<AccountStats> {
  if (USE_MOCK) return mock.getAccountStats(accountId)
  return httpGet<AccountStats>(`/accounts/${accountId}/stats`)
}

export async function getAccountRecentTasks(
  accountId: string,
): Promise<TaskListItem[]> {
  if (USE_MOCK) return mock.getAccountRecentTasks(accountId)
  return httpGet<TaskListItem[]>(`/accounts/${accountId}/tasks/recent`)
}
