import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { habitApi, progressApi, settingsApi, statsApi } from '@/api/client'
import type { HabitFormData, Settings, TodayHabit } from '@/types'
import { useOfflineStore } from '@/store/offline'
import { today } from '@/lib/date'

// ─── Query keys ──────────────────────────────────────────────────────────────

export const QK = {
  habits:      ['habits']              as const,
  todayHabits: ['today-habits']        as const,
  habit:       (id: string) => ['habits', id] as const,
  stats:       (id?: string) => ['stats', id ?? 'global'] as const,
  settings:    ['settings']            as const,
}

// ─── Habits ──────────────────────────────────────────────────────────────────

export function useHabits() {
  return useQuery({
    queryKey: QK.habits,
    queryFn:  () => habitApi.list(),
    staleTime: 30_000,
  })
}

export function useTodayHabits() {
  return useQuery({
    queryKey:        QK.todayHabits,
    queryFn:         () => habitApi.getToday(),
    staleTime:       10_000,
    refetchInterval: 60_000,
  })
}

export function useHabit(id: string) {
  return useQuery({
    queryKey: QK.habit(id),
    queryFn:  () => habitApi.get(id),
    staleTime: 30_000,
    enabled:   !!id,
  })
}

// ─── Habit mutations ──────────────────────────────────────────────────────────

export function useCreateHabit() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: HabitFormData) => habitApi.create(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: QK.habits })
      qc.invalidateQueries({ queryKey: QK.todayHabits })
    },
  })
}

export function useUpdateHabit() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: HabitFormData }) =>
      habitApi.update(id, data),
    onSuccess: (habit) => {
      qc.invalidateQueries({ queryKey: QK.habits })
      qc.invalidateQueries({ queryKey: QK.todayHabits })
      qc.setQueryData(QK.habit(habit.id), habit)
    },
  })
}

export function useDeleteHabit() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => habitApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: QK.habits })
      qc.invalidateQueries({ queryKey: QK.todayHabits })
    },
  })
}

// ─── Progress with optimistic updates ────────────────────────────────────────

export function useIncrementProgress() {
  const qc      = useQueryClient()
  const enqueue = useOfflineStore(s => s.enqueue)

  return useMutation({
    mutationFn: ({ habitId, date }: { habitId: string; date?: string }) =>
      progressApi.increment(habitId, date),

    onMutate: async ({ habitId }) => {
      await qc.cancelQueries({ queryKey: QK.todayHabits })
      const previous = qc.getQueryData<TodayHabit[]>(QK.todayHabits)

      qc.setQueryData<TodayHabit[]>(QK.todayHabits, (old = []) =>
        old.map(h =>
          h.habit.id === habitId
            ? { ...h, progress: h.progress + 1, completed: h.progress + 1 >= h.target }
            : h,
        ),
      )
      return { previous }
    },

    onError: (_err, { habitId }, ctx) => {
      if (ctx?.previous) qc.setQueryData(QK.todayHabits, ctx.previous)
      if (!navigator.onLine) {
        enqueue({ type: 'INCREMENT', habitId, date: today() })
      }
    },

    onSettled: () => qc.invalidateQueries({ queryKey: QK.todayHabits }),
  })
}

export function useDecrementProgress() {
  const qc = useQueryClient()

  return useMutation({
    mutationFn: ({ habitId, date }: { habitId: string; date?: string }) =>
      progressApi.decrement(habitId, date),

    onMutate: async ({ habitId }) => {
      await qc.cancelQueries({ queryKey: QK.todayHabits })
      const previous = qc.getQueryData<TodayHabit[]>(QK.todayHabits)

      qc.setQueryData<TodayHabit[]>(QK.todayHabits, (old = []) =>
        old.map(h =>
          h.habit.id === habitId
            ? { ...h, progress: Math.max(0, h.progress - 1), completed: Math.max(0, h.progress - 1) >= h.target }
            : h,
        ),
      )
      return { previous }
    },

    onError: (_err, _vars, ctx) => {
      if (ctx?.previous) qc.setQueryData(QK.todayHabits, ctx.previous)
    },

    onSettled: () => qc.invalidateQueries({ queryKey: QK.todayHabits }),
  })
}

export function useResetProgress() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ habitId, date }: { habitId: string; date?: string }) =>
      progressApi.reset(habitId, date),
    onSuccess: () => qc.invalidateQueries({ queryKey: QK.todayHabits }),
  })
}

// ─── Stats ────────────────────────────────────────────────────────────────────

export function useStats(habitId?: string) {
  return useQuery({
    queryKey: QK.stats(habitId),
    queryFn:  () => statsApi.get(habitId),
    staleTime: 60_000,
  })
}

// ─── Settings ─────────────────────────────────────────────────────────────────

export function useSettings() {
  return useQuery({
    queryKey: QK.settings,
    queryFn:  () => settingsApi.get(),   // returns Settings directly now
    staleTime: 60_000,
  })
}

export function useUpdateSettings() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (s: Settings) => settingsApi.update(s),
    onSuccess: (updated) => {
      // settingsApi.update returns Settings directly
      qc.setQueryData(QK.settings, updated)
    },
  })
}
