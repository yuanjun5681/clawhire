export type PlatformKind = 'trustmesh'

export interface PlatformConnection {
  id: string
  platform: PlatformKind
  platformNodeId: string
  localUserId: string
  remoteUserId: string
  linkedAt: string
}

export interface CreateConnectionPayload {
  platform: PlatformKind
  remoteUserId: string
  platformNodeId?: string
}
