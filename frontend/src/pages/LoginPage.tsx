import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Eye, EyeOff } from 'lucide-react'
import { authApi, ApiError } from '@/api/client'
import { useAuthStore } from '@/store/auth'

export function LoginPage() {
  const navigate   = useNavigate()
  const setAuth    = useAuthStore(s => s.setAuth)
  const [email,    setEmail]    = useState('')
  const [password, setPassword] = useState('')
  const [showPw,   setShowPw]   = useState(false)
  const [error,    setError]    = useState('')
  const [loading,  setLoading]  = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const res = await authApi.login(email, password)
      setAuth({ id: res.user.id, email: res.user.email, name: res.user.name }, res.accessToken, res.refreshToken)
      navigate('/', { replace: true })
    } catch (err) {
      if (err instanceof ApiError) setError('Invalid email or password')
      else setError('Something went wrong. Try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-dvh bg-bg flex flex-col items-center justify-center px-6 py-12 max-w-md mx-auto">
      {/* Logo mark */}
      <motion.div
        initial={{ opacity: 0, scale: 0.8 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ type: 'spring', stiffness: 300, damping: 20 }}
        className="mb-10"
      >
        <div className="w-16 h-16 rounded-2xl bg-green/15 border border-green/30 flex items-center justify-center shadow-green mx-auto mb-6">
          <div className="w-6 h-6 rounded-full bg-green shadow-green" />
        </div>
        <h1 className="font-display font-bold text-4xl text-primary text-center leading-none">
          habit
        </h1>
        <p className="text-secondary font-body text-sm text-center mt-2">
          track what matters
        </p>
      </motion.div>

      <motion.form
        onSubmit={handleSubmit}
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1, duration: 0.4 }}
        className="w-full space-y-4"
      >
        {/* Email */}
        <div className="space-y-1.5">
          <label className="text-secondary font-display text-xs font-semibold uppercase tracking-widest">
            Email
          </label>
          <input
            type="email"
            value={email}
            onChange={e => setEmail(e.target.value)}
            placeholder="you@example.com"
            className="input-field"
            autoComplete="email"
            required
          />
        </div>

        {/* Password */}
        <div className="space-y-1.5">
          <label className="text-secondary font-display text-xs font-semibold uppercase tracking-widest">
            Password
          </label>
          <div className="relative">
            <input
              type={showPw ? 'text' : 'password'}
              value={password}
              onChange={e => setPassword(e.target.value)}
              placeholder="••••••••"
              className="input-field pr-12"
              autoComplete="current-password"
              required
            />
            <button
              type="button"
              onClick={() => setShowPw(v => !v)}
              className="absolute right-4 top-1/2 -translate-y-1/2 text-faint hover:text-secondary transition-colors"
            >
              {showPw ? <EyeOff size={16} /> : <Eye size={16} />}
            </button>
          </div>
        </div>

        {/* Error */}
        {error && (
          <motion.p
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-red-400 text-sm font-body text-center"
          >
            {error}
          </motion.p>
        )}

        {/* Submit */}
        <motion.button
          type="submit"
          disabled={loading}
          whileTap={{ scale: 0.97 }}
          className="btn-primary mt-2 disabled:opacity-50"
        >
          {loading ? (
            <span className="flex items-center justify-center gap-2">
              <span className="w-4 h-4 border-2 border-bg/30 border-t-bg rounded-full animate-spin" />
              Signing in…
            </span>
          ) : 'Sign in'}
        </motion.button>
      </motion.form>

      <motion.p
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.3 }}
        className="text-secondary font-body text-sm text-center mt-8"
      >
        No account?{' '}
        <Link to="/register" className="text-green hover:text-green/80 font-medium transition-colors">
          Create one
        </Link>
      </motion.p>
    </div>
  )
}
