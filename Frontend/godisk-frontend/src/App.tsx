import { BrowserRouter, Link, Route, Routes, NavLink } from 'react-router-dom'
import { useState } from 'react'
import { CommandTerminal } from '@/components/CommandTerminal'
import { LoginModal } from '@/components/LoginModal'
import { DiskExplorer } from '@/components/DiskExplorer'
import { JournalPanel } from '@/components/JournalPanel'
import { ScriptRunner } from '@/components/ScriptRunner'
import ReportsPage from '@/pages/Reports'
import LogsPage from '@/pages/Logs'

export default function App() {
  const [showLogin, setShowLogin] = useState(false)

  return (
    <BrowserRouter>
      <div className="min-h-screen flex flex-col bg-gray-50">
        <header className="bg-white shadow-sm border-b sticky top-0 z-20">
          <div className="mx-auto max-w-7xl px-4 py-3 flex flex-wrap items-center gap-4">
            <Link to="/" className="text-2xl font-bold tracking-tight text-blue-700 hover:text-blue-900 transition">GoDisk 2.0</Link>
            <nav className="flex items-center gap-2 text-base">
              <NavLink to="/" end className={({isActive})=>isActive?"font-bold text-blue-700 border-b-2 border-blue-700 px-2 py-1":"text-slate-600 hover:text-blue-700 px-2 py-1 transition"}>Terminal</NavLink>
              <NavLink to="/explorer" className={({isActive})=>isActive?"font-bold text-blue-700 border-b-2 border-blue-700 px-2 py-1":"text-slate-600 hover:text-blue-700 px-2 py-1 transition"}>Explorador</NavLink>
              <NavLink to="/reports" className={({isActive})=>isActive?"font-bold text-blue-700 border-b-2 border-blue-700 px-2 py-1":"text-slate-600 hover:text-blue-700 px-2 py-1 transition"}>Reportes</NavLink>
              <NavLink to="/logs" className={({isActive})=>isActive?"font-bold text-blue-700 border-b-2 border-blue-700 px-2 py-1":"text-slate-600 hover:text-blue-700 px-2 py-1 transition"}>Logs</NavLink>
            </nav>
            <div className="ml-auto flex items-center gap-2">
              <button className="px-4 py-2 rounded-lg bg-blue-700 text-white font-semibold shadow hover:bg-blue-800 transition" onClick={() => setShowLogin(true)}>Login</button>
            </div>
          </div>
        </header>

        <main className="flex-1 mx-auto max-w-7xl w-full p-4">
          <Routes>
            <Route path="/" element={
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <section className="lg:col-span-2 bg-white rounded-2xl shadow-md border p-6 flex flex-col">
                  <h2 className="font-bold text-xl mb-4 text-blue-700">Terminal</h2>
                  <CommandTerminal />
                </section>
                <aside className="grid gap-6">
                  <section className="bg-white rounded-2xl shadow-md border p-6">
                    <h2 className="font-bold text-lg mb-3 text-blue-700">Journaling</h2>
                    <JournalPanel />
                  </section>
                  <section className="bg-white rounded-2xl shadow-md border p-6">
                    <h2 className="font-bold text-lg mb-3 text-blue-700">Ejecutar Script .smia</h2>
                    <ScriptRunner />
                  </section>
                </aside>
              </div>
            }/>
            <Route path="/explorer" element={
              <section className="bg-white rounded-2xl shadow-md border p-6">
                <h2 className="font-bold text-xl mb-4 text-blue-700">Explorador de Disco / Partición / FS</h2>
                <DiskExplorer />
              </section>
            }/>
            <Route path="/reports" element={<ReportsPage />} />
            <Route path="/logs" element={<LogsPage />} />
          </Routes>
        </main>

        <LoginModal open={showLogin} onClose={() => setShowLogin(false)} />
      </div>
    </BrowserRouter>
  )
}