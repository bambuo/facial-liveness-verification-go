import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Toaster } from 'react-hot-toast'
import 'antd/dist/reset.css'
import './index.css'
import App from './App.tsx'
import { AntdProvider } from './components/antd/AntdProvider'
import { initTheme } from './theme'
import { initI18n } from './i18n'

// Initialize theme from localStorage (dark default)
initTheme()

// Create a client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

async function bootstrap() {
  await initI18n()

  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <BrowserRouter>
        <QueryClientProvider client={queryClient}>
          <AntdProvider>
            <App />
          </AntdProvider>
          <Toaster
            position="top-right"
            toastOptions={{
              duration: 4000,
              style: {
                background: '#363636',
                color: '#fff',
              },
              success: {
                duration: 3000,
                iconTheme: {
                  primary: '#22c55e',
                  secondary: '#fff',
                },
              },
              error: {
                duration: 5000,
                iconTheme: {
                  primary: '#ef4444',
                  secondary: '#fff',
                },
              },
            }}
          />
        </QueryClientProvider>
      </BrowserRouter>
    </StrictMode>,
  )
}

void bootstrap()
