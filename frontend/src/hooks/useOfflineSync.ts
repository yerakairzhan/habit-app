import { useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { progressApi } from '@/api/client'
import { useOfflineStore } from '@/store/offline'
import { QK } from './queries'

/**
 * useOfflineSync — listens for network-online events and replays
 * any queued progress mutations in order.
 */
export function useOfflineSync() {
  const qc         = useQueryClient()
  const queue      = useOfflineStore(s => s.queue)
  const dequeue    = useOfflineStore(s => s.dequeue)
  const setSyncing = useOfflineStore(s => s.setSyncing)

  useEffect(() => {
    async function sync() {
      if (queue.length === 0) return
      setSyncing(true)

      let processed = 0
      for (const action of queue) {
        try {
          if      (action.type === 'INCREMENT') await progressApi.increment(action.habitId, action.date)
          else if (action.type === 'DECREMENT') await progressApi.decrement(action.habitId, action.date)
          else if (action.type === 'RESET')     await progressApi.reset(action.habitId, action.date)
          processed++
        } catch {
          // Stop on first failure — keep remaining in queue
          break
        }
      }

      dequeue(processed)
      setSyncing(false)

      if (processed > 0) {
        qc.invalidateQueries({ queryKey: QK.todayHabits })
      }
    }

    window.addEventListener('online', sync)
    // Also try on mount (if we're online and have queued items)
    if (navigator.onLine) sync()

    return () => window.removeEventListener('online', sync)
  }, [queue, dequeue, setSyncing, qc])
}
