import { useParams, useNavigate } from 'react-router-dom'
import { PageHeader } from '@/components/layout/PageHeader'
import { HabitForm } from '@/components/habit/HabitForm'
import { useHabit, useUpdateHabit, useDeleteHabit } from '@/hooks/queries'
import { LoadingState } from '@/components/common/LoadingState'
import type { HabitFormData } from '@/types'
import { motion } from 'framer-motion'
import { Trash2 } from 'lucide-react'
import { useState } from 'react'

export function EditHabitPage() {
  const { id }       = useParams<{ id: string }>()
  const navigate     = useNavigate()
  const { data: habit, isLoading } = useHabit(id!)
  const updateMut    = useUpdateHabit()
  const deleteMut    = useDeleteHabit()
  const [confirmDel, setConfirmDel] = useState(false)

  function handleSubmit(data: HabitFormData) {
    updateMut.mutate({ id: id!, data }, {
      onSuccess: () => navigate('/habits', { replace: true }),
    })
  }

  function handleDelete() {
    if (!confirmDel) { setConfirmDel(true); return }
    deleteMut.mutate(id!, {
      onSuccess: () => navigate('/habits', { replace: true }),
    })
  }

  if (isLoading) return <LoadingState count={3} />

  return (
    <div className="page-container pb-28 overflow-y-auto">
      <PageHeader
        title="edit habit"
        back
        action={
          <motion.button
            whileTap={{ scale: 0.9 }}
            onClick={handleDelete}
            className={`w-10 h-10 rounded-xl flex items-center justify-center transition-all duration-200 ${
              confirmDel
                ? 'bg-red-500/20 border border-red-500/40 text-red-400'
                : 'bg-card border border-border text-secondary hover:text-red-400'
            }`}
          >
            <Trash2 size={16} />
          </motion.button>
        }
      />
      {confirmDel && (
        <motion.div
          initial={{ opacity: 0, y: -8 }}
          animate={{ opacity: 1, y: 0 }}
          className="mx-6 mb-2 p-3 rounded-xl bg-red-500/10 border border-red-500/20
                     text-red-400 text-xs font-body"
        >
          Tap the delete button again to confirm deletion. This cannot be undone.
        </motion.div>
      )}
      <div className="mt-2">
        {habit && (
          <HabitForm
            initial={habit}
            onSubmit={handleSubmit}
            isLoading={updateMut.isPending}
            submitLabel="Save changes"
          />
        )}
      </div>
    </div>
  )
}
