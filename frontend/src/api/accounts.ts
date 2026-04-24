import { httpGet, httpGetPaginated } from './http'
import { USE_MOCK } from './config'
import * as mock from './mock/accounts'
import {
  normalizeAccountDetail,
  normalizeAccountListItem,
} from './normalizers'
import type {
  AccountDetail,
  AccountListItem,
  Paginated,
  AccountQuery,
} from '@/types'

export async function listAccounts(
  query: AccountQuery = {},
): Promise<Paginated<AccountListItem>> {
  if (USE_MOCK) return mock.listAccounts(query)
  const res = await httpGetPaginated<AccountListItem>('/accounts', { params: query })
  return {
    ...res,
    items: res.items.map(normalizeAccountListItem),
  }
}

export async function getAccount(accountId: string): Promise<AccountDetail> {
  if (USE_MOCK) return mock.getAccount(accountId)
  const res = await httpGet<AccountDetail>(`/accounts/${accountId}`)
  return normalizeAccountDetail(res)
}

export async function listAccountAgents(
  accountId: string,
): Promise<AccountListItem[]> {
  if (USE_MOCK) return mock.listAccountAgents(accountId)
  const res = await httpGetPaginated<AccountListItem>(`/accounts/${accountId}/agents`)
  return res.items.map(normalizeAccountListItem)
}
