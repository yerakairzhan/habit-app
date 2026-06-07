import { motion } from 'framer-motion'
import type { Weekday } from '@/types'
import { WEEKDAY_LABELS } from '@/types'
import { cn } from '@/lib/utils'

interface WeekdaySelectorProps {
  selected:  Weekday[]
  onChange:  (days: Weekday[]) => void
}

const ALL_DAYS: Weekday[] = [1, 2, 3, 4, 5, 6, 0]  // Mon → Sun order

export function WeekdaySelector({ selected, onChange }: WeekdaySelectorProps) {
  function toggle(day: Weekday) {
    if (selected.includes(day)) {
      onChange(selected.filter(d => d !== day))
    } else {
      onChange([...selected, day].sort((a, b) => a - b))
    }
  }

  return (
    <div className="flex gap-2 justify-between">
      {ALL_DAYS.map(day => {
        const active = selected.includes(day)
        return (
          <motion.button
            key={day}
            type="button"
            whileTap={{ scale: 0.88 }}
            onClick={() => toggle(day)}
            className={cn(
              'flex-1 h-11 rounded-xl font-display font-semibold text-xs',
              'border transition-all duration-200',
              active
                ? 'bg-green/15 border-green/50 text-green shadow-glow'
                : 'bg-card border-border text-faint hover:border-muted hover:text-secondary',
            )}
          >
            {WEEKDAY_LABELS[day]}
          </motion.button>
        )
      })}
    </div>
  )
}
