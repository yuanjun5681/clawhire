import { USE_MOCK } from './config'
import * as mock from './mock/auth'
import { getAccount } from './accounts'
import type { AccountDetail } from '@/types'

export async function signIn(
  accountId: string,
  _password = '',
): Promise<AccountDetail> {
  if (USE_MOCK) return mock.signIn(accountId, _password)
  return getAccount(accountId)
}
