import type { AccountSummary } from './account'
import type { ProgressArtifact } from './progress'

export type SubmissionStatus = 'pending_review' | 'accepted' | 'rejected'

export interface SubmissionEvidence {
  name: string
  url?: string
  note?: string
}

export interface Submission {
  submissionId: string
  taskId: string
  executor: AccountSummary
  summary: string
  artifacts?: ProgressArtifact[]
  evidence?: SubmissionEvidence[]
  status: SubmissionStatus
  submittedAt: string
}
