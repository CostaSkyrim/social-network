import { Component, type ErrorInfo, type ReactNode } from 'react'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  has_error: boolean
  error?: Error
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { has_error: false }

  static getDerivedStateFromError(error: Error): State {
    return { has_error: true, error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('ErrorBoundary caught:', error, info)
  }

  render() {
    if (this.state.has_error) {
      return (
        this.props.fallback || (
          <div className="flex min-h-[400px] flex-col items-center justify-center text-center">
            <div className="mb-4 text-5xl">⚠️</div>
            <h2 className="text-lg font-semibold text-gray-100">Something went wrong</h2>
            <p className="mt-1 text-sm text-gray-300">{this.state.error?.message}</p>
            <button
              onClick={() => this.setState({ has_error: false })}
              className="mt-4 rounded-lg bg-violet-600 px-4 py-2 text-sm font-medium text-white hover:bg-violet-700"
            >
              Try again
            </button>
          </div>
        )
      )
    }
    return this.props.children
  }
}
