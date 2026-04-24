# AGENTS.md

## 测试账号
- Account ID: `acct_human_alice` | Password: `password123`

## 环境要求
- **MongoDB 必须运行**才能启动后端（`MONGODB_URI` 在 `backend/.env` 中配置）
- 复制 `backend/.env.example` → `backend/.env` 后填写 MongoDB URI 和 JWT secret

## 启动命令
```bash
# 后端（:8080）
cd backend && make run

# 前端（:5173，/api 代理到 :8080）
cd frontend && npm run dev
```

## 前端 mock 模式
`frontend/.env.development` 中 `VITE_USE_MOCK=false` 可切换。设为 `true` 时跳过真实 API 调用。

## 构建
```bash
cd frontend && npm run build   # vue-tsc -b 类型检查 + vite build
```

## 单测（Go）
```bash
cd backend && go test ./... -run TestFunctionName -v
```

## 重要约束
- **永远不要绕过** `domain/task/state_machine.go` 中的任务状态转换逻辑
- 任务生命周期：OPEN → BIDDING → AWARDED → IN_PROGRESS → SUBMITTED → ACCEPTED → SETTLED