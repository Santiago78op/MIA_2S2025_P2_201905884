import { useState, useCallback } from 'react'

type Toast = { id: number; message: string; type?: 'info'|'success'|'error' }

export function useToast() {
  const [toasts, setToasts] = useState<Toast[]>([])
  const push = useCallback((message: string, type?: Toast['type']) => {
    const t = { id: Date.now() + Math.random(), message, type }
    setToasts(v => [...v, t])
    setTimeout(() => setToasts(v => v.filter(x => x.id !== t.id)), 3200)
  }, [])
  const View = () => (
    <div className="fixed top-4 right-4 z-50 space-y-2">
      {toasts.map(t => (
        <div key={t.id} className={`px-3 py-2 rounded-lg shadow border text-sm bg-white ${t.type==='error'?'border-red-300':'border-slate-200'}`}>
          {t.message}
        </div>
      ))}
    </div>
  )
  return { push, View }
}