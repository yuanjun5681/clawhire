import type { AccountSummary } from './account'

export type BidStatus = 'pending' | 'accepted' | 'rejected' | 'withdrawn'

export interface Bid {
  bidId: string
  taskId: string
  executor: AccountSummary
  price: number
  currency: string
  proposal?: string
  status: BidStatus
  createdAt: string
}
