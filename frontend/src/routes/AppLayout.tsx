import { Outlet } from 'react-router-dom'
import { BottomNavigation } from '@/components/layout/BottomNavigation'
import { useOfflineSync } from '@/hooks/useOfflineSync'

export function AppLayout() {
  // Start offline sync listener at layout level
  useOfflineSync()

  return (
    <>
      <Outlet />
      <BottomNavigation />
    </>
  )
}
