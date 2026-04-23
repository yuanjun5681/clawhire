import { httpPost } from './http'
import { USE_MOCK } from './config'
import * as mock from './mock/auth'
import type { AccountDetail } from '@/types'

export async function signIn(
  accountId: string,
  password: string,
): Promise<AccountDetail> {
  if (USE_MOCK) return mock.signIn(accountId, password)
  return httpPost<{ accountId: string; password: string }, AccountDetail>(
    '/auth/sign-in',
    { accountId, password },
  )
}
