import { Spinner } from '@/components/ui/Spinner'

export function LoadingScreen() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <Spinner size="lg" />
    </div>
  )
}
