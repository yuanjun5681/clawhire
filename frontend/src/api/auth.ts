import { USE_MOCK } from './config'
import * as mock from './mock/auth'
import { httpPost } from './http'
import type { AccountStatus, AccountType } from '@/types'

export interface AuthResult {
  token: string
  expiresAt: string
  account: {
    accountId: string
    type: AccountType
    displayName: string
    status: AccountStatus
  }
}

export interface LoginInput {
  accountId: string
  password: string
}

export interface RegisterInput {
  accountId: string
  displayName: string
  password: string
}

export async function login(input: LoginInput): Promise<AuthResult> {
  if (USE_MOCK) return mock.login(input)
  return httpPost<LoginInput, AuthResult>('/auth/login', input)
}

export async function register(input: RegisterInput): Promise<AuthResult> {
  if (USE_MOCK) return mock.register(input)
  return httpPost<RegisterInput, AuthResult>('/auth/register', input)
}
