import type { AccountSummary } from './account'

export interface ProgressArtifact {
  name: string
  url?: string
  type?: string
}

export interface Progress {
  progressId: string
  taskId: string
  executor: AccountSummary
  stage?: string
  percent?: number
  summary: string
  artifacts?: ProgressArtifact[]
  reportedAt: string
}
