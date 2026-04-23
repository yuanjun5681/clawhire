import { httpGet } from './http'
import { USE_MOCK } from './config'
import * as mock from './mock/taskResources'
import type { Bid, Progress, Review, Settlement, Submission } from '@/types'

export async function listBids(taskId: string): Promise<Bid[]> {
  if (USE_MOCK) return mock.listBids(taskId)
  return httpGet<Bid[]>(`/tasks/${taskId}/bids`)
}

export async function listProgress(taskId: string): Promise<Progress[]> {
  if (USE_MOCK) return mock.listProgress(taskId)
  return httpGet<Progress[]>(`/tasks/${taskId}/progress`)
}

export async function listSubmissions(taskId: string): Promise<Submission[]> {
  if (USE_MOCK) return mock.listSubmissions(taskId)
  return httpGet<Submission[]>(`/tasks/${taskId}/submissions`)
}

export async function listReviews(taskId: string): Promise<Review[]> {
  if (USE_MOCK) return mock.listReviews(taskId)
  return httpGet<Review[]>(`/tasks/${taskId}/reviews`)
}

export async function listSettlements(taskId: string): Promise<Settlement[]> {
  if (USE_MOCK) return mock.listSettlements(taskId)
  return httpGet<Settlement[]>(`/tasks/${taskId}/settlements`)
}
