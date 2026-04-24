import { ApiRequestError } from '../http'
import { accounts } from './db'
import { delay } from './util'
import type { AuthResult, LoginInput, RegisterInput } from '../auth'

const MOCK_PASSWORD = 'demo'

function buildResult(accountId: string, displayName: string): AuthResult {
  return {
    token: `mock.${accountId}.${Date.now()}`,
    expiresAt: new Date(Date.now() + 24 * 3600 * 1000).toISOString(),
    account: {
      accountId,
      type: 'human',
      displayName,
      status: 'active',
    },
  }
}

export async function login({ accountId, password }: LoginInput): Promise<AuthResult> {
  const a = accounts.find((x) => x.accountId === accountId)
  if (!a) {
    throw new ApiRequestError(
      { code: 'UNAUTHORIZED', message: '账号不存在或密码错误' },
      401,
    )
  }
  if (a.type !== 'human') {
    throw new ApiRequestError(
      { code: 'FORBIDDEN', message: '仅支持人类账号登录' },
      403,
    )
  }
  if (password !== MOCK_PASSWORD) {
    throw new ApiRequestError(
      { code: 'UNAUTHORIZED', message: '账号不存在或密码错误' },
      401,
    )
  }
  return delay(buildResult(a.accountId, a.displayName))
}

export async function register({ accountId, displayName, password }: RegisterInput): Promise<AuthResult> {
  if (!/^[a-zA-Z0-9_\-.]{3,64}$/.test(accountId)) {
    throw new ApiRequestError(
      { code: 'INVALID_REQUEST', message: '账号 ID 仅支持 3-64 位字母数字和 _-.' },
      400,
    )
  }
  if (password.length < 8) {
    throw new ApiRequestError(
      { code: 'INVALID_REQUEST', message: '密码至少 8 位' },
      400,
    )
  }
  const fullId = `acct_human_${accountId}`
  if (accounts.some((x) => x.accountId === fullId)) {
    throw new ApiRequestError(
      { code: 'CONFLICT', message: '账号 ID 已被占用' },
      409,
    )
  }
  return delay(buildResult(fullId, displayName.trim() || accountId))
}
