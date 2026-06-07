import { motion, AnimatePresence } from 'framer-motion'
import { Plus, LayoutGrid } from 'lucide-react'
import { Link } from 'react-router-dom'
import { useHabits } from '@/hooks/queries'
import { HabitTile } from '@/components/habit/HabitTile'
import { PageHeader } from '@/components/layout/PageHeader'
import { LoadingState, EmptyState, HabitCardSkeleton } from '@/components/common/LoadingState'

export function HabitsPage() {
  const { data, isLoading, isError } = useHabits()
  const habits = data?.habits ?? []

  return (
    <div className="page-container pb-28">
      <PageHeader
        title="all habits"
        subtitle={!isLoading && habits.length > 0 ? `${habits.length} habit${habits.length !== 1 ? 's' : ''}` : undefined}
        action={
          <Link
            to="/habits/new"
            className="w-10 h-10 rounded-xl bg-green/15 border border-green/30 flex items-center
                       justify-center text-green hover:bg-green/25 transition-colors active:scale-95"
          >
            <Plus size={18} strokeWidth={2.5} />
          </Link>
        }
      />

      <div className="flex-1 mt-2">
        {isLoading && (
          <LoadingState count={4} Card={HabitCardSkeleton} />
        )}

        {isError && (
          <div className="px-6">
            <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-sm font-body">
              Failed to load habits.
            </div>
          </div>
        )}

        {!isLoading && !isError && habits.length === 0 && (
          <EmptyState
            icon={<LayoutGrid size={28} />}
            title="No habits yet"
            message="Create your first habit and start building a streak."
            action={
              <Link
                to="/habits/new"
                className="btn-primary inline-flex items-center gap-2 px-6 py-3 w-auto rounded-xl"
              >
                <Plus size={16} />
                Create habit
              </Link>
            }
          />
        )}

        {!isLoading && habits.length > 0 && (
          <motion.div
            className="space-y-3 px-6"
            initial="hidden"
            animate="show"
            variants={{
              hidden: {},
              show: { transition: { staggerChildren: 0.05 } },
            }}
          >
            <AnimatePresence mode="popLayout">
              {habits.map((habit, i) => (
                <HabitTile key={habit.id} habit={habit} index={i} />
              ))}
            </AnimatePresence>
          </motion.div>
        )}
      </div>
    </div>
  )
}
