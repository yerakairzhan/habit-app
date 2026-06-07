import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { AppLayout }      from './AppLayout'
import { RequireAuth, RequireGuest } from './guards'
import { LoginPage }      from '@/pages/LoginPage'
import { RegisterPage }   from '@/pages/RegisterPage'
import { TodayPage }      from '@/pages/TodayPage'
import { HabitsPage }     from '@/pages/HabitsPage'
import { CreateHabitPage} from '@/pages/CreateHabitPage'
import { EditHabitPage }  from '@/pages/EditHabitPage'
import { SettingsPage }   from '@/pages/SettingsPage'
import { Navigate }       from 'react-router-dom'

const router = createBrowserRouter([
  // ─── Auth routes (guests only) ─────────────────────────────────────────────
  {
    element: <RequireGuest />,
    children: [
      { path: '/login',    element: <LoginPage /> },
      { path: '/register', element: <RegisterPage /> },
    ],
  },
  // ─── App routes (require auth) ─────────────────────────────────────────────
  {
    element: <RequireAuth />,
    children: [
      {
        element: <AppLayout />,
        children: [
          { path: '/',               element: <TodayPage /> },
          { path: '/habits',         element: <HabitsPage /> },
          { path: '/habits/new',     element: <CreateHabitPage /> },
          { path: '/habits/:id/edit',element: <EditHabitPage /> },
          { path: '/settings',       element: <SettingsPage /> },
        ],
      },
    ],
  },
  // ─── Fallback ─────────────────────────────────────────────────────────────
  { path: '*', element: <Navigate to="/" replace /> },
])

export function AppRouter() {
  return <RouterProvider router={router} />
}
