import { useNavigate } from 'react-router-dom'
import { PageHeader } from '@/components/layout/PageHeader'
import { HabitForm } from '@/components/habit/HabitForm'
import { useCreateHabit } from '@/hooks/queries'
import type { HabitFormData } from '@/types'

export function CreateHabitPage() {
  const navigate = useNavigate()
  const { mutate, isPending } = useCreateHabit()

  function handleSubmit(data: HabitFormData) {
    mutate(data, {
      onSuccess: () => navigate('/habits', { replace: true }),
    })
  }

  return (
    <div className="page-container pb-28 overflow-y-auto">
      <PageHeader title="new habit" back />
      <div className="mt-4">
        <HabitForm
          onSubmit={handleSubmit}
          isLoading={isPending}
          submitLabel="Create habit"
        />
      </div>
    </div>
  )
}
