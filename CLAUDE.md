# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ClawHire is a task contract and settlement layer built on ClawSynapse, enabling task trading and fulfillment between humans and agents. It consists of a Go backend API and a Vue 3 frontend.

## Development Commands

### Backend (Go)
```bash
cd backend
make run      # Start API server on :8080
make build    # Compile to bin/clawhire-api
make test     # Run all tests
make lint     # go vet ./...
make tidy     # go mod tidy
```

### Frontend (Vue 3)
```bash
cd frontend
npm run dev      # Vite dev server on :5173 (proxies /api → :8080)
npm run build    # Type-check + production build
npm run preview  # Preview production build
```

### Running a Single Go Test
```bash
cd backend
go test ./internal/application/command/... -run TestFunctionName -v
```

### Local Environment
- Copy `backend/.env.example` to `backend/.env` and fill in MongoDB URI, JWT secret
- Frontend mock mode: set `VITE_USE_MOCK=true` in `frontend/.env.development`

### Test Account
- Account ID: `acct_human_alice` | Password: `password123`

## Architecture

### Backend — Hexagonal / DDD (Go, Gin, MongoDB)

Entry point: `backend/cmd/clawhire-api/main.go` — wires all dependencies and starts the Gin HTTP server.

```
internal/
├── transport/http/     # Gin handlers, middleware, DTOs, router registration
├── application/        # Use-case layer
│   ├── command/        # State-changing operations (PostTask, PlaceBid, AwardTask, …)
│   ├── query/          # Read-only handlers
│   ├── auth/           # Login/register, JWT issuance
│   └── webhook/        # ClawSynapse event dispatcher
├── domain/             # Core business logic — no external dependencies
│   ├── task/           # Task aggregate + state_machine.go (strict lifecycle rules)
│   ├── bid/, contract/, submission/, review/, settlement/
│   └── shared/         # Actor, Money, shared value objects
├── infrastructure/
│   ├── mongo/repository/  # MongoDB implementations of domain repository interfaces
│   ├── auth/           # JWT issuer
│   └── config/         # Env-based config via struct tags
└── protocol/
    ├── clawhire/        # ClawHire message types
    └── clawsynapse/     # Webhook envelope definitions
```

**Key patterns:**
- `domain/task/state_machine.go` enforces all task state transitions (OPEN → BIDDING → AWARDED → IN_PROGRESS → SUBMITTED → ACCEPTED → SETTLED). Never bypass it.
- `application/command/` and `application/query/` are strictly separated (CQRS).
- Webhook events from ClawSynapse arrive at `POST /webhooks/clawsynapse` and are dispatched by `application/webhook/`.
- Raw webhook payloads are stored in `raw_events`; domain events in `domain_events` for audit/replay.
- MongoDB repositories use abstract interfaces — pass mock implementations in tests.

### Frontend — Vue 3 + TypeScript + Pinia

Entry point: `frontend/src/main.ts` — mounts Vue app with Pinia and Vue Router.

```
src/
├── api/         # Axios client (http.ts), per-resource modules, mock/ implementations
├── stores/      # Pinia stores — identity.ts manages JWT session + localStorage
├── router/      # Vue Router with beforeEach auth guards
├── pages/       # Full-page components (Marketplace, TaskDetail, MyTasks, Account, …)
├── components/  # Reusable UI (TaskCard, Timeline, StatusBadge, Pagination, …)
└── types/       # Shared TypeScript interfaces (task, bid, submission, account, …)
```

**Key patterns:**
- `stores/identity.ts` is the single source of truth for the authenticated user. Auth guards in `router/index.ts` redirect unauthenticated users to `/login`.
- The API layer in `api/http.ts` auto-injects the JWT Bearer token and emits a custom `unauthorized` event on 401 (picked up in `main.ts`) to avoid circular store imports.
- `api/normalizers.ts` transforms raw API responses to frontend types — normalise here, not in components.
- Pages use `<script setup>` (Composition API) with explicit loading/error states; skeleton loaders for perceived performance.
- UI components use **DaisyUI 5** component classes on top of **Tailwind CSS 4**.

### HTTP Routes Summary
- `GET /healthz`, `/readyz` — health checks (no auth)
- `POST /api/auth/register`, `POST /api/auth/login` — public
- Authenticated write: `/api/tasks`, `/api/tasks/:taskId/bids`, `/api/tasks/:taskId/award`, `/api/tasks/:taskId/submissions`, `/api/tasks/:taskId/submissions/:submissionId/accept`
- Authenticated read: `/api/tasks`, `/api/tasks/:taskId`, `/api/accounts`, `/api/accounts/:accountId`
- Webhook: `POST /webhooks/clawsynapse`
