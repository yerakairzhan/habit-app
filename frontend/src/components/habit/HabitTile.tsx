import { motion } from 'framer-motion'
import { ChevronRight, Pencil, Trash2 } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import type { Habit } from '@/types'
import { useDeleteHabit } from '@/hooks/queries'
import { formatSchedule } from '@/lib/utils'
import { useState } from 'react'

interface HabitTileProps {
  habit: Habit
  index: number
}

export function HabitTile({ habit, index }: HabitTileProps) {
  const navigate   = useNavigate()
  const deleteMut  = useDeleteHabit()
  const [confirm, setConfirm] = useState(false)

  function handleDelete() {
    if (!confirm) { setConfirm(true); return }
    deleteMut.mutate(habit.id)
  }

  return (
    <motion.div
      initial={{ opacity: 0, x: -12 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ delay: index * 0.05, duration: 0.3 }}
      className="glass-card p-4 flex items-center gap-4 group"
    >
      {/* Color dot */}
      <div
        className="w-3 h-3 rounded-full flex-shrink-0 ring-2 ring-offset-2 ring-offset-card"
        style={{ backgroundColor: habit.color, '--tw-ring-color': `${habit.color}33` } as React.CSSProperties}
      />

      {/* Info */}
      <div className="flex-1 min-w-0 cursor-pointer" onClick={() => navigate(`/habits/${habit.id}/edit`)}>
        <p className="font-display font-semibold text-primary text-base truncate">
          {habit.title}
        </p>
        <p className="text-secondary font-body text-xs mt-0.5">
          {formatSchedule(habit)}
          {habit.targetPerDay > 1 && (
            <span className="text-faint ml-1">· {habit.targetPerDay}× per day</span>
          )}
        </p>
      </div>

      {/* Actions */}
      <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
        <button
          onClick={() => navigate(`/habits/${habit.id}/edit`)}
          className="w-8 h-8 rounded-lg bg-surface hover:bg-muted flex items-center justify-center
                     text-secondary hover:text-primary transition-colors"
        >
          <Pencil size={13} />
        </button>
        <button
          onClick={handleDelete}
          className={`w-8 h-8 rounded-lg flex items-center justify-center transition-all duration-200 ${
            confirm
              ? 'bg-red-500/20 text-red-400 hover:bg-red-500/30'
              : 'bg-surface hover:bg-muted text-secondary hover:text-red-400'
          }`}
          onBlur={() => setConfirm(false)}
        >
          <Trash2 size={13} />
        </button>
      </div>

      <ChevronRight
        size={16}
        className="text-faint flex-shrink-0 group-hover:text-secondary transition-colors"
      />
    </motion.div>
  )
}
