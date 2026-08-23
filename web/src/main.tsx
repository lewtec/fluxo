import React from 'react'
import ReactDOM from 'react-dom/client'
import { RelayEnvironmentProvider } from 'react-relay'
import { Buffer } from 'buffer'
import { environment } from './relay/environment'
import App from './App'
import './index.css'

// parse-torrent/bencode expect Buffer on the same object Vite's `global`
// define resolves to (globalThis), not a one-off window assignment in a page.
const g = globalThis as typeof globalThis & { Buffer?: typeof Buffer }
g.Buffer = Buffer

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <RelayEnvironmentProvider environment={environment}>
      <App />
    </RelayEnvironmentProvider>
  </React.StrictMode>,
)
