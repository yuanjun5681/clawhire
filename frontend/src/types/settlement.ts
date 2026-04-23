import type { AccountSummary } from './account'

export type SettlementStatus =
  | 'pending'
  | 'processing'
  | 'settled'
  | 'failed'
  | 'cancelled'

export interface Settlement {
  settlementId: string
  taskId: string
  payee: AccountSummary
  amount: number
  currency: string
  status: SettlementStatus
  channel?: string
  externalRef?: string
  recordedAt: string
}
