import { httpGet } from './http'
import { USE_MOCK } from './config'
import * as mock from './mock/taskResources'
import {
  normalizeBid,
  normalizeProgress,
  normalizeReview,
  normalizeSettlement,
  normalizeSubmission,
} from './normalizers'
import type { Bid, Progress, Review, Settlement, Submission } from '@/types'

export async function listBids(taskId: string): Promise<Bid[]> {
  if (USE_MOCK) return mock.listBids(taskId)
  const res = await httpGet<Bid[]>(`/tasks/${taskId}/bids`)
  return res.map(normalizeBid)
}

export async function listProgress(taskId: string): Promise<Progress[]> {
  if (USE_MOCK) return mock.listProgress(taskId)
  const res = await httpGet<Progress[]>(`/tasks/${taskId}/progress`)
  return res.map(normalizeProgress)
}

export async function listSubmissions(taskId: string): Promise<Submission[]> {
  if (USE_MOCK) return mock.listSubmissions(taskId)
  const res = await httpGet<Submission[]>(`/tasks/${taskId}/submissions`)
  return res.map(normalizeSubmission)
}

export async function listReviews(taskId: string): Promise<Review[]> {
  if (USE_MOCK) return mock.listReviews(taskId)
  const res = await httpGet<Review[]>(`/tasks/${taskId}/reviews`)
  return res.map(normalizeReview)
}

export async function listSettlements(taskId: string): Promise<Settlement[]> {
  if (USE_MOCK) return mock.listSettlements(taskId)
  const res = await httpGet<Settlement[]>(`/tasks/${taskId}/settlements`)
  return res.map(normalizeSettlement)
}
