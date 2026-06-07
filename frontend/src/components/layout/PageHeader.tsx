import { motion } from 'framer-motion'
import { ArrowLeft } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'

interface PageHeaderProps {
  title:    string
  subtitle?: string
  back?:    boolean
  action?:  React.ReactNode
  className?: string
}

export function PageHeader({ title, subtitle, back, action, className }: PageHeaderProps) {
  const navigate = useNavigate()

  return (
    <motion.header
      initial={{ opacity: 0, y: -8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className={cn('px-6 pt-14 pb-4 flex items-start justify-between', className)}
    >
      <div className="flex items-center gap-3">
        {back && (
          <button
            onClick={() => navigate(-1)}
            className="w-10 h-10 rounded-xl bg-card border border-border flex items-center
                       justify-center text-secondary hover:text-primary transition-colors
                       active:scale-95 mr-1"
          >
            <ArrowLeft size={18} strokeWidth={2} />
          </button>
        )}
        <div>
          <h1 className="font-display font-bold text-3xl text-primary leading-none tracking-tight">
            {title}
          </h1>
          {subtitle && (
            <p className="text-secondary font-body text-sm mt-1">{subtitle}</p>
          )}
        </div>
      </div>

      {action && <div className="ml-auto">{action}</div>}
    </motion.header>
  )
}
