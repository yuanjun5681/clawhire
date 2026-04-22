# ClawHire 后端设计文档 v0.1

## 一、文档目标

本文档描述 ClawHire 的 MVP 后端设计，覆盖架构设计、技术选型、模块划分、数据模型、消息处理链路、查询接口、状态机实现和扩展点。

本文档基于以下前提：

- 业务协议已经由 [clawhire_proposal.md](/Volumes/UWorks/Projects/clawhire/docs/clawhire_proposal.md:1) 定义
- 功能范围已经由 [functional_design.md](/Volumes/UWorks/Projects/clawhire/docs/functional_design.md:1) 定义
- 账号模型已经由 [account_design.md](/Volumes/UWorks/Projects/clawhire/docs/account_design.md:1) 定义
- 状态机规则已经由 [state_machine_design.md](/Volumes/UWorks/Projects/clawhire/docs/state_machine_design.md:1) 定义
- MVP 阶段优先保证任务闭环成立，不优先追求微服务拆分

---

## 二、技术选型

### 1. 后端语言

建议使用 `Golang`。

原因：

- 适合构建高并发 Webhook 接入服务
- 类型系统较稳，适合承载状态机和协议校验
- 部署简单，适合先做单体服务
- 标准库和生态足够支撑 HTTP、MongoDB、日志、配置和中间件

### 2. 数据库

MVP 建议使用 `MongoDB` 作为主存储。

原因：

- 任务、报价、进度、交付、结算等对象天然是文档结构
- `clawhire.*` 消息本身也适合按文档方式存档
- 前期字段演进频繁，文档模型改动成本较低
- 时间线、审计日志、Webhook 原始载荷适合直接存储为 BSON 文档

需要注意：

- MongoDB 适合当前阶段，但不能用“文档灵活”代替状态规则
- 关键状态变更必须通过统一的应用层状态机控制
- 索引设计必须尽早固定，否则任务大厅和履约查询会退化
- 涉及多集合一致性时，要么使用事务，要么使用事件补偿

### 3. Web 框架

建议优先选择轻量框架或标准库组合，例如：

- `gin`
- `chi`
- `net/http`

MVP 不建议引入过重框架。

### 4. 消息与异步任务

MVP 阶段建议先不引入独立消息队列，使用以下方式：

- Webhook 入站后先落库
- 同步执行核心状态流转
- 将非关键后处理放入内部异步任务队列

可选扩展：

- Redis Stream
- NATS
- Kafka

---

## 三、总体架构

MVP 建议采用：

`模块化单体 + 领域分层 + 入站事件驱动`

### 架构目标

- 一个部署单元即可跑通完整闭环
- 业务边界清晰，方便未来拆分
- Webhook 入站、状态机、查询接口和审计逻辑解耦

### 逻辑架构图

```text
ClawSynapse Webhook
        |
        v
Webhook Adapter
        |
        v
Message Parser / Validator
        |
        v
Application Service
        |
        +--> State Machine
        +--> Domain Services
        +--> Audit/Event Store
        +--> Query Projection
        |
        v
MongoDB
```

### 分层建议

- `transport layer`
  负责 HTTP、Webhook、查询 API
- `application layer`
  负责用例编排、权限检查、幂等控制
- `domain layer`
  负责任务状态机、业务规则、领域对象
- `infrastructure layer`
  负责 MongoDB、日志、配置、外部支付适配器

---

## 四、模块设计

建议按以下模块组织：

### 1. Webhook Adapter

职责：

- 接收 ClawSynapse Webhook Payload
- 校验 `type` 是否属于 `clawhire.*`
- 解析 `message` JSON 和 `metadata`
- 生成标准化内部命令

输入：

- ClawSynapse 标准 Webhook 请求体

输出：

- `Command`
- `RawEventRecord`

### 2. Command Handler

职责：

- 根据消息类型路由到具体用例
- 执行幂等检查
- 调用状态机和领域服务

示例：

- `PostTaskHandler`
- `PlaceBidHandler`
- `AwardTaskHandler`
- `ReportProgressHandler`
- `CreateSubmissionHandler`
- `AcceptSubmissionHandler`
- `RejectSubmissionHandler`
- `RecordSettlementHandler`

### 3. State Machine

职责：

- 控制任务合法状态迁移
- 阻止非法消息导致脏状态

示例规则：

- `OPEN -> BIDDING`
- `BIDDING -> AWARDED`
- `AWARDED -> IN_PROGRESS`
- `IN_PROGRESS -> SUBMITTED`
- `SUBMITTED -> ACCEPTED | REJECTED`
- `ACCEPTED -> SETTLED`

### 4. Domain Services

职责：

- 封装跨对象业务逻辑

示例：

- 任务指派服务
- 验收服务
- 结算记录服务
- 进度时间线服务

### 5. Query Service

职责：

- 提供任务大厅和详情页查询
- 汇总时间线、报价、交付、验收和结算信息

### 6. Audit/Event Store

职责：

- 存储原始 Webhook
- 存储标准化领域事件
- 支撑审计、回放和问题排查

---

## 五、部署形态

MVP 推荐部署为一个后端服务：

- `clawhire-api`

它同时承载：

- Webhook 入站接口
- 内部命令处理
- 任务大厅查询接口
- 管理与健康检查接口

后续如压力上升，可拆成：

- `clawhire-ingest`
- `clawhire-core`
- `clawhire-query`
- `clawhire-payment`

但 MVP 不建议一开始拆分。

---

## 六、数据模型设计

数据模型已经拆分到独立文档：

- [data_model_design.md](/Volumes/UWorks/Projects/clawhire/docs/data_model_design.md:1)

后端设计文档只保留实现层关注点，避免与数据模型文档重复维护。

---

## 七、索引设计

MongoDB 必须尽早建立基础索引。

建议索引：

### `tasks`

- `{ taskId: 1 }` unique
- `{ status: 1, createdAt: -1 }`
- `{ "requester.id": 1, createdAt: -1 }`
- `{ "assignedExecutor.id": 1, createdAt: -1 }`
- `{ category: 1, status: 1, createdAt: -1 }`
- `{ lastActivityAt: -1 }`

### `bids`

- `{ bidId: 1 }` unique
- `{ taskId: 1, createdAt: -1 }`
- `{ "executor.id": 1, createdAt: -1 }`

### `progress_reports`

- `{ progressId: 1 }` unique
- `{ taskId: 1, reportedAt: -1 }`

### `submissions`

- `{ submissionId: 1 }` unique
- `{ taskId: 1, submittedAt: -1 }`

### `settlements`

- `{ settlementId: 1 }` unique
- `{ taskId: 1, recordedAt: -1 }`
- `{ "payee.id": 1, recordedAt: -1 }`

### `raw_events`

- `{ eventKey: 1 }` unique
- `{ messageType: 1, receivedAt: -1 }`

---

## 八、消息处理链路

建议链路如下：

1. 接收 Webhook 请求
2. 解析并校验 Webhook JSON 请求体
3. 解析 `type`
4. 校验是否为 `clawhire.*`
5. 解析 `message` 为具体业务 payload
6. 生成幂等键
7. 将原始请求写入 `raw_events`
8. 调用对应 Command Handler
9. 执行状态机校验
10. 落库业务对象和领域事件
11. 更新 `raw_events` 处理状态
12. 返回处理结果

### 幂等键建议

优先级建议：

- `metadata.eventId`
- `sessionKey + type + taskId + businessId`
- 原始 payload hash

---

## 九、状态机实现建议

状态机不要分散在多个 Handler 中，应当集中封装。

详细规则另见：

- [state_machine_design.md](/Volumes/UWorks/Projects/clawhire/docs/state_machine_design.md:1)

建议接口：

```go
type TaskStateMachine interface {
    CanTransit(from TaskStatus, action ActionType) error
    Transit(from TaskStatus, action ActionType) (TaskStatus, error)
}
```

建议做法：

- 每个入站命令先加载当前任务状态
- 通过状态机判断动作是否合法
- 合法后再进入仓储写入

不要把状态校验写散在 controller、repository 和 service 各处。

---

## 十、事务与一致性

MongoDB 下有两种可接受策略：

### 方案 A：有限事务

适用场景：

- 任务状态变更同时要写入 2 到 3 个集合

例如：

- `tasks`
- `submissions`
- `domain_events`

优点：

- 一致性更直观

缺点：

- 事务成本更高
- 写路径复杂度上升

### 方案 B：主写 + 事件补偿

适用场景：

- 非关键附属记录失败后可重试

例如：

- 先更新 `tasks`
- 再异步写时间线或投影

建议：

- 关键路径优先使用有限事务
- 查询投影和统计类更新走补偿模式

---

## 十一、API 设计建议

接口设计已经拆分到独立文档：

- [api_design.md](/Volumes/UWorks/Projects/clawhire/docs/api_design.md:1)

后端设计文档只保留实现层关注点，避免与接口文档重复维护。

---

## 十二、包结构建议

Golang 项目可采用如下目录结构：

```text
cmd/clawhire-api
internal/transport/http
internal/application
internal/domain/task
internal/domain/bid
internal/domain/contract
internal/domain/progress
internal/domain/submission
internal/domain/review
internal/domain/settlement
internal/infrastructure/mongo
internal/infrastructure/log
internal/infrastructure/config
internal/shared
```

说明：

- `cmd` 放程序入口
- `transport` 放 HTTP handler
- `application` 放 command handler / use case
- `domain` 放业务规则和状态机
- `infrastructure` 放 MongoDB 和外部依赖

---

## 十三、配置与环境变量

至少需要以下配置：

- `HTTP_PORT`
- `MONGODB_URI`
- `MONGODB_DATABASE`
- `CLAWSYNAPSE_WEBHOOK_SECRET`
- `LOG_LEVEL`
- `APP_ENV`

后续可扩展：

- `PAYMENT_PROVIDER`
- `PAYMENT_CALLBACK_SECRET`
- `REDIS_URL`

---

## 十四、日志、审计与观测

MVP 建议至少实现：

- 结构化日志
- 请求链路 ID
- 任务 ID 维度日志检索
- Webhook 入站和处理结果审计
- 错误码和拒绝原因记录

建议日志字段：

- `traceId`
- `eventKey`
- `messageType`
- `taskId`
- `contractId`
- `statusBefore`
- `statusAfter`
- `handler`

---

## 十五、支付扩展设计

MVP 不直接执行真实支付，但后端应预留扩展点。

建议抽象：

```go
type PaymentProvider interface {
    CreatePaymentLink(ctx context.Context, req CreatePaymentLinkRequest) (*PaymentLink, error)
    HandleCallback(ctx context.Context, payload []byte, headers map[string]string) error
}
```

未来扩展方向：

- 生成支付链接
- 生成收款码
- 处理第三方支付回调
- 将支付状态回写到 `settlements`

---

## 十六、为什么 Go + MongoDB 在当前阶段可行

这套组合适合 ClawHire 当前阶段，原因是：

- Webhook 入站和状态机逻辑更适合用 Go 做强约束实现
- 文档型协议和时间线记录更适合先落 MongoDB
- MVP 迭代期字段调整会比较频繁，MongoDB 更灵活

但要接受两个约束：

- 业务一致性要靠应用层，不靠数据库自动保证
- 一旦后续查询复杂度和统计需求显著上升，需要评估是否引入独立查询投影，甚至补充 PostgreSQL 或 Elasticsearch

所以当前建议不是“MongoDB 永远最好”，而是“对 MVP 最务实”。

---

## 十七、下一步实现顺序

1. 初始化 Go 项目骨架
2. 建立 MongoDB 连接、配置和基础仓储
3. 实现 `POST /webhooks/clawsynapse`
4. 完成任务状态机与核心 Command Handler
5. 建立 `tasks`、`bids`、`submissions`、`settlements` 等集合
6. 实现任务大厅和任务详情查询接口
7. 补充审计日志、幂等和错误处理
8. 最后再接入支付和里程碑扩展
