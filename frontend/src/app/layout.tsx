import type { ReactNode } from 'react'
import { Orbitron } from 'next/font/google'
import { Providers } from './providers'
import './globals.css'

const orbitron = Orbitron({
  subsets: ['latin'],
  variable: '--font-orbitron',
  display: 'swap',
})

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning className={orbitron.variable}>
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  )
}
