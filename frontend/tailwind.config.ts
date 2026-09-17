import type { Config } from 'tailwindcss'

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        display: ['var(--font-orbitron)', 'sans-serif'],
      },
    },
  },
  plugins: [],
} satisfies Config
