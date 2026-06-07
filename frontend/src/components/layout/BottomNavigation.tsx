import { NavLink } from 'react-router-dom'
import { motion } from 'framer-motion'
import { CheckSquare, LayoutGrid, Settings } from 'lucide-react'
import { cn } from '@/lib/utils'

const NAV_ITEMS = [
  { to: '/',        icon: CheckSquare, label: 'today'  },
  { to: '/habits',  icon: LayoutGrid,  label: 'habits' },
  { to: '/settings',icon: Settings,    label: 'settings'},
]

export function BottomNavigation() {
  return (
    <nav className="fixed bottom-0 left-0 right-0 max-w-md mx-auto z-50">
      {/* Blur backdrop */}
      <div className="absolute inset-0 bg-bg/80 backdrop-blur-xl border-t border-border" />

      <div className="relative flex items-center justify-around px-4 safe-bottom pt-3 pb-2">
        {NAV_ITEMS.map(({ to, icon: Icon, label }) => (
          <NavLink key={to} to={to} end={to === '/'}>
            {({ isActive }) => (
              <motion.div
                className="flex flex-col items-center gap-1 px-6 py-1"
                whileTap={{ scale: 0.9 }}
              >
                <div className={cn(
                  'relative p-2 rounded-xl transition-all duration-200',
                  isActive && 'bg-green/10',
                )}>
                  <Icon
                    size={22}
                    className={cn(
                      'transition-colors duration-200',
                      isActive ? 'text-green' : 'text-faint',
                    )}
                    strokeWidth={isActive ? 2.5 : 1.8}
                  />
                  {isActive && (
                    <motion.div
                      layoutId="nav-indicator"
                      className="absolute inset-0 rounded-xl bg-green/10"
                      transition={{ type: 'spring', stiffness: 400, damping: 35 }}
                    />
                  )}
                </div>
                <span className={cn(
                  'text-[10px] font-display font-medium tracking-wide transition-colors',
                  isActive ? 'text-green' : 'text-faint',
                )}>
                  {label}
                </span>
              </motion.div>
            )}
          </NavLink>
        ))}
      </div>
    </nav>
  )
}
