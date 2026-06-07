import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

// ─── Tailwind class merging ───────────────────────────────────────────────────
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// ─── Date utilities ───────────────────────────────────────────────────────────
export function today(): string {
  return new Date().toISOString().split('T')[0]
}

export function formatDate(date: Date | string): string {
  const d = typeof date === 'string' ? new Date(date) : date
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}

export function getDayName(): string {
  return new Date().toLocaleDateString('en-US', { weekday: 'long' }).toLowerCase()
}

export function getGreeting(): string {
  const hour = new Date().getHours()
  if (hour < 12) return 'good morning'
  if (hour < 17) return 'good afternoon'
  return 'good evening'
}

// ─── Schedule formatting ──────────────────────────────────────────────────────
import type { Habit } from '@/types'
import { WEEKDAY_LABELS } from '@/types'

export function formatSchedule(habit: Habit): string {
  switch (habit.scheduleType) {
    case 'daily':           return 'Every day'
    case 'every_other_day': return 'Every other day'
    case 'weekdays':
      return habit.weekdays.map(d => WEEKDAY_LABELS[d]).join(', ')
    default:                return ''
  }
}

// ─── Long press hook ──────────────────────────────────────────────────────────
import { useRef, useCallback } from 'react'

export function useLongPress(
  onLongPress: () => void,
  onClick: () => void,
  delay = 500,
) {
  const timeout = useRef<ReturnType<typeof setTimeout>>()
  const fired   = useRef(false)

  const start = useCallback(() => {
    fired.current = false
    timeout.current = setTimeout(() => {
      fired.current = true
      onLongPress()
    }, delay)
  }, [onLongPress, delay])

  const cancel = useCallback(() => {
    clearTimeout(timeout.current)
  }, [])

  const handleClick = useCallback(() => {
    if (!fired.current) onClick()
    fired.current = false
  }, [onClick])

  return {
    onMouseDown:  start,
    onMouseUp:    cancel,
    onMouseLeave: cancel,
    onTouchStart: start,
    onTouchEnd:   (e: React.TouchEvent) => { cancel(); handleClick() },
    onClick:      handleClick,
  }
}

// ─── Haptic feedback (where supported) ───────────────────────────────────────
export function haptic(style: 'light' | 'medium' | 'heavy' = 'light') {
  if ('vibrate' in navigator) {
    const patterns: Record<typeof style, number[]> = {
      light:  [5],
      medium: [10],
      heavy:  [20],
    }
    navigator.vibrate(patterns[style])
  }
}
