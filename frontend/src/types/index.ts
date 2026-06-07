// ─── Auth ─────────────────────────────────────────────────────────────────────

export interface User {
  id: string
  email: string
  name: string
}

export interface AuthTokens {
  accessToken: string
  refreshToken: string
}

// ─── Habits ──────────────────────────────────────────────────────────────────

export type ScheduleType = 'daily' | 'every_other_day' | 'weekdays'

export type Weekday = 0 | 1 | 2 | 3 | 4 | 5 | 6  // 0=Sun … 6=Sat

export const WEEKDAY_LABELS: Record<Weekday, string> = {
  0: 'Sun', 1: 'Mon', 2: 'Tue', 3: 'Wed', 4: 'Thu', 5: 'Fri', 6: 'Sat',
}

export const WEEKDAY_LABELS_FULL: Record<Weekday, string> = {
  0: 'Sunday', 1: 'Monday', 2: 'Tuesday', 3: 'Wednesday',
  4: 'Thursday', 5: 'Friday', 6: 'Saturday',
}

export interface Habit {
  id: string
  userId: string
  title: string
  color: string
  targetPerDay: number
  scheduleType: ScheduleType
  weekdays: Weekday[]
  createdAt: string
  updatedAt: string
}

export interface TodayHabit {
  habit: Habit
  progress: number
  target: number
  completed: boolean
}

// ─── Progress ────────────────────────────────────────────────────────────────

export interface ProgressResult {
  habitId: string
  date: string
  progress: number
  target: number
  completed: boolean
}

// ─── Settings ────────────────────────────────────────────────────────────────

export interface Settings {
  language: string
  notificationsEnabled: boolean
  showDates: boolean
}

// ─── Stats ───────────────────────────────────────────────────────────────────

export interface Stats {
  currentStreak: number
  bestStreak: number
  trackedHabitsCount: number
  completionRate30d: number
}

// ─── Forms ───────────────────────────────────────────────────────────────────

export interface HabitFormData {
  title: string
  color: string
  targetPerDay: number
  scheduleType: ScheduleType
  weekdays: Weekday[]
}

export interface LoginFormData {
  email: string
  password: string
}

export interface RegisterFormData {
  name: string
  email: string
  password: string
}

// ─── API ─────────────────────────────────────────────────────────────────────

export interface ApiError {
  code: string
  message: string
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
}

// ─── Offline Queue ────────────────────────────────────────────────────────────

export type OfflineAction =
  | { type: 'INCREMENT'; habitId: string; date: string }
  | { type: 'DECREMENT'; habitId: string; date: string }
  | { type: 'RESET';     habitId: string; date: string }

// ─── UI ──────────────────────────────────────────────────────────────────────

export const HABIT_COLORS = [
  '#22C55E', // green (default)
  '#3B82F6', // blue
  '#A855F7', // purple
  '#F59E0B', // amber
  '#EF4444', // red
  '#EC4899', // pink
  '#14B8A6', // teal
  '#F97316', // orange
  '#6366F1', // indigo
  '#84CC16', // lime
] as const

export type HabitColor = typeof HABIT_COLORS[number]
