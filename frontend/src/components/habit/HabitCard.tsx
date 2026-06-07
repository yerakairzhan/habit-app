import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Check, Minus, Plus, RotateCcw } from 'lucide-react'
import type { TodayHabit } from '@/types'
import { useIncrementProgress, useDecrementProgress, useResetProgress } from '@/hooks/queries'
import { cn, haptic } from '@/lib/utils'

interface HabitCardProps {
  item: TodayHabit
}

export function HabitCard({ item }: HabitCardProps) {
  const { habit, progress, target, completed } = item
  const [showActions, setShowActions] = useState(false)

  const increment = useIncrementProgress()
  const decrement = useDecrementProgress()
  const reset     = useResetProgress()

  const pct = Math.min(100, (progress / target) * 100)

  function handleTap() {
    if (completed) {
      setShowActions(v => !v)
      return
    }
    haptic('light')
    increment.mutate({ habitId: habit.id })
  }

  function handleLongPress() {
    haptic('medium')
    setShowActions(v => !v)
  }

  function handleDecrement(e: React.MouseEvent) {
    e.stopPropagation()
    haptic('light')
    decrement.mutate({ habitId: habit.id })
  }

  function handleIncrement(e: React.MouseEvent) {
    e.stopPropagation()
    haptic('light')
    increment.mutate({ habitId: habit.id })
  }

  function handleReset(e: React.MouseEvent) {
    e.stopPropagation()
    haptic('medium')
    reset.mutate({ habitId: habit.id })
    setShowActions(false)
  }

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.96 }}
      transition={{ type: 'spring', stiffness: 400, damping: 30 }}
      className="relative"
    >
      <motion.div
        className={cn(
          'glass-card overflow-hidden cursor-pointer select-none',
          'transition-all duration-200',
          completed && 'border-green/30 shadow-green',
        )}
        whileTap={{ scale: 0.98 }}
        onClick={handleTap}
        onContextMenu={(e) => { e.preventDefault(); handleLongPress() }}
        onTouchStart={() => {
          const t = setTimeout(handleLongPress, 500)
          const cleanup = () => clearTimeout(t)
          window.addEventListener('touchend', cleanup, { once: true })
          window.addEventListener('touchmove', cleanup, { once: true })
        }}
      >
        {/* Green completion overlay */}
        {completed && (
          <motion.div
            className="absolute inset-0 bg-green-gradient opacity-40 pointer-events-none"
            initial={{ opacity: 0 }}
            animate={{ opacity: 0.4 }}
          />
        )}

        <div className="p-4 flex items-center gap-4">
          {/* Color square with check */}
          <div className="relative flex-shrink-0">
            <div
              className="w-14 h-14 rounded-xl flex items-center justify-center shadow-lg"
              style={{ backgroundColor: `${habit.color}22`, border: `1.5px solid ${habit.color}44` }}
            >
              <AnimatePresence mode="wait">
                {completed ? (
                  <motion.div
                    key="check"
                    initial={{ scale: 0, rotate: -10 }}
                    animate={{ scale: 1, rotate: 0 }}
                    exit={{ scale: 0 }}
                    transition={{ type: 'spring', stiffness: 500, damping: 25 }}
                  >
                    <Check
                      size={24}
                      strokeWidth={3}
                      style={{ color: habit.color }}
                    />
                  </motion.div>
                ) : (
                  <motion.div
                    key="dot"
                    initial={{ scale: 0 }}
                    animate={{ scale: 1 }}
                    exit={{ scale: 0 }}
                    className="w-3 h-3 rounded-full"
                    style={{ backgroundColor: habit.color }}
                  />
                )}
              </AnimatePresence>
            </div>
          </div>

          {/* Content */}
          <div className="flex-1 min-w-0">
            <h3 className={cn(
              'font-display font-semibold text-base truncate transition-colors',
              completed ? 'text-green' : 'text-primary',
            )}>
              {habit.title}
            </h3>
            <p className="text-secondary font-mono text-sm mt-0.5">
              {progress}
              <span className="text-faint">/{target}</span>
              {completed && (
                <span className="text-green font-body text-xs ml-2">completed</span>
              )}
            </p>
          </div>

          {/* Progress ring / counter */}
          <ProgressRing progress={pct} color={habit.color} size={40} />
        </div>

        {/* Progress bar */}
        <div className="h-0.5 bg-muted/50 mx-4 mb-4 rounded-full overflow-hidden">
          <motion.div
            className="h-full rounded-full"
            style={{ backgroundColor: habit.color }}
            initial={{ width: 0 }}
            animate={{ width: `${pct}%` }}
            transition={{ type: 'spring', stiffness: 200, damping: 30 }}
          />
        </div>
      </motion.div>

      {/* Inline action row */}
      <AnimatePresence>
        {showActions && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="overflow-hidden"
          >
            <div className="flex items-center gap-2 px-2 pt-2 pb-1">
              <ActionButton icon={<Minus size={14} />} onClick={handleDecrement} label="Less" />
              <ActionButton icon={<Plus size={14} />}  onClick={handleIncrement} label="More" accent />
              <ActionButton icon={<RotateCcw size={14} />} onClick={handleReset} label="Reset" />
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  )
}

// ─── Progress Ring ────────────────────────────────────────────────────────────

function ProgressRing({ progress, color, size }: { progress: number; color: string; size: number }) {
  const r     = (size - 6) / 2
  const circ  = 2 * Math.PI * r
  const dash  = (progress / 100) * circ

  return (
    <svg width={size} height={size} className="flex-shrink-0 -rotate-90">
      <circle cx={size/2} cy={size/2} r={r} stroke="#1C2133" strokeWidth={3} fill="none" />
      <motion.circle
        cx={size/2} cy={size/2} r={r}
        stroke={color}
        strokeWidth={3}
        fill="none"
        strokeLinecap="round"
        strokeDasharray={circ}
        initial={{ strokeDashoffset: circ }}
        animate={{ strokeDashoffset: circ - dash }}
        transition={{ type: 'spring', stiffness: 150, damping: 25 }}
      />
    </svg>
  )
}

function ActionButton({
  icon, onClick, label, accent = false,
}: {
  icon: React.ReactNode; onClick: (e: React.MouseEvent) => void; label: string; accent?: boolean
}) {
  return (
    <motion.button
      whileTap={{ scale: 0.92 }}
      onClick={onClick}
      className={cn(
        'flex-1 flex items-center justify-center gap-1.5 py-2.5 rounded-xl text-xs font-display font-medium',
        'border transition-colors duration-150',
        accent
          ? 'bg-green/10 border-green/30 text-green hover:bg-green/20'
          : 'bg-card border-border text-secondary hover:text-primary',
      )}
    >
      {icon}
      {label}
    </motion.button>
  )
}
