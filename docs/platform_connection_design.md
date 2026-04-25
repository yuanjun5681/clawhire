# ClawHire 跨平台连接设计文档 v0.1

## 一、文档目标

本文档定义 ClawHire 与外部平台（如 TrustMesh）之间的账号连接机制，重点覆盖：

- 跨平台身份绑定模型
- `platform_connections` 集合设计
- 事件发布时的身份解析流程
- 与 TrustMesh 数据模型的对称关系

本文档不描述具体业务流程，不重复 ClawSynapse 通信协议细节。

---

## 二、设计背景

ClawHire 与 TrustMesh 均基于 ClawSynapse 网络构建。ClawSynapse 是两个平台共同的通信总线，消息通过 ClawSynapse 节点路由投递。

两个平台各有独立的内部身份体系：

| 平台 | 内部用户标识 | ClawSynapse 身份 |
| --- | --- | --- |
| ClawHire | `accountId` | 平台节点 `nodeId`（配置项） |
| TrustMesh | `userId` | 平台节点 `nodeId`（配置项） |

跨平台事件投递需要解决两个问题：

1. **路由**：ClawHire 向哪个 ClawSynapse 节点发布消息 → 目标平台的平台级 `nodeId`，作为配置项维护，不随账号变化
2. **身份解析**：消息到达对方平台后，关联到哪个本地用户 → 依赖 `platform_connections` 绑定关系

---

## 三、设计原则

1. 跨平台绑定使用独立集合，不嵌入 `accounts` 主文档
2. 一个账号对同一平台只允许绑定一次
3. 绑定关系由用户主动发起，平台不自动推断
4. 两个平台的连接集合结构保持对称，便于跨平台协作理解
5. MVP 阶段不做握手验证，由用户自行输入对方平台的 `userId`

---

## 四、`platform_connections` 集合设计

用途：

- 记录 ClawHire 账号与外部平台账号的绑定关系
- 支持事件发布时解析目标节点与用户身份
- 支持接收外部平台事件时反查本地账号

字段定义：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `_id` | ObjectId | 是 | MongoDB 主键 |
| `platform` | string | 是 | 外部平台标识，如 `trustmesh` |
| `platformNodeId` | string | 是 | 对方平台的 ClawSynapse nodeId；创建时默认取对应平台的环境变量配置值 |
| `localUserId` | string | 是 | 本平台账号 ID，对应 `accounts.accountId` |
| `remoteUserId` | string | 是 | 对方平台的用户 ID |
| `linkedAt` | datetime | 是 | 绑定时间 |

说明：

- `platform` 为枚举型字符串，扩展新平台时添加新值即可
- `platformNodeId` 默认由后端填入环境变量中的平台节点配置值；用户可在绑定时显式指定，以支持同一平台类型在不同节点上的多实例（如多个独立部署的 TrustMesh）
- `localUserId` + `platformNodeId` 组合唯一，同一账号可绑定同类平台的不同节点实例，但不能重复绑定同一节点
- `remoteUserId` 含义由 `platform` 决定：对 `trustmesh` 而言是 TrustMesh 的 `userId`

---

## 五、示例文档

```json
{
  "_id": "ObjectId",
  "platform": "trustmesh",
  "platformNodeId": "node_trustmesh_prod",
  "localUserId": "acct_human_alice",
  "remoteUserId": "usr_xxxx",
  "linkedAt": "2026-04-25T10:00:00Z"
}
```

---

## 六、索引设计

| 集合 | 索引 | 类型 | 用途 |
| --- | --- | --- | --- |
| `platform_connections` | `{ localUserId: 1, platformNodeId: 1 }` | unique | 防止同一账号重复绑定同一节点实例 |
| `platform_connections` | `{ platform: 1, localUserId: 1 }` | normal | 按平台类型查询某账号的所有连接 |
| `platform_connections` | `{ platformNodeId: 1, remoteUserId: 1 }` | normal | 收到外部事件时反查本地账号 |

---

## 七、Go 结构定义

```go
// internal/domain/account/platform_connection.go

type PlatformConnection struct {
    ID             primitive.ObjectID `bson:"_id"`
    Platform       string             `bson:"platform"`
    PlatformNodeID string             `bson:"platformNodeId"`
    LocalUserID    string             `bson:"localUserId"`
    RemoteUserID   string             `bson:"remoteUserId"`
    LinkedAt       time.Time          `bson:"linkedAt"`
}
```

---

## 八、事件发布时的身份解析流程

ClawHire 向外部平台发布事件时（如任务授权后通知 TrustMesh），身份解析步骤如下：

```
handleTaskAwarded(awardeeAccountId)
    ↓
查询 platform_connections
  where platform = "trustmesh"
    and localUserId = awardeeAccountId
    ↓
找到 → { platformNodeId: "node_trustmesh_prod", remoteUserId: "usr_xxxx" }
找不到 → 跳过发布（graceful degradation，不影响主流程）
    ↓
Publisher.Publish(
  targetNode: conn.PlatformNodeID,              // 直接取连接记录中的 nodeId
  type:       "clawhire.task.awarded",
  message:    { taskId, title, ... },
  metadata: {
    clawhireAccountId: awardeeAccountId,         // 对方反查用
    trustmeshUserId:   conn.RemoteUserID,        // 对方直接定位用
  }
)
```

说明：

- `targetNode` 直接取 `platform_connections.platformNodeId`，无需再读全局配置；支持同一用户绑定多个不同节点的同类平台实例
- 新建绑定时，若用户未显式指定 `platformNodeId`，后端自动填入环境变量中对应平台的默认节点 ID
- 消息 `metadata` 同时携带双方 ID，对方平台可按需选择解析方式
- 未绑定账号不阻塞 ClawHire 主业务流程

---

## 九、接收外部事件时的反查流程

TrustMesh 向 ClawHire 发送事件时（如 todo 完成后同步提交物），ClawHire 通过以下方式定位本地账号：

```
收到 webhook 消息
  metadata.trustmeshUserId = "usr_xxxx"
    ↓
查询 platform_connections
  where platformNodeId = 来源节点 nodeId
    and remoteUserId   = "usr_xxxx"
    ↓
找到 → localUserId = "acct_human_alice"
找不到 → 拒绝处理，返回 4xx
```

---

## 十、与 TrustMesh 的对称结构

TrustMesh 侧应维护结构对称的集合：

```
platform_connections（TrustMesh）
  _id
  platform          "clawhire"
  platform_node_id  ClawHire 平台的 ClawSynapse nodeId
  local_user_id     TrustMesh userId
  remote_user_id    ClawHire accountId
  linked_at
```

索引：`unique({ local_user_id, platform_node_id })`

两个平台的集合逻辑完全对称，字段语义一致，仅命名风格按各自约定（ClawHire camelCase，TrustMesh snake_case）。

---

## 十一、API 设计（ClawHire 侧）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/accounts/me/connections` | 查询当前账号已绑定的平台列表 |
| `POST` | `/api/accounts/me/connections` | 新增平台绑定 |
| `DELETE` | `/api/accounts/me/connections/:platform` | 解除指定平台绑定 |

`POST` 请求体：

```json
{
  "platform": "trustmesh",
  "remoteUserId": "usr_xxxx",
  "platformNodeId": "node_trustmesh_prod"
}
```

说明：

- `platformNodeId` 为可选字段；若不填，后端自动使用环境变量中该平台的默认节点 ID
- 填入自定义 `platformNodeId` 可绑定同一平台类型的非默认节点实例

---

## 十二、配置项

| 环境变量 | 说明 |
| --- | --- |
| `TRUSTMESH_PLATFORM_NODE_ID` | TrustMesh 平台的 ClawSynapse nodeId，事件路由目标 |
| `CLAWSYNAPSE_NODE_API_URL` | 本地 ClawSynapse 节点 API 地址，用于出站发布 |
| `CLAWSYNAPSE_SELF_NODE_ID` | ClawHire 自身的 ClawSynapse nodeId |

---

## 十三、职责边界总结

### 由 ClawHire 维护

- `platform_connections` 集合
- 绑定关系的增删查
- 出站事件发布时的身份解析

### 由 TrustMesh 维护

- 对称的 `platform_connections` 集合
- 入站事件时的本地账号定位
- 项目/任务的创建与同步

### 由 ClawSynapse 维护

- 节点间消息路由与投递
- 平台节点身份认证

---

## 十四、后续可扩展项

MVP 阶段不实现，预留扩展：

- 绑定时双向握手验证（防止用户填错对方 userId）
- 绑定状态字段（`active / revoked`）
- 绑定变更审计日志
