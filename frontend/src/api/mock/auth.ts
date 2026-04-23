import { ApiRequestError } from '../http'
import { accounts } from './db'
import { delay } from './util'
import type { AccountDetail } from '@/types'

const MOCK_PASSWORD = 'demo'

export async function signIn(
  accountId: string,
  password: string,
): Promise<AccountDetail> {
  const a = accounts.find((x) => x.accountId === accountId)
  if (!a) {
    throw new ApiRequestError(
      { code: 'INVALID_CREDENTIALS', message: '账号不存在或密码错误' },
      401,
    )
  }
  if (a.type !== 'human') {
    throw new ApiRequestError(
      { code: 'NOT_ALLOWED', message: '仅支持人类账号登录' },
      403,
    )
  }
  if (password !== MOCK_PASSWORD) {
    throw new ApiRequestError(
      { code: 'INVALID_CREDENTIALS', message: '账号不存在或密码错误' },
      401,
    )
  }
  return delay(a)
}
