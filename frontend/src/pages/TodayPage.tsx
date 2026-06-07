import { motion, AnimatePresence } from 'framer-motion'
import { CheckSquare } from 'lucide-react'
import { useTodayHabits } from '@/hooks/queries'
import { useAuthStore } from '@/store/auth'
import { useOfflineStore } from '@/store/offline'
import { HabitCard } from '@/components/habit/HabitCard'
import { LoadingState, EmptyState } from '@/components/common/LoadingState'
import { getGreeting, getDayName } from '@/lib/utils'
import { Link } from 'react-router-dom'
import { Plus, WifiOff } from 'lucide-react'

const QUOTES = [
  'small steps, big results.',
  'consistency is the key.',
  'progress over perfection.',
  'show up every day.',
  'one habit at a time.',
]

function getDailyQuote() {
  const day = new Date().getDay()
  return QUOTES[day % QUOTES.length]
}

export function TodayPage() {
  const user       = useAuthStore(s => s.user)
  const queueLen   = useOfflineStore(s => s.queue.length)
  const isSyncing  = useOfflineStore(s => s.isSyncing)
  const { data: habits, isLoading, isError } = useTodayHabits()

  const completed  = habits?.filter(h => h.completed).length ?? 0
  const total      = habits?.length ?? 0
  const allDone    = total > 0 && completed === total

  return (
    <div className="page-container pb-28">
      {/* Header */}
      <motion.header
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className="px-6 pt-14 pb-2"
      >
        {/* Offline banner */}
        <AnimatePresence>
          {(queueLen > 0 || isSyncing) && (
            <motion.div
              initial={{ opacity: 0, y: -8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              className="flex items-center gap-2 bg-amber-500/10 border border-amber-500/20
                         rounded-xl px-3 py-2 mb-4 text-amber-400 text-xs font-body"
            >
              <WifiOff size={12} />
              {isSyncing
                ? 'Syncing offline progress…'
                : `${queueLen} action${queueLen > 1 ? 's' : ''} queued — will sync when online`
              }
            </motion.div>
          )}
        </AnimatePresence>

        {/* Greeting */}
        <p className="text-secondary font-body text-sm mb-1">
          {getGreeting()}, <span className="text-primary">{user?.name?.split(' ')[0] ?? 'friend'}</span>
        </p>
        <h1 className="font-display font-bold text-5xl text-primary leading-none tracking-tight">
          {getDayName()}
        </h1>
        <p className="text-faint font-body text-sm mt-2 italic">
          "{getDailyQuote()}"
        </p>

        {/* Progress summary */}
        {!isLoading && total > 0 && (
          <motion.div
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2 }}
            className="mt-5 flex items-center gap-3"
          >
            <div className="flex-1 h-1.5 bg-muted rounded-full overflow-hidden">
              <motion.div
                className="h-full bg-green rounded-full"
                initial={{ width: 0 }}
                animate={{ width: `${(completed / total) * 100}%` }}
                transition={{ delay: 0.3, type: 'spring', stiffness: 150, damping: 25 }}
              />
            </div>
            <span className={`font-mono text-xs font-medium ${allDone ? 'text-green' : 'text-secondary'}`}>
              {completed}/{total}
            </span>
          </motion.div>
        )}

        {/* All done banner */}
        <AnimatePresence>
          {allDone && (
            <motion.div
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.95 }}
              className="mt-4 p-3 rounded-xl bg-green/10 border border-green/20
                         flex items-center gap-2 text-green text-sm font-display font-semibold"
            >
              <CheckSquare size={16} />
              All habits done for today! 🎉
            </motion.div>
          )}
        </AnimatePresence>
      </motion.header>

      {/* Content */}
      <div className="flex-1 mt-4">
        {isLoading && <LoadingState count={3} />}

        {isError && (
          <div className="px-6">
            <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-sm font-body">
              Failed to load habits. Pull to refresh.
            </div>
          </div>
        )}

        {!isLoading && !isError && habits?.length === 0 && (
          <EmptyState
            icon={<CheckSquare size={28} />}
            title="No habits today"
            message="You have no habits scheduled for today. Create one to get started."
            action={
              <Link to="/habits/new" className="btn-primary inline-flex items-center gap-2 px-6 py-3 w-auto rounded-xl">
                <Plus size={16} />
                Add habit
              </Link>
            }
          />
        )}

        {!isLoading && habits && habits.length > 0 && (
          <motion.div
            className="space-y-3 px-6"
            initial="hidden"
            animate="show"
            variants={{
              hidden: {},
              show: { transition: { staggerChildren: 0.07 } },
            }}
          >
            <AnimatePresence mode="popLayout">
              {habits.map(item => (
                <HabitCard key={item.habit.id} item={item} />
              ))}
            </AnimatePresence>
          </motion.div>
        )}
      </div>
    </div>
  )
}
