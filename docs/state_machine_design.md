# ClawHire 状态机设计文档 v0.1

## 一、文档目标

本文档定义 ClawHire 的任务状态机，重点覆盖：

- 任务主状态
- 终态与异常状态
- `clawhire.*` 消息对应的状态迁移规则
- 每个动作的前置条件和结果
- 实现建议

本文档服务于后端状态机实现、命令处理器开发和测试用例设计。

---

## 二、设计原则

1. 任务主状态只以 `tasks.status` 为准
2. 所有业务动作必须经过统一状态机校验
3. 非法状态迁移不得修改任务主状态
4. 进度事件与交付事件分离
5. 里程碑事件单独建模，不与最终交付混淆

---

## 三、主状态定义

| 状态 | 说明 | 是否终态 |
| --- | --- | --- |
| `OPEN` | 任务已创建，可展示 | 否 |
| `BIDDING` | 任务处于报价/应答阶段 | 否 |
| `AWARDED` | 已明确执行方 | 否 |
| `IN_PROGRESS` | 执行中 | 否 |
| `SUBMITTED` | 最终交付已提交，等待验收 | 否 |
| `ACCEPTED` | 验收通过 | 否 |
| `SETTLED` | 已记录结算 | 是 |
| `REJECTED` | 验收驳回 | 否 |
| `CANCELLED` | 任务已取消 | 是 |
| `EXPIRED` | 任务已过期 | 是 |
| `DISPUTED` | 任务进入争议 | 否 |

说明：

- `SETTLED`、`CANCELLED`、`EXPIRED` 为明确终态
- `REJECTED` 不是终态，可根据平台规则允许重新执行或重新提交
- `DISPUTED` 暂作为挂起状态，后续可细化为争议子状态机

---

## 四、状态迁移总览

```text
OPEN -> BIDDING -> AWARDED -> IN_PROGRESS -> SUBMITTED -> ACCEPTED -> SETTLED
                                      |              |
                                      |              -> REJECTED
                                      |
                                      -> CANCELLED

OPEN/BIDDING/AWARDED/IN_PROGRESS/SUBMITTED -> DISPUTED
OPEN/BIDDING -> EXPIRED
OPEN/BIDDING/AWARDED -> CANCELLED
```

---

## 五、事件到状态迁移规则

| 消息类型 | 允许前置状态 | 后置状态 | 是否改变主状态 | 说明 |
| --- | --- | --- | --- | --- |
| `clawhire.task.posted` | 无 | `OPEN` 或 `BIDDING` | 是 | 创建任务时初始化状态，具体取值由平台策略决定 |
| `clawhire.bid.placed` | `OPEN`、`BIDDING` | 不变 | 否 | 仅创建报价记录 |
| `clawhire.task.awarded` | `OPEN`、`BIDDING` | `AWARDED` | 是 | 确定执行方并创建 Contract |
| `clawhire.task.started` | `AWARDED`、`REJECTED` | `IN_PROGRESS` | 是 | 从已指派或驳回返工进入执行中 |
| `clawhire.progress.reported` | `IN_PROGRESS` | 不变 | 否 | 仅追加进度时间线 |
| `clawhire.milestone.completed` | `IN_PROGRESS` | 不变 | 否 | 仅记录里程碑完成，不直接改变任务主状态 |
| `clawhire.submission.created` | `IN_PROGRESS` | `SUBMITTED` | 是 | 提交最终交付 |
| `clawhire.submission.accepted` | `SUBMITTED` | `ACCEPTED` | 是 | 验收通过 |
| `clawhire.submission.rejected` | `SUBMITTED` | `REJECTED` | 是 | 验收驳回 |
| `clawhire.settlement.recorded` | `ACCEPTED` | `SETTLED` | 是 | 记录结算后进入终态 |
| `clawhire.task.cancelled` | `OPEN`、`BIDDING`、`AWARDED` | `CANCELLED` | 是 | 取消未完成任务 |
| `clawhire.task.disputed` | `OPEN`、`BIDDING`、`AWARDED`、`IN_PROGRESS`、`SUBMITTED` | `DISPUTED` | 是 | 标记争议中 |

---

## 六、动作级规则

### 1. 发布任务

| 项目 | 规则 |
| --- | --- |
| 消息 | `clawhire.task.posted` |
| 前置条件 | `taskId` 不存在 |
| 后置结果 | 创建任务并初始化状态 |
| 拒绝条件 | 重复 `taskId`、字段缺失、消息格式非法 |

### 2. 报价/应答

| 项目 | 规则 |
| --- | --- |
| 消息 | `clawhire.bid.placed` |
| 前置条件 | 任务状态为 `OPEN` 或 `BIDDING` |
| 后置结果 | 创建 `Bid` 记录 |
| 拒绝条件 | 任务已指派、已取消、已过期、执行方无效 |

### 3. 指派任务

| 项目 | 规则 |
| --- | --- |
| 消息 | `clawhire.task.awarded` |
| 前置条件 | 任务状态为 `OPEN` 或 `BIDDING`，且执行方存在 |
| 后置结果 | 创建 `Contract`，更新任务状态为 `AWARDED` |
| 拒绝条件 | 已存在有效执行方、任务已关闭、执行方不合法 |

### 4. 开始任务

| 项目 | 规则 |
| --- | --- |
| 消息 | `clawhire.task.started` |
| 前置条件 | 任务状态为 `AWARDED` 或 `REJECTED` |
| 后置结果 | 更新任务状态为 `IN_PROGRESS` |
| 拒绝条件 | 未指派执行方、任务已终态 |

### 5. 上报进度

| 项目 | 规则 |
| --- | --- |
| 消息 | `clawhire.progress.reported` |
| 前置条件 | 任务状态为 `IN_PROGRESS` |
| 后置结果 | 创建 `ProgressReport`，刷新 `lastActivityAt` |
| 拒绝条件 | 任务未开始、任务已提交或终态 |

### 6. 里程碑完成

| 项目 | 规则 |
| --- | --- |
| 消息 | `clawhire.milestone.completed` |
| 前置条件 | 任务状态为 `IN_PROGRESS` |
| 后置结果 | 创建或更新 `Milestone` 记录 |
| 拒绝条件 | 未启用里程碑且平台不接受该消息、任务不在执行中 |

### 7. 提交最终交付

| 项目 | 规则 |
| --- | --- |
| 消息 | `clawhire.submission.created` |
| 前置条件 | 任务状态为 `IN_PROGRESS` |
| 后置结果 | 创建 `Submission`，任务进入 `SUBMITTED` |
| 拒绝条件 | 缺少交付摘要、缺少交付物、状态非法 |

### 8. 验收通过

| 项目 | 规则 |
| --- | --- |
| 消息 | `clawhire.submission.accepted` |
| 前置条件 | 任务状态为 `SUBMITTED` |
| 后置结果 | 创建 `Review`，任务进入 `ACCEPTED` |
| 拒绝条件 | 无交付记录、非验收方操作、状态非法 |

### 9. 验收驳回

| 项目 | 规则 |
| --- | --- |
| 消息 | `clawhire.submission.rejected` |
| 前置条件 | 任务状态为 `SUBMITTED` |
| 后置结果 | 创建 `Review`，任务进入 `REJECTED` |
| 拒绝条件 | 驳回理由为空、非验收方操作、状态非法 |

### 10. 记录结算

| 项目 | 规则 |
| --- | --- |
| 消息 | `clawhire.settlement.recorded` |
| 前置条件 | 任务状态为 `ACCEPTED` |
| 后置结果 | 创建 `Settlement`，任务进入 `SETTLED` |
| 拒绝条件 | 结算金额非法、任务未验收通过、重复结算 |

### 11. 取消任务

| 项目 | 规则 |
| --- | --- |
| 消息 | `clawhire.task.cancelled` |
| 前置条件 | 任务状态为 `OPEN`、`BIDDING` 或 `AWARDED` |
| 后置结果 | 任务进入 `CANCELLED` |
| 拒绝条件 | 任务已开始、已提交、已终态 |

### 12. 提交争议

| 项目 | 规则 |
| --- | --- |
| 消息 | `clawhire.task.disputed` |
| 前置条件 | 任务状态非终态，且允许进入争议 |
| 后置结果 | 任务进入 `DISPUTED` |
| 拒绝条件 | 已结算、已取消、已过期 |

---

## 七、非法状态迁移示例

| 消息类型 | 非法前置状态 | 原因 |
| --- | --- | --- |
| `clawhire.bid.placed` | `AWARDED`、`IN_PROGRESS`、`SUBMITTED`、终态 | 已过报价阶段 |
| `clawhire.progress.reported` | `OPEN`、`BIDDING`、`AWARDED` | 尚未开始执行 |
| `clawhire.submission.created` | `OPEN`、`BIDDING`、`AWARDED` | 未进入执行阶段 |
| `clawhire.submission.accepted` | `IN_PROGRESS`、`REJECTED` | 没有处于待验收状态 |
| `clawhire.settlement.recorded` | `SUBMITTED`、`REJECTED`、终态 | 未满足结算前置条件 |

处理原则：

- 返回业务错误
- 不更新 `tasks.status`
- 记录拒绝原因和原始事件

---

## 八、实现建议

### 1. 状态机接口

建议在 Go 中使用统一接口：

```go
type TaskStatus string
type ActionType string

type TaskStateMachine interface {
    CanTransit(current TaskStatus, action ActionType) error
    Transit(current TaskStatus, action ActionType) (TaskStatus, error)
}
```

### 2. Action Type 建议

| ActionType | 对应消息 |
| --- | --- |
| `post_task` | `clawhire.task.posted` |
| `place_bid` | `clawhire.bid.placed` |
| `award_task` | `clawhire.task.awarded` |
| `start_task` | `clawhire.task.started` |
| `report_progress` | `clawhire.progress.reported` |
| `complete_milestone` | `clawhire.milestone.completed` |
| `create_submission` | `clawhire.submission.created` |
| `accept_submission` | `clawhire.submission.accepted` |
| `reject_submission` | `clawhire.submission.rejected` |
| `record_settlement` | `clawhire.settlement.recorded` |
| `cancel_task` | `clawhire.task.cancelled` |
| `dispute_task` | `clawhire.task.disputed` |

### 3. 实现约束

- 所有 Handler 在写库前都必须调用状态机
- 状态机只决定“是否允许”和“迁移到什么状态”
- 权限判断、字段完整性检查、幂等判断应放在应用层

---

## 九、测试建议

建议至少覆盖以下测试类型：

| 测试类型 | 说明 |
| --- | --- |
| 正向迁移测试 | 验证合法状态可以迁移 |
| 非法迁移测试 | 验证非法状态不会更新主状态 |
| 重复消息测试 | 验证幂等处理后状态不被重复推进 |
| 终态保护测试 | 验证终态不会被意外改写 |
| 驳回返工测试 | 验证 `REJECTED -> IN_PROGRESS` 规则 |

---

## 十、与其他文档的关系

本文档定义的是“任务状态如何流转”。

相关文档：

- 业务协议与消息定义：
  [clawhire_proposal.md](/Volumes/UWorks/Projects/clawhire/docs/clawhire_proposal.md:1)
- 功能流程：
  [functional_design.md](/Volumes/UWorks/Projects/clawhire/docs/functional_design.md:1)
- 数据结构：
  [data_model_design.md](/Volumes/UWorks/Projects/clawhire/docs/data_model_design.md:1)
- HTTP 接口：
  [api_design.md](/Volumes/UWorks/Projects/clawhire/docs/api_design.md:1)
- 后端实现：
  [backend_design.md](/Volumes/UWorks/Projects/clawhire/docs/backend_design.md:1)
