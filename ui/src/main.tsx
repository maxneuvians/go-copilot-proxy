import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { MantineProvider } from '@mantine/core'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <MantineProvider defaultColorScheme="light">
      <App />
    </MantineProvider>
  </StrictMode>,
)
