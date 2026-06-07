import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AppRouter } from '@/routes'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Keep cached data fresh for 30s; serve stale while revalidating
      staleTime:   30_000,
      gcTime:      5 * 60_000,
      retry:       1,
      // Don't retry on 4xx
      retryDelay:  (attempt) => Math.min(1000 * 2 ** attempt, 10_000),
    },
    mutations: {
      retry: 0,
    },
  },
})

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AppRouter />
    </QueryClientProvider>
  )
}
