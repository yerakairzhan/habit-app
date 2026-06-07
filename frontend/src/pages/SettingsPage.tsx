import { motion } from 'framer-motion'
import {
  Flame, Trophy, BarChart2, Globe, Bell, Calendar,
  ChevronRight, LogOut, Mail, Shield, FileText, Info,
} from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useStats, useSettings, useUpdateSettings } from '@/hooks/queries'
import { useAuthStore } from '@/store/auth'
import { authApi } from '@/api/client'
import { PageHeader } from '@/components/layout/PageHeader'
import { Skeleton } from '@/components/common/LoadingState'
import type { Settings } from '@/types'

// ─── Section wrapper ─────────────────────────────────────────────────────────
function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="mb-6">
      <p className="text-faint font-display text-xs font-semibold uppercase tracking-widest px-6 mb-2">
        {title}
      </p>
      <div className="mx-6 rounded-card border border-border overflow-hidden divide-y divide-border">
        {children}
      </div>
    </div>
  )
}

// ─── Settings row ─────────────────────────────────────────────────────────────
function SettingsRow({
  icon, label, value, onToggle, href, onPress,
}: {
  icon: React.ReactNode
  label: string
  value?: string | boolean
  onToggle?: (v: boolean) => void
  href?: string
  onPress?: () => void
}) {
  const isToggle = typeof value === 'boolean'

  return (
    <div
      className="flex items-center gap-3 px-4 py-3.5 bg-card cursor-pointer active:bg-surface transition-colors"
      onClick={() => {
        if (isToggle && onToggle) onToggle(!value)
        if (href) window.open(href, '_blank')
        if (onPress) onPress()
      }}
    >
      <span className="text-secondary w-5 flex-shrink-0">{icon}</span>
      <span className="flex-1 font-body text-sm text-primary">{label}</span>
      {isToggle ? (
        <div className={`w-11 h-6 rounded-full transition-all duration-200 relative ${value ? 'bg-green' : 'bg-muted'}`}>
          <div className={`absolute top-0.5 w-5 h-5 rounded-full bg-white shadow transition-all duration-200 ${value ? 'left-5' : 'left-0.5'}`} />
        </div>
      ) : value ? (
        <span className="text-secondary font-body text-sm">{value}</span>
      ) : (
        <ChevronRight size={14} className="text-faint" />
      )}
    </div>
  )
}

// ─── Stat card ────────────────────────────────────────────────────────────────
function StatCard({ icon, label, value, accent = false }: {
  icon: React.ReactNode; label: string; value: string | number; accent?: boolean
}) {
  return (
    <div className="flex-1 bg-card border border-border rounded-xl p-4 text-center">
      <div className={`flex justify-center mb-2 ${accent ? 'text-green' : 'text-secondary'}`}>
        {icon}
      </div>
      <div className={`font-mono font-semibold text-2xl ${accent ? 'text-green' : 'text-primary'}`}>
        {value}
      </div>
      <div className="text-faint font-body text-xs mt-0.5">{label}</div>
    </div>
  )
}

// ─── Page ────────────────────────────────────────────────────────────────────
export function SettingsPage() {
  const navigate      = useNavigate()
  const clearAuth     = useAuthStore(s => s.clearAuth)
  const refreshToken  = useAuthStore(s => s.refreshToken)
  const user          = useAuthStore(s => s.user)

  const { data: stats,    isLoading: statsLoading }    = useStats()
  const { data: settings, isLoading: settingsLoading } = useSettings()
  const updateSettings = useUpdateSettings()

  function patch(partial: Partial<Settings>) {
    if (!settings) return
    updateSettings.mutate({ ...settings, ...partial })
  }

  async function handleLogout() {
    if (refreshToken) {
      try { await authApi.logout(refreshToken) } catch { /* ignore */ }
    }
    clearAuth()
    navigate('/login', { replace: true })
  }

  return (
    <div className="page-container pb-28 overflow-y-auto">
      <PageHeader title="settings" />

      {/* User info */}
      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        className="mx-6 mb-6 p-4 rounded-card bg-card border border-border flex items-center gap-3"
      >
        <div className="w-10 h-10 rounded-xl bg-green/15 border border-green/30 flex items-center justify-center">
          <span className="text-green font-display font-bold text-sm">
            {user?.name?.[0]?.toUpperCase() ?? '?'}
          </span>
        </div>
        <div>
          <p className="font-display font-semibold text-primary text-sm">{user?.name}</p>
          <p className="text-secondary font-body text-xs">{user?.email}</p>
        </div>
      </motion.div>

      {/* Stats */}
      <div className="px-6 mb-6">
        <p className="text-faint font-display text-xs font-semibold uppercase tracking-widest mb-3">
          Statistics
        </p>
        {statsLoading ? (
          <div className="flex gap-3">
            <Skeleton className="flex-1 h-24 rounded-xl" />
            <Skeleton className="flex-1 h-24 rounded-xl" />
            <Skeleton className="flex-1 h-24 rounded-xl" />
          </div>
        ) : (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="flex gap-3"
          >
            <StatCard icon={<BarChart2 size={18} />} label="tracked"  value={stats?.trackedHabitsCount ?? 0} />
            <StatCard icon={<Flame size={18} />}     label="streak"   value={stats?.currentStreak ?? 0}     accent />
            <StatCard icon={<Trophy size={18} />}    label="best"     value={stats?.bestStreak ?? 0} />
          </motion.div>
        )}
        {stats && stats.completionRate30d > 0 && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.2 }}
            className="mt-3 p-3 rounded-xl bg-card border border-border"
          >
            <div className="flex items-center justify-between mb-2">
              <span className="text-secondary font-body text-xs">30-day completion</span>
              <span className="text-green font-mono text-xs font-medium">
                {Math.round(stats.completionRate30d * 100)}%
              </span>
            </div>
            <div className="h-1.5 bg-muted rounded-full overflow-hidden">
              <motion.div
                className="h-full bg-green rounded-full"
                initial={{ width: 0 }}
                animate={{ width: `${stats.completionRate30d * 100}%` }}
                transition={{ delay: 0.3, type: 'spring', stiffness: 150, damping: 25 }}
              />
            </div>
          </motion.div>
        )}
      </div>

      {/* Preferences */}
      {settingsLoading ? (
        <div className="px-6 space-y-2 mb-6">
          {[1, 2, 3].map(i => <Skeleton key={i} className="h-12 rounded-xl" />)}
        </div>
      ) : (
        <Section title="Preferences">
          <SettingsRow
            icon={<Globe size={15} />}
            label="Language"
            value={settings?.language?.toUpperCase() ?? 'EN'}
          />
          <SettingsRow
            icon={<Bell size={15} />}
            label="Notifications"
            value={settings?.notificationsEnabled ?? true}
            onToggle={v => patch({ notificationsEnabled: v })}
          />
          <SettingsRow
            icon={<Calendar size={15} />}
            label="Show dates"
            value={settings?.showDates ?? true}
            onToggle={v => patch({ showDates: v })}
          />
        </Section>
      )}

      {/* Links */}
      <Section title="About">
        <SettingsRow icon={<Mail size={15} />}    label="Contact Us"      href="mailto:support@habitapp.io" />
        <SettingsRow icon={<Shield size={15} />}  label="Privacy Policy"  href="/privacy" />
        <SettingsRow icon={<FileText size={15} />}label="Terms of Use"    href="/terms" />
        <SettingsRow icon={<Info size={15} />}    label="About"           href="/about" />
      </Section>

      {/* Logout */}
      <div className="px-6 mb-6">
        <motion.button
          whileTap={{ scale: 0.97 }}
          onClick={handleLogout}
          className="w-full p-4 rounded-xl bg-card border border-border flex items-center
                     justify-center gap-2 text-secondary hover:text-red-400 hover:border-red-500/30
                     transition-all duration-200 font-body text-sm"
        >
          <LogOut size={15} />
          Sign out
        </motion.button>
      </div>

      <p className="text-center text-faint font-body text-xs pb-4">
        Habit Tracker v1.0.0
      </p>
    </div>
  )
}
