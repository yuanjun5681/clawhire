# 任务测试用例

本文档提供文案设计、图片生成等任务的测试用例。

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
| **Description** | 撰写一个 60 秒 AI Agent 产品功能介绍视频的脚本，包含开场白、功能点讲解、结尾号召，时长分配合理，语言生动易懂。需要一个 markdown 格式的脚本文件。|
| **Category** | 文案撰写 |
| **Status** | `ACCEPTED` |
| **Requester** | `Actor{ID: "user_alice", Kind: "user", Name: "Alice"}` |
| **AssignedExecutor** | `Actor{ID: "user_eve", Kind: "user", Name: "Eve"}` |
| **Reviewer** | `Actor{ID: "user_alice", Kind: "user", Name: "Alice"}` |
| **Reward** | `Reward{Mode: "fixed", Amount: 450.00, Currency: "CNY"}` |
| **AcceptanceSpec** | `AcceptanceSpec{Mode: "manual"}` |
| **Deadline** | 2026-05-20 23:59:59 |

---

## 任务八：电商商品主图

| 字段 | 值 |
|------|-----|
| **TaskID** | `task_008` |
| **Title** | 无线耳机电商主图设计 |
| **Description** | 为新款无线降噪耳机制作 5 张电商主图，尺寸 800x800 像素，白底突出产品，包含核心卖点角标，符合天猫/京东主图规范。 |
| **Category** | 图片生成 |
| **Status** | `OPEN` |
| **Requester** | `Actor{ID: "user_alice", Kind: "user", Name: "Alice"}` |
| **Reward** | `Reward{Mode: "fixed", Amount: 600.00, Currency: "CNY"}` |
| **AcceptanceSpec** | `AcceptanceSpec{Mode: "schema", Rules: ["尺寸 800x800", "白底", "JPG/PNG 格式", "5 张数量"]}` |
| **Deadline** | 2026-05-03 23:59:59 |

---

## 任务九：双十一促销 Banner

| 字段 | 值 |
|------|-----|
| **TaskID** | `task_009` |
| **Title** | 双十一大促首页 Banner 设计 |
| **Description** | 为电商平台双十一活动设计首页 Banner 一组（PC 端 1920x600、移动端 750x400），主题"狂欢盛典 全场五折"，色调热烈，需突出优惠力度与紧迫感。 |
| **Category** | 图片生成 |
| **Status** | `BIDDING` |
| **Requester** | `Actor{ID: "user_alice", Kind: "user", Name: "Alice"}` |
| **Reward** | `Reward{Mode: "bid", Amount: 0, Currency: "CNY"}` |
| **AcceptanceSpec** | `AcceptanceSpec{Mode: "manual"}` |
| **Deadline** | 2026-05-08 23:59:59 |

---

## 任务十：小红书种草配图

| 字段 | 值 |
|------|-----|
| **TaskID** | `task_010` |
| **Title** | 护肤品小红书种草九宫格 |
| **Description** | 为新品保湿精华生成小红书种草九宫格图片（9 张，1080x1350 像素），风格清新治愈，含产品特写、使用场景、文字贴纸，适合年轻女性用户群体。 |
| **Category** | 图片生成 |
| **Status** | `AWARDED` |
| **Requester** | `Actor{ID: "user_alice", Kind: "user", Name: "Alice"}` |
| **AssignedExecutor** | `Actor{ID: "user_bob", Kind: "user", Name: "Bob"}` |
| **Reward** | `Reward{Mode: "fixed", Amount: 450.00, Currency: "CNY"}` |
| **AcceptanceSpec** | `AcceptanceSpec{Mode: "schema", Rules: ["9 张图片", "尺寸 1080x1350", "包含产品/场景/文字贴纸三类"]}` |
| **Deadline** | 2026-05-06 23:59:59 |

---

## 任务十一：品牌推广海报

| 字段 | 值 |
|------|-----|
| **TaskID** | `task_011` |
| **Title** | 新品发布会推广海报 |
| **Description** | 为新能源汽车新品发布会设计推广海报一张（A3 竖版，300dpi），主题"未来已来"，需融合科技感与未来感，包含品牌 LOGO、发布会时间地点、车型剪影。 |
| **Category** | 图片生成 |
| **Status** | `IN_PROGRESS` |
| **Requester** | `Actor{ID: "user_alice", Kind: "user", Name: "Alice"}` |
| **AssignedExecutor** | `Actor{ID: "user_carol", Kind: "user", Name: "Carol"}` |
| **Reward** | `Reward{Mode: "milestone", Amount: 1200.00, Currency: "CNY"}` |
| **AcceptanceSpec** | `AcceptanceSpec{Mode: "hybrid"}` |
| **SettlementTerms** | `SettlementTerms{Trigger: "submission_accepted"}` |
| **Deadline** | 2026-05-18 23:59:59 |

---

## 任务十二：信息流广告素材

| 字段 | 值 |
|------|-----|
| **TaskID** | `task_012` |
| **Title** | 教育课程信息流广告图 |
| **Description** | 为在线英语课程制作信息流广告素材一组（1080x1080 方图 3 张、1080x1920 竖图 2 张），强调"30 天开口说"卖点，含人物使用场景与价格优惠贴纸。 |
| **Category** | 图片生成 |
| **Status** | `SUBMITTED` |
| **Requester** | `Actor{ID: "user_alice", Kind: "user", Name: "Alice"}` |
| **AssignedExecutor** | `Actor{ID: "user_dave", Kind: "user", Name: "Dave"}` |
| **Reward** | `Reward{Mode: "fixed", Amount: 700.00, Currency: "CNY"}` |
| **AcceptanceSpec** | `AcceptanceSpec{Mode: "test", Rules: ["方图 3 张 1080x1080", "竖图 2 张 1080x1920", "包含价格贴纸"]}` |
| **Deadline** | 2026-04-30 23:59:59 |

---

## 任务十三：运营活动表情包

| 字段 | 值 |
|------|-----|
| **TaskID** | `task_013` |
| **Title** | 品牌 IP 微信表情包设计 |
| **Description** | 围绕品牌吉祥物"小爪"设计微信表情包一套（16 个），含问候、加油、吐槽等常用情绪，PNG 透明底，240x240 像素，风格可爱呆萌。 |
| **Category** | 图片生成 |
| **Status** | `ACCEPTED` |
| **Requester** | `Actor{ID: "user_alice", Kind: "user", Name: "Alice"}` |
| **AssignedExecutor** | `Actor{ID: "user_eve", Kind: "user", Name: "Eve"}` |
| **Reviewer** | `Actor{ID: "user_alice", Kind: "user", Name: "Alice"}` |
| **Reward** | `Reward{Mode: "fixed", Amount: 900.00, Currency: "CNY"}` |
| **AcceptanceSpec** | `AcceptanceSpec{Mode: "schema", Rules: ["16 个表情", "PNG 透明底", "240x240 像素"]}` |
| **Deadline** | 2026-05-22 23:59:59 |

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