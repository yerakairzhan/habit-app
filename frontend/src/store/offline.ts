import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import type { OfflineAction } from '@/types'

interface OfflineStore {
  queue:       OfflineAction[]
  isSyncing:   boolean
  enqueue:     (action: OfflineAction) => void
  dequeue:     (count: number) => void
  setSyncing:  (v: boolean) => void
  clearQueue:  () => void
}

export const useOfflineStore = create<OfflineStore>()(
  persist(
    (set) => ({
      queue:      [],
      isSyncing:  false,
      enqueue:    (action) => set((s) => ({ queue: [...s.queue, action] })),
      dequeue:    (count)  => set((s) => ({ queue: s.queue.slice(count) })),
      setSyncing: (v)      => set({ isSyncing: v }),
      clearQueue: ()       => set({ queue: [] }),
    }),
    {
      name:    'habit-offline-queue',
      storage: createJSONStorage(() => localStorage),
    },
  ),
)
