import { BrowserRouter, Link, Route, Routes, NavLink } from 'react-router-dom'
import { useState } from 'react'
import { CommandTerminal } from '@/components/CommandTerminal'
import { LoginModal } from '@/components/LoginModal'
import { DiskExplorer } from '@/components/DiskExplorer'
import { JournalPanel } from '@/components/JournalPanel'
import { ScriptRunner } from '@/components/ScriptRunner'
import ReportsPage from '@/pages/Reports'

export default function App() {
  const [showLogin, setShowLogin] = useState(false)

  return (
    <BrowserRouter>
      <div className="min-h-screen grid grid-rows-[auto_1fr]">
        <header className="border-b bg-white">
          <div className="mx-auto max-w-7xl px-4 py-3 flex items-center gap-4">
            <Link to="/" className="text-xl font-semibold">GoDisk 2.0</Link>
            <nav className="ml-6 flex items-center gap-4 text-sm">
              <NavLink to="/" end className={({isActive})=>isActive?"font-semibold":"text-slate-600 hover:text-black"}>Terminal</NavLink>
              <NavLink to="/explorer" className={({isActive})=>isActive?"font-semibold":"text-slate-600 hover:text-black"}>Explorador</NavLink>
              <NavLink to="/reports" className={({isActive})=>isActive?"font-semibold":"text-slate-600 hover:text-black"}>Reportes (DOT)</NavLink>
            </nav>
            <div className="ml-auto flex items-center gap-2">
              <button className="px-3 py-1.5 rounded-lg bg-black text-white hover:opacity-90" onClick={() => setShowLogin(true)}>Login</button>
            </div>
          </div>
        </header>

        <main className="mx-auto max-w-7xl w-full p-4">
          <Routes>
            <Route path="/" element={
              <div className="grid lg:grid-cols-3 gap-4">
                <section className="lg:col-span-2 bg-white rounded-2xl shadow-sm border p-3">
                  <h2 className="font-medium mb-2">Terminal</h2>
                  <CommandTerminal />
                </section>
                <aside className="grid gap-4">
                  <section className="bg-white rounded-2xl shadow-sm border p-3">
                    <h2 className="font-medium mb-2">Journaling</h2>
                    <JournalPanel />
                  </section>
                  <section className="bg-white rounded-2xl shadow-sm border p-3">
                    <h2 className="font-medium mb-2">Ejecutar Script .smia</h2>
                    <ScriptRunner />
                  </section>
                </aside>
              </div>
            }/>
            <Route path="/explorer" element={
              <section className="bg-white rounded-2xl shadow-sm border p-3">
                <h2 className="font-medium mb-2">Explorador de Disco / Partición / FS</h2>
                <DiskExplorer />
              </section>
            }/>
            <Route path="/reports" element={<ReportsPage />} />
          </Routes>
        </main>

        <LoginModal open={showLogin} onClose={() => setShowLogin(false)} />
      </div>
    </BrowserRouter>
  )
}