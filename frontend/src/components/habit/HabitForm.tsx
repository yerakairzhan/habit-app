import { useState } from 'react'
import { motion } from 'framer-motion'
import { Minus, Plus } from 'lucide-react'
import type { Habit, HabitFormData, ScheduleType, Weekday } from '@/types'
import { HABIT_COLORS } from '@/types'
import { WeekdaySelector } from './WeekdaySelector'
import { cn } from '@/lib/utils'

interface HabitFormProps {
  initial?:   Partial<Habit>
  onSubmit:   (data: HabitFormData) => void
  isLoading?: boolean
  submitLabel?: string
}

const SCHEDULE_OPTIONS: { value: ScheduleType; label: string; sub: string }[] = [
  { value: 'daily',           label: 'Every day',       sub: 'Runs daily'           },
  { value: 'every_other_day', label: 'Every other day', sub: 'Alternating days'     },
  { value: 'weekdays',        label: 'Custom days',     sub: 'You pick the days'    },
]

export function HabitForm({ initial, onSubmit, isLoading, submitLabel = 'Save habit' }: HabitFormProps) {
  const [title,    setTitle]    = useState(initial?.title        ?? '')
  const [color,    setColor]    = useState(initial?.color        ?? HABIT_COLORS[0])
  const [schedule, setSchedule] = useState<ScheduleType>(initial?.scheduleType ?? 'daily')
  const [weekdays, setWeekdays] = useState<Weekday[]>(initial?.weekdays ?? [1, 2, 3, 4, 5])
  const [target,   setTarget]   = useState(initial?.targetPerDay ?? 1)
  const [errors,   setErrors]   = useState<Record<string, string>>({})

  function validate(): boolean {
    const e: Record<string, string> = {}
    if (!title.trim())               e.title    = 'Title is required'
    if (target < 1)                  e.target   = 'Target must be at least 1'
    if (schedule === 'weekdays' && weekdays.length === 0)
                                     e.weekdays = 'Select at least one day'
    setErrors(e)
    return Object.keys(e).length === 0
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!validate()) return
    onSubmit({ title: title.trim(), color, targetPerDay: target, scheduleType: schedule, weekdays })
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-5 px-6 pb-32">

      {/* Title */}
      <div className="space-y-2">
        <label className="text-secondary font-display text-xs font-semibold uppercase tracking-widest">
          Habit name
        </label>
        <input
          value={title}
          onChange={e => { setTitle(e.target.value); setErrors(p => ({ ...p, title: '' })) }}
          placeholder="e.g. Morning run"
          className={cn('input-field', errors.title && 'border-red-500/60 focus:border-red-500/80')}
          autoFocus
          maxLength={80}
        />
        {errors.title && (
          <p className="text-red-400 text-xs font-body">{errors.title}</p>
        )}
      </div>

      {/* Color picker */}
      <div className="space-y-3">
        <label className="text-secondary font-display text-xs font-semibold uppercase tracking-widest">
          Color
        </label>
        <div className="flex flex-wrap gap-3">
          {HABIT_COLORS.map(c => (
            <motion.button
              key={c}
              type="button"
              whileTap={{ scale: 0.85 }}
              onClick={() => setColor(c)}
              className={cn(
                'w-9 h-9 rounded-xl transition-all duration-200',
                color === c && 'ring-2 ring-offset-2 ring-offset-bg scale-110',
              )}
              style={{
                backgroundColor: c,
                '--tw-ring-color': c,
              } as React.CSSProperties}
            />
          ))}
        </div>
      </div>

      {/* Schedule */}
      <div className="space-y-3">
        <label className="text-secondary font-display text-xs font-semibold uppercase tracking-widest">
          Schedule
        </label>
        <div className="space-y-2">
          {SCHEDULE_OPTIONS.map(opt => (
            <motion.button
              key={opt.value}
              type="button"
              whileTap={{ scale: 0.98 }}
              onClick={() => setSchedule(opt.value)}
              className={cn(
                'w-full text-left p-4 rounded-xl border transition-all duration-200 flex items-center justify-between',
                schedule === opt.value
                  ? 'bg-green/10 border-green/40 text-primary'
                  : 'bg-card border-border text-secondary hover:border-muted',
              )}
            >
              <div>
                <p className={cn('font-display font-semibold text-sm',
                  schedule === opt.value ? 'text-primary' : 'text-secondary'
                )}>{opt.label}</p>
                <p className="text-faint font-body text-xs mt-0.5">{opt.sub}</p>
              </div>
              <div className={cn(
                'w-4 h-4 rounded-full border-2 flex items-center justify-center transition-all',
                schedule === opt.value ? 'border-green bg-green' : 'border-faint',
              )}>
                {schedule === opt.value && (
                  <div className="w-1.5 h-1.5 rounded-full bg-bg" />
                )}
              </div>
            </motion.button>
          ))}
        </div>

        {/* Weekday selector */}
        {schedule === 'weekdays' && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            className="space-y-2"
          >
            <WeekdaySelector selected={weekdays} onChange={setWeekdays} />
            {errors.weekdays && (
              <p className="text-red-400 text-xs font-body">{errors.weekdays}</p>
            )}
          </motion.div>
        )}
      </div>

      {/* Target per day */}
      <div className="space-y-3">
        <label className="text-secondary font-display text-xs font-semibold uppercase tracking-widest">
          Target per day
        </label>
        <div className="flex items-center gap-4 bg-card border border-border rounded-xl p-4">
          <motion.button
            type="button"
            whileTap={{ scale: 0.88 }}
            onClick={() => setTarget(t => Math.max(1, t - 1))}
            className="w-10 h-10 rounded-lg bg-muted flex items-center justify-center text-secondary hover:text-primary transition-colors"
          >
            <Minus size={16} />
          </motion.button>

          <div className="flex-1 text-center">
            <span className="font-mono font-medium text-3xl text-primary">{target}</span>
            <p className="text-faint font-body text-xs mt-0.5">
              {target === 1 ? 'time' : 'times'} per day
            </p>
          </div>

          <motion.button
            type="button"
            whileTap={{ scale: 0.88 }}
            onClick={() => setTarget(t => Math.min(99, t + 1))}
            className="w-10 h-10 rounded-lg bg-green/15 border border-green/30 flex items-center justify-center text-green hover:bg-green/25 transition-colors"
          >
            <Plus size={16} />
          </motion.button>
        </div>
      </div>

      {/* Submit */}
      <motion.button
        type="submit"
        disabled={isLoading}
        whileTap={{ scale: 0.97 }}
        className="btn-primary mt-2 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {isLoading ? (
          <span className="flex items-center justify-center gap-2">
            <span className="w-4 h-4 border-2 border-bg/30 border-t-bg rounded-full animate-spin" />
            Saving…
          </span>
        ) : submitLabel}
      </motion.button>
    </form>
  )
}
