/**
 * Connect-protocol API client.
 *
 * The Connect protocol (https://connectrpc.com) uses:
 *   POST /{package}.{Service}/{Method}
 *   Content-Type: application/json
 *   Connect-Protocol-Version: 1
 *
 * Proto-JSON encoding rules (what the server sends and expects):
 *   - Field names are snake_case  (matches proto field names)
 *   - Enums are string names      e.g. "SCHEDULE_TYPE_DAILY"
 *   - Timestamps are RFC3339      e.g. "2024-01-15T10:30:00Z"
 *   - Missing fields are omitted  (not null, not zero)
 *
 * This client talks directly to the Go Connect server — no Envoy needed.
 * The server runs net/http + h2c, so HTTP/1.1 and HTTP/2 both work.
 */

import { useAuthStore } from '@/store/auth'
import type { Habit, HabitFormData, ProgressResult, Settings, Stats } from '@/types'

const BASE_URL = (import.meta.env.VITE_API_URL as string | undefined) ?? ''

// ─── Connect error shape ─────────────────────────────────────────────────────

export class ApiError extends Error {
  constructor(
    public readonly code: string,
    message: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

// ─── Core RPC helper ─────────────────────────────────────────────────────────

async function rpc<TReq, TRes>(
  service: string,
  method: string,
  body: TReq,
  requiresAuth = true,
): Promise<TRes> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'Connect-Protocol-Version': '1',
  }

  if (requiresAuth) {
    const token = useAuthStore.getState().accessToken
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }
  }

  const url = `${BASE_URL}/habit.v1.${service}/${method}`

  let res: Response
  try {
    res = await fetch(url, {
      method: 'POST',
      headers,
      body: JSON.stringify(body),
    })
  } catch (e) {
    // Network error (offline, DNS failure, etc.)
    throw new ApiError('unavailable', 'Network error — check your connection')
  }

  // Connect errors: { "code": "not_found", "message": "..." }
  if (!res.ok) {
    let errBody: { code?: string; message?: string } = {}
    try { errBody = await res.json() } catch { /* ignore parse error */ }
    throw new ApiError(
      errBody.code ?? connectCodeFromStatus(res.status),
      errBody.message ?? `HTTP ${res.status}`,
    )
  }

  const data = await res.json() as TRes
  return data
}

function connectCodeFromStatus(status: number): string {
  if (status === 401) return 'unauthenticated'
  if (status === 403) return 'permission_denied'
  if (status === 404) return 'not_found'
  if (status === 409) return 'already_exists'
  if (status === 400) return 'invalid_argument'
  return 'internal'
}

// ─── Proto-JSON response shapes (snake_case) ─────────────────────────────────
// These match exactly what the Go server emits via proto-JSON encoding.

interface ProtoUser {
  id: string
  email: string
  name: string
}

interface ProtoHabit {
  id: string
  user_id: string
  title: string
  color: string
  target_per_day: number
  schedule_type: string   // e.g. "SCHEDULE_TYPE_DAILY"
  weekdays: string[]      // e.g. ["WEEKDAY_MONDAY", "WEEKDAY_WEDNESDAY"]
  created_at: string      // RFC3339
  updated_at: string      // RFC3339
}

interface ProtoTodayHabit {
  habit: ProtoHabit
  progress: number
  target: number
  completed: boolean
}

interface ProtoProgressResponse {
  habit_id: string
  date: string
  progress: number
  target: number
  completed: boolean
}

interface ProtoSettings {
  language: string
  notifications_enabled: boolean
  show_dates: boolean
}

interface ProtoStats {
  current_streak: number
  best_streak: number
  tracked_habits_count: number
  completion_rate_30d: number
}

// ─── Proto enum → domain type maps ───────────────────────────────────────────

const SCHEDULE_MAP: Record<string, Habit['scheduleType']> = {
  SCHEDULE_TYPE_DAILY:           'daily',
  SCHEDULE_TYPE_EVERY_OTHER_DAY: 'every_other_day',
  SCHEDULE_TYPE_WEEKDAYS:        'weekdays',
}

// Proto weekday enum → domain Weekday (0=Sun…6=Sat)
const WEEKDAY_MAP: Record<string, number> = {
  WEEKDAY_SUNDAY:    0,
  WEEKDAY_MONDAY:    1,
  WEEKDAY_TUESDAY:   2,
  WEEKDAY_WEDNESDAY: 3,
  WEEKDAY_THURSDAY:  4,
  WEEKDAY_FRIDAY:    5,
  WEEKDAY_SATURDAY:  6,
}

// Domain ScheduleType → proto enum string
const SCHEDULE_TO_PROTO: Record<Habit['scheduleType'], string> = {
  daily:           'SCHEDULE_TYPE_DAILY',
  every_other_day: 'SCHEDULE_TYPE_EVERY_OTHER_DAY',
  weekdays:        'SCHEDULE_TYPE_WEEKDAYS',
}

// Domain Weekday (0-6) → proto enum string
const WEEKDAY_TO_PROTO: Record<number, string> = {
  0: 'WEEKDAY_SUNDAY',
  1: 'WEEKDAY_MONDAY',
  2: 'WEEKDAY_TUESDAY',
  3: 'WEEKDAY_WEDNESDAY',
  4: 'WEEKDAY_THURSDAY',
  5: 'WEEKDAY_FRIDAY',
  6: 'WEEKDAY_SATURDAY',
}

// ─── Response mappers ─────────────────────────────────────────────────────────

function mapHabit(p: ProtoHabit): Habit {
  return {
    id:           p.id,
    userId:       p.user_id,
    title:        p.title,
    color:        p.color,
    targetPerDay: p.target_per_day,
    scheduleType: SCHEDULE_MAP[p.schedule_type] ?? 'daily',
    weekdays:     (p.weekdays ?? []).map(w => WEEKDAY_MAP[w] ?? 0) as Habit['weekdays'],
    createdAt:    p.created_at,
    updatedAt:    p.updated_at,
  }
}

function mapProgress(p: ProtoProgressResponse): ProgressResult {
  return {
    habitId:   p.habit_id,
    date:      p.date,
    progress:  p.progress,
    target:    p.target,
    completed: p.completed,
  }
}

function mapSettings(p: ProtoSettings): Settings {
  return {
    language:             p.language,
    notificationsEnabled: p.notifications_enabled,
    showDates:            p.show_dates,
  }
}

function mapStats(p: ProtoStats): Stats {
  return {
    currentStreak:      p.current_streak,
    bestStreak:         p.best_streak,
    trackedHabitsCount: p.tracked_habits_count,
    completionRate30d:  p.completion_rate_30d,
  }
}

// ─── Auth API ─────────────────────────────────────────────────────────────────

interface AuthResponse {
  accessToken: string
  refreshToken: string
  user: ProtoUser
}

export const authApi = {
  register: async (email: string, password: string, name: string) => {
    const res = await rpc<
      { email: string; password: string; name: string },
      AuthResponse
    >('AuthService', 'Register', { email, password, name }, false)
    return {
      accessToken:  res.accessToken,
      refreshToken: res.refreshToken,
      user: { id: res.user.id, email: res.user.email, name: res.user.name },
    }
  },

  login: async (email: string, password: string) => {
    const res = await rpc<
      { email: string; password: string },
      AuthResponse
    >('AuthService', 'Login', { email, password }, false)
    return {
      accessToken:  res.accessToken,
      refreshToken: res.refreshToken,
      user: { id: res.user.id, email: res.user.email, name: res.user.name },
    }
  },

  refreshToken: async (refreshToken: string) => {
    const res = await rpc<
      { refresh_token: string },
      { accessToken: string; refreshToken: string }
    >('AuthService', 'RefreshToken', { refresh_token: refreshToken }, false)
    return { accessToken: res.accessToken, refreshToken: res.refreshToken }
  },

  logout: async (refreshToken: string) => {
    await rpc<{ refresh_token: string }, Record<string, never>>(
      'AuthService', 'Logout', { refresh_token: refreshToken },
    )
  },
}

// ─── Habit API ────────────────────────────────────────────────────────────────

export const habitApi = {
  list: async (page = 1, pageSize = 50) => {
    const res = await rpc<
      { page: number; page_size: number },
      { habits: ProtoHabit[]; total: number }
    >('HabitService', 'ListHabits', { page, page_size: pageSize })
    return {
      habits: (res.habits ?? []).map(mapHabit),
      total:  res.total ?? 0,
    }
  },

  getToday: async () => {
    const res = await rpc<
      Record<string, never>,
      { habits: ProtoTodayHabit[] }
    >('HabitService', 'GetTodaysHabits', {})
    return (res.habits ?? []).map(h => ({
      habit:     mapHabit(h.habit),
      progress:  h.progress  ?? 0,
      target:    h.target    ?? 1,
      completed: h.completed ?? false,
    }))
  },

  get: async (id: string) => {
    const res = await rpc<{ id: string }, { habit: ProtoHabit }>(
      'HabitService', 'GetHabit', { id },
    )
    return mapHabit(res.habit)
  },

  create: async (data: HabitFormData) => {
    const res = await rpc<Record<string, unknown>, { habit: ProtoHabit }>(
      'HabitService', 'CreateHabit', {
        title:          data.title,
        color:          data.color,
        target_per_day: data.targetPerDay,
        schedule_type:  SCHEDULE_TO_PROTO[data.scheduleType],
        weekdays:       data.weekdays.map(w => WEEKDAY_TO_PROTO[w]),
      },
    )
    return mapHabit(res.habit)
  },

  update: async (id: string, data: HabitFormData) => {
    const res = await rpc<Record<string, unknown>, { habit: ProtoHabit }>(
      'HabitService', 'UpdateHabit', {
        id,
        title:          data.title,
        color:          data.color,
        target_per_day: data.targetPerDay,
        schedule_type:  SCHEDULE_TO_PROTO[data.scheduleType],
        weekdays:       data.weekdays.map(w => WEEKDAY_TO_PROTO[w]),
      },
    )
    return mapHabit(res.habit)
  },

  delete: async (id: string) => {
    await rpc<{ id: string }, Record<string, never>>(
      'HabitService', 'DeleteHabit', { id },
    )
  },
}

// ─── Progress API ─────────────────────────────────────────────────────────────

export const progressApi = {
  increment: async (habitId: string, date?: string) => {
    const res = await rpc<{ habit_id: string; date: string }, ProtoProgressResponse>(
      'ProgressService', 'IncrementProgress', { habit_id: habitId, date: date ?? '' },
    )
    return mapProgress(res)
  },

  decrement: async (habitId: string, date?: string) => {
    const res = await rpc<{ habit_id: string; date: string }, ProtoProgressResponse>(
      'ProgressService', 'DecrementProgress', { habit_id: habitId, date: date ?? '' },
    )
    return mapProgress(res)
  },

  reset: async (habitId: string, date?: string) => {
    const res = await rpc<{ habit_id: string; date: string }, ProtoProgressResponse>(
      'ProgressService', 'ResetProgress', { habit_id: habitId, date: date ?? '' },
    )
    return mapProgress(res)
  },
}

// ─── Settings API ─────────────────────────────────────────────────────────────

export const settingsApi = {
  get: async (): Promise<Settings> => {
    const res = await rpc<Record<string, never>, { settings: ProtoSettings }>(
      'SettingsService', 'GetSettings', {},
    )
    return mapSettings(res.settings)
  },

  // Fix #8: send top-level snake_case fields, not a nested { settings: ... } wrapper
  update: async (s: Settings): Promise<Settings> => {
    const res = await rpc<Record<string, unknown>, { settings: ProtoSettings }>(
      'SettingsService', 'UpdateSettings', {
        language:              s.language,
        notifications_enabled: s.notificationsEnabled,
        show_dates:            s.showDates,
      },
    )
    return mapSettings(res.settings)
  },
}

// ─── Stats API ────────────────────────────────────────────────────────────────

export const statsApi = {
  get: async (habitId?: string): Promise<Stats> => {
    const res = await rpc<{ habit_id: string }, ProtoStats>(
      'StatsService', 'GetStats', { habit_id: habitId ?? '' },
    )
    return mapStats(res)
  },
}
