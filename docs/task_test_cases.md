# 任务测试用例

本文档提供以文案设计等文本生成为主的任务测试用例。

---

## 任务一：产品介绍文案

| 字段 | 值 |
|------|-----|
| **TaskID** | `task_001` |
| **Title** | 为智能手环撰写产品介绍文案 |
| **Description** | 为公司新款智能手环撰写 300 字以内的产品介绍文案，突出健康监测功能，语言简洁有力，适合电商平台展示。 |
| **Category** | 文案撰写 |
| **Status** | `OPEN` |
| **Requester** | `Actor{ID: "user_alice", Kind: "user", Name: "Alice"}` |
| **Reward** | `Reward{Mode: "fixed", Amount: 200.00, Currency: "CNY"}` |
| **AcceptanceSpec** | `AcceptanceSpec{Mode: "manual"}` |
| **Deadline** | 2026-05-01 23:59:59 |

---

## 任务二：品牌故事

| 字段 | 值 |
|------|-----|
| **TaskID** | `task_002` |
| **Title** | 咖啡品牌故事文案 |
| **Description** | 为精品咖啡品牌写一篇 500 字左右的品牌故事，突出从豆子到杯子的匠心过程，情感真挚，可用于官网和社交媒体。 |
| **Category** | 文案撰写 |
| **Status** | `BIDDING` |
| **Requester** | `Actor{ID: "user_alice", Kind: "user", Name: "Alice"}` |
| **Reward** | `Reward{Mode: "bid", Amount: 0, Currency: "CNY"}` |
| **AcceptanceSpec** | `AcceptanceSpec{Mode: "manual"}` |
| **Deadline** | 2026-05-10 23:59:59 |

---

## 任务三：社交媒体文案集

| 字段 | 值 |
|------|-----|
| **TaskID** | `task_003` |
| **Title** | 端午节社交媒体推广文案 |
| **Description** | 为端午节营销活动撰写微博、小红书、微信公众号三个平台的推广文案，每个平台不少于 3 条，风格符合平台调性。 |
| **Category** | 文案撰写 |
| **Status** | `OPEN` |
| **Requester** | `Actor{ID: "user_alice", Kind: "user", Name: "Alice"}` |
| **Reward** | `Reward{Mode: "fixed", Amount: 500.00, Currency: "CNY"}` |
| **AcceptanceSpec** | `AcceptanceSpec{Mode: "schema", Rules: ["字数符合平台要求", "包含端午节元素", "有明确行动号召"]}` |
| **Deadline** | 2026-04-28 23:59:59 |

---

## 任务四：App Store 应用描述

| 字段 | 值 |
|------|-----|
| **TaskID** | `task_004` |
| **Title** | 健身 App 应用商店描述 |
| **Description** | 撰写健身 App 在 App Store 的描述文字，不超过 3000 字符，需涵盖核心功能介绍、用户评价摘要、隐私政策提示。 |
| **Category** | 文案撰写 |
| **Status** | `AWARDED` |
| **Requester** | `Actor{ID: "user_alice", Kind: "user", Name: "Alice"}` |
| **AssignedExecutor** | `Actor{ID: "user_bob", Kind: "user", Name: "Bob"}` |
| **Reward** | `Reward{Mode: "fixed", Amount: 350.00, Currency: "CNY"}` |
| **AcceptanceSpec** | `AcceptanceSpec{Mode: "test", Rules: ["字符数 <= 3000", "包含关键词: 健身、训练、进度"]}` |
| **Deadline** | 2026-05-05 23:59:59 |

---

## 任务五：广告标语集合

| 字段 | 值 |
|------|-----|
| **TaskID** | `task_005` |
| **Title** | 新能源汽车广告标语 |
| **Description** | 为新款新能源汽车创作 10 条广告标语，每条不超过 15 字，需体现环保、科技、豪华感，可用于不同场景（线上广告、线下展厅、宣传册）。 |
| **Category** | 文案撰写 |
| **Status** | `IN_PROGRESS` |
| **Requester** | `Actor{ID: "user_alice", Kind: "user", Name: "Alice"}` |
| **AssignedExecutor** | `Actor{ID: "user_carol", Kind: "user", Name: "Carol"}` |
| **Reward** | `Reward{Mode: "milestone", Amount: 800.00, Currency: "CNY"}` |
| **AcceptanceSpec** | `AcceptanceSpec{Mode: "hybrid"}` |
| **SettlementTerms** | `SettlementTerms{Trigger: "submission_accepted"}` |
| **Deadline** | 2026-05-15 23:59:59 |

---

## 任务六：邮件营销文案

| 字段 | 值 |
|------|-----|
| **TaskID** | `task_006` |
| **Title** | 618 大促邮件营销文案 |
| **Description** | 为 618 电商大促活动撰写一封邮件营销文案，包括标题、副标题、主 body、CTA 按钮文案，全文不超过 800 字，需引导用户点击下单。 |
| **Category** | 文案撰写 |
| **Status** | `SUBMITTED` |
| **Requester** | `Actor{ID: "user_alice", Kind: "user", Name: "Alice"}` |
| **AssignedExecutor** | `Actor{ID: "user_dave", Kind: "user", Name: "Dave"}` |
| **Reward** | `Reward{Mode: "fixed", Amount: 180.00, Currency: "CNY"}` |
| **AcceptanceSpec** | `AcceptanceSpec{Mode: "manual"}` |
| **Deadline** | 2026-04-25 23:59:59 |

---

## 任务七：视频脚本

| 字段 | 值 |
|------|-----|
| **TaskID** | `task_007` |
| **Title** | 产品功能介绍视频脚本 |
| **Description** | 撰写一个 60 秒产品功能介绍视频的脚本，包含开场白、功能点讲解、结尾号召，时长分配合理，语言生动易懂。 |
| **Category** | 文案撰写 |
| **Status** | `ACCEPTED` |
| **Requester** | `Actor{ID: "user_alice", Kind: "user", Name: "Alice"}` |
| **AssignedExecutor** | `Actor{ID: "user_eve", Kind: "user", Name: "Eve"}` |
| **Reviewer** | `Actor{ID: "user_alice", Kind: "user", Name: "Alice"}` |
| **Reward** | `Reward{Mode: "fixed", Amount: 450.00, Currency: "CNY"}` |
| **AcceptanceSpec** | `AcceptanceSpec{Mode: "manual"}` |
| **Deadline** | 2026-05-20 23:59:59 |

---

## 任务状态说明

| Status | 说明 |
|--------|------|
| `OPEN` | 任务开放，可接单 |
| `BIDDING` | 竞标中 |
| `AWARDED` | 已授标给执行方 |
| `IN_PROGRESS` | 执行中 |
| `SUBMITTED` | 已提交，待验收 |
| `ACCEPTED` | 验收通过 |
| `SETTLED` | 已结算 |
| `CANCELLED` | 已取消 |
| `DISPUTED` | 争议中 |