import type { AccountSummary } from './account'

export type ReviewDecision = 'approved' | 'rejected'

export interface Review {
  reviewId: string
  taskId: string
  submissionId?: string
  reviewer: AccountSummary
  decision: ReviewDecision
  reason?: string
  reviewedAt: string
}
