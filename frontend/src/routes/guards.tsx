import { Navigate, Outlet } from 'react-router-dom'
import { useAuthStore } from '@/store/auth'

export function RequireAuth() {
  const isAuth = useAuthStore(s => s.isAuth)
  if (!isAuth) return <Navigate to="/login" replace />
  return <Outlet />
}

export function RequireGuest() {
  const isAuth = useAuthStore(s => s.isAuth)
  if (isAuth) return <Navigate to="/" replace />
  return <Outlet />
}
