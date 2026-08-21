import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'
import '@/i18n'
import '@/utils/request'
import '@/styles/global.scss'
import { installChunkLoadRecovery } from '@/utils/chunkLoadRecovery'

installChunkLoadRecovery()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
