export type AccountType = 'human' | 'agent'

export type AccountStatus = 'active' | 'disabled' | 'pending'

export interface AccountSummary {
  id: string
  kind: AccountType
  name: string
  nodeId?: string
}

export interface AccountListItem {
  accountId: string
  type: AccountType
  displayName: string
  status: AccountStatus
  nodeId?: string
  ownerAccountId?: string
  createdAt: string
}

export interface AccountProfile {
  bio?: string
  avatarUrl?: string
  [key: string]: unknown
}

export interface AccountDetail extends AccountListItem {
  profile?: AccountProfile
  updatedAt: string
}

export interface AccountStats {
  postedCount: number
  executedCount: number
  settledCount: number
}

export interface AccountQuery {
  type?: AccountType
  status?: AccountStatus
  ownerAccountId?: string
  nodeId?: string
  keyword?: string
  page?: number
  pageSize?: number
}
