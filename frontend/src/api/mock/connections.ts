import { ApiRequestError } from '../http'
import { delay } from './util'
import type { PlatformConnection, CreateConnectionPayload } from '@/types'

const store: PlatformConnection[] = []

export async function listConnections(platform?: string): Promise<PlatformConnection[]> {
  const items = platform ? store.filter((c) => c.platform === platform) : [...store]
  return delay(items)
}

export async function createConnection(
  payload: CreateConnectionPayload,
): Promise<PlatformConnection> {
  const nodeId = payload.platformNodeId ?? 'node_default_trustmesh'
  const exists = store.some(
    (c) => c.platform === payload.platform && c.platformNodeId === nodeId,
  )
  if (exists) {
    throw new ApiRequestError(
      { code: 'CONFLICT', message: '该平台账号已绑定，请勿重复添加' },
      409,
    )
  }
  const conn: PlatformConnection = {
    id: `conn_${Date.now()}`,
    platform: payload.platform,
    platformNodeId: nodeId,
    localUserId: 'user_001',
    remoteUserId: payload.remoteUserId,
    linkedAt: new Date().toISOString(),
  }
  store.push(conn)
  return delay(conn)
}

export async function deleteConnection(
  platform: string,
  platformNodeId?: string,
): Promise<void> {
  const idx = store.findIndex(
    (c) =>
      c.platform === platform &&
      (!platformNodeId || c.platformNodeId === platformNodeId),
  )
  if (idx === -1) {
    throw new ApiRequestError(
      { code: 'NOT_FOUND', message: '连接不存在' },
      404,
    )
  }
  store.splice(idx, 1)
  return delay(undefined)
}
