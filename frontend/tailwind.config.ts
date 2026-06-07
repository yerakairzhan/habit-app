import type { Config } from 'tailwindcss'

const config: Config = {
  darkMode: ['class'],
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        // Core palette
        bg:       '#05070D',
        surface:  '#0C0F1A',
        card:     '#111520',
        border:   '#1C2133',
        muted:    '#2A3047',
        // Text
        primary:  '#FFFFFF',
        secondary:'#8892A4',
        faint:    '#4A5568',
        // Accent
        green: {
          DEFAULT: '#22C55E',
          dim:     '#16A34A',
          glow:    '#4ADE8022',
          subtle:  '#052210',
        },
        // Semantic
        destructive: '#EF4444',
      },
      fontFamily: {
        // Display: editorial, bold
        display: ['"Space Grotesk"', 'system-ui', 'sans-serif'],
        // Mono: for numbers/progress
        mono: ['"JetBrains Mono"', 'monospace'],
        // Body: clean readable
        body: ['"DM Sans"', 'system-ui', 'sans-serif'],
      },
      borderRadius: {
        card: '1rem',
        pill: '9999px',
      },
      animation: {
        'fade-up':    'fadeUp 0.4s ease both',
        'fade-in':    'fadeIn 0.3s ease both',
        'scale-in':   'scaleIn 0.2s ease both',
        'shimmer':    'shimmer 1.5s infinite',
        'pulse-green':'pulseGreen 2s ease-in-out infinite',
        'check-pop':  'checkPop 0.35s cubic-bezier(0.34,1.56,0.64,1) both',
      },
      keyframes: {
        fadeUp: {
          from: { opacity: '0', transform: 'translateY(12px)' },
          to:   { opacity: '1', transform: 'translateY(0)' },
        },
        fadeIn: {
          from: { opacity: '0' },
          to:   { opacity: '1' },
        },
        scaleIn: {
          from: { opacity: '0', transform: 'scale(0.95)' },
          to:   { opacity: '1', transform: 'scale(1)' },
        },
        shimmer: {
          '0%':   { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' },
        },
        pulseGreen: {
          '0%, 100%': { boxShadow: '0 0 0 0 #22C55E33' },
          '50%':      { boxShadow: '0 0 0 8px transparent' },
        },
        checkPop: {
          from: { opacity: '0', transform: 'scale(0.5) rotate(-10deg)' },
          to:   { opacity: '1', transform: 'scale(1) rotate(0deg)' },
        },
      },
      boxShadow: {
        card:  '0 1px 3px rgba(0,0,0,0.5), 0 0 0 1px rgba(28,33,51,0.8)',
        green: '0 0 20px #22C55E33, 0 0 40px #22C55E11',
        glow:  '0 0 30px #22C55E22',
      },
      backgroundImage: {
        'shimmer-gradient': 'linear-gradient(90deg, transparent 0%, rgba(255,255,255,0.03) 50%, transparent 100%)',
        'green-gradient':   'linear-gradient(135deg, #22C55E22 0%, transparent 60%)',
        'card-gradient':    'linear-gradient(145deg, #111520 0%, #0C0F1A 100%)',
      },
    },
  },
  plugins: [],
}

export default config
