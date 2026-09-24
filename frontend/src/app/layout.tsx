import type { ReactNode } from 'react'
import type { Metadata } from 'next'
import { Orbitron } from 'next/font/google'
import { Providers } from './providers'
import './globals.css'

export const metadata: Metadata = {
  title: 'Andromeda',
  description:
    'Andromeda is a social network for sharing posts, joining groups and organising events with your people.',
}

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
