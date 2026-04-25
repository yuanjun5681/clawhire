import { httpDelete, httpGet, httpPost } from './http'
import { USE_MOCK } from './config'
import * as mock from './mock/connections'
import type { PlatformConnection, CreateConnectionPayload } from '@/types'

export async function listConnections(platform?: string): Promise<PlatformConnection[]> {
  if (USE_MOCK) return mock.listConnections(platform)
  const res = await httpGet<PlatformConnection[] | null>('/accounts/me/connections', {
    params: platform ? { platform } : undefined,
  })
  return res ?? []
}

export async function createConnection(
  payload: CreateConnectionPayload,
): Promise<PlatformConnection> {
  if (USE_MOCK) return mock.createConnection(payload)
  return httpPost<CreateConnectionPayload, PlatformConnection>(
    '/accounts/me/connections',
    payload,
  )
}

export async function deleteConnection(
  platform: string,
  platformNodeId?: string,
): Promise<void> {
  if (USE_MOCK) return mock.deleteConnection(platform, platformNodeId)
  await httpDelete(`/accounts/me/connections/${platform}`, {
    params: platformNodeId ? { platformNodeId } : undefined,
  })
}
