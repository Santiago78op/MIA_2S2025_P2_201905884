import { useState } from 'react'
import { runCmd } from '@/lib/api'

export function LoginModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const [usr, setUsr] = useState('')
  const [pwd, setPwd] = useState('')
  const [busy, setBusy] = useState(false)
  if (!open) return null

  async function doLogin() {
    setBusy(true)
    try {
      // Mientras no exista endpoint /api/login, usa el comando si tu backend lo soporta:
      // "login -usr=... -pwd=..."
      const res = await runCmd(`login -usr=${JSON.stringify(usr)} -pwd=${JSON.stringify(pwd)}`)
      if (res.ok) {
        onClose()
      } else {
        alert(res.error || 'Login falló')
      }
    } finally { setBusy(false) }
  }

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-2xl w-full max-w-md p-8 border shadow-xl flex flex-col gap-4">
        <h3 className="text-2xl font-bold text-blue-700 text-center">Iniciar sesión</h3>
        <p className="text-base text-slate-500 mb-2 text-center">Ingresa tus credenciales para acceder al sistema.</p>
        <div className="space-y-3">
          <input className="w-full px-4 py-2 rounded-lg border outline-none focus:ring-2 ring-blue-700/20 bg-white text-base shadow" placeholder="Usuario" value={usr} onChange={e=>setUsr(e.target.value)} />
          <input type="password" className="w-full px-4 py-2 rounded-lg border outline-none focus:ring-2 ring-blue-700/20 bg-white text-base shadow" placeholder="Contraseña" value={pwd} onChange={e=>setPwd(e.target.value)} />
        </div>
        <div className="mt-4 flex justify-end gap-2">
          <button className="px-5 py-2 rounded-lg border bg-gray-100 text-blue-700 font-semibold shadow hover:bg-blue-50 transition" onClick={onClose} disabled={busy}>Cancelar</button>
          <button className="px-5 py-2 rounded-lg bg-blue-700 text-white font-semibold shadow hover:bg-blue-800 transition disabled:opacity-50" onClick={doLogin} disabled={busy}>Entrar</button>
        </div>
      </div>
    </div>
  )
}