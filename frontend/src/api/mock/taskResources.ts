import { bids, progress, reviews, settlements, submissions } from './db'
import { delay } from './util'
import type { Bid, Progress, Review, Settlement, Submission } from '@/types'

export async function listBids(taskId: string): Promise<Bid[]> {
  return delay(bids[taskId] ?? [])
}

export async function listProgress(taskId: string): Promise<Progress[]> {
  return delay(progress[taskId] ?? [])
}

export async function listSubmissions(taskId: string): Promise<Submission[]> {
  return delay(submissions[taskId] ?? [])
}

export async function listReviews(taskId: string): Promise<Review[]> {
  return delay(reviews[taskId] ?? [])
}

export async function listSettlements(taskId: string): Promise<Settlement[]> {
  return delay(settlements[taskId] ?? [])
}
