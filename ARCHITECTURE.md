# Habit Tracker – Architecture Decision Record

## Layer Map

```
proto/habit/v1/          ← gRPC contracts (source of truth)
internal/
  domain/                ← Pure Go entities + business logic (no deps)
  repository/            ← Interfaces + PostgreSQL implementations
  service/               ← Use-case orchestration
  handler/               ← gRPC handlers (thin adapters)
  middleware/            ← JWT interceptors, logging, recovery
  auth/                  ← Token issuing, validation
  config/                ← Env-driven config
pkg/
  logger/                ← Structured zerolog wrapper
  validate/              ← Proto input validation helpers
cmd/server/              ← Wire-up and main()
migrations/              ← golang-migrate SQL files
```

## Key Decisions

### gRPC-only
All transport is gRPC. The PWA frontend uses grpc-web or a separate Envoy sidecar.
No REST layer — keeps surface area minimal and typed end-to-end.

### JWT Strategy
- Access token: 15-minute HS256 JWT, claims carry user_id + email.
- Refresh token: 7-day opaque UUID stored in `refresh_tokens` table (enables revocation).
- Token rotation: every Refresh call invalidates the old token and issues a new pair.

### Streak Calculation
Streaks are computed on-read (not stored) from the `habit_logs` table.
This avoids background jobs and stale counters. For scale, a materialized cache
column can be added later with a single migration.

### Schedule Logic
- `daily`          → every calendar day
- `every_other_day`→ day diff from `created_at` is even (0, 2, 4 …)
- `weekdays`       → habit.weekdays bitmask matched against today's weekday (0=Sun … 6=Sat)

### Progress / Completion
Each (habit_id, date) pair has exactly one `habit_logs` row (UNIQUE constraint).
`progress` is incremented/decremented atomically via `UPDATE … SET progress = GREATEST(0, progress ± 1)`.
`completed = progress >= target_per_day` is derived, never stored.

### Clean Architecture Dependency Rule
domain → (nothing)
repository interfaces → domain
service → repository interfaces + domain
handler → service + proto types
middleware → auth + config
main → wires everything

No layer imports a layer above it.
