import { motion } from 'framer-motion'
import { cn } from '@/lib/utils'

// ─── Skeleton ─────────────────────────────────────────────────────────────────

export function Skeleton({ className }: { className?: string }) {
  return (
    <div className={cn(
      'skeleton bg-surface rounded-xl',
      className,
    )} />
  )
}

export function HabitCardSkeleton() {
  return (
    <div className="flex items-center gap-4 p-4 rounded-card bg-card border border-border">
      <Skeleton className="w-14 h-14 rounded-xl flex-shrink-0" />
      <div className="flex-1 space-y-2">
        <Skeleton className="h-4 w-3/4 rounded-md" />
        <Skeleton className="h-3 w-1/2 rounded-md" />
      </div>
      <Skeleton className="w-12 h-6 rounded-full" />
    </div>
  )
}

export function TodayCardSkeleton() {
  return (
    <div className="p-5 rounded-card bg-card border border-border space-y-4">
      <div className="flex items-center gap-4">
        <Skeleton className="w-12 h-12 rounded-xl flex-shrink-0" />
        <div className="flex-1 space-y-2">
          <Skeleton className="h-4 w-2/3 rounded-md" />
          <Skeleton className="h-3 w-1/3 rounded-md" />
        </div>
      </div>
      <Skeleton className="h-1.5 w-full rounded-full" />
    </div>
  )
}

export function LoadingState({ count = 3, Card = TodayCardSkeleton }: { count?: number; Card?: React.FC }) {
  return (
    <div className="space-y-3 px-6">
      {Array.from({ length: count }).map((_, i) => (
        <motion.div
          key={i}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: i * 0.07 }}
        >
          <Card />
        </motion.div>
      ))}
    </div>
  )
}

// ─── Empty State ──────────────────────────────────────────────────────────────

interface EmptyStateProps {
  icon?:    React.ReactNode
  title:    string
  message:  string
  action?:  React.ReactNode
}

export function EmptyState({ icon, title, message, action }: EmptyStateProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4 }}
      className="flex flex-col items-center justify-center py-16 px-8 text-center"
    >
      {icon && (
        <div className="w-16 h-16 rounded-2xl bg-green/10 border border-green/20
                        flex items-center justify-center mb-6 text-green">
          {icon}
        </div>
      )}
      <h3 className="font-display font-semibold text-lg text-primary mb-2">{title}</h3>
      <p className="font-body text-secondary text-sm leading-relaxed mb-6">{message}</p>
      {action}
    </motion.div>
  )
}
