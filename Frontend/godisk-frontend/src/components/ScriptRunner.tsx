import { useRef, useState } from 'react'
import { executeScript, type ScriptResponse } from '@/lib/api'
import { useToast } from '@/lib/useToast'

export function ScriptRunner() {
  const [busy, setBusy] = useState(false)
  const [log, setLog] = useState('')
  const [script, setScript] = useState('')
  const [stats, setStats] = useState<{ total: number; success: number; error: number } | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  const { push, View: Toasts } = useToast()

  async function handleFile(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return

    try {
      const text = await file.text()
      setScript(text)
      await runScript(text)
    } catch (e: any) {
      push(e.message || 'Error al leer el archivo', 'error')
    }

    if (inputRef.current) inputRef.current.value = ''
  }

  async function runScript(scriptText?: string) {
    const text = scriptText || script
    if (!text.trim()) {
      push('No hay script para ejecutar', 'error')
      return
    }

    setBusy(true)
    setLog('Ejecutando script...\n')
    setStats(null)

    try {
      const res: ScriptResponse = await executeScript(text)

      if (res.results) {
        const out: string[] = []

        for (const result of res.results) {
          out.push(`[Línea ${result.line}] $ ${result.input}`)
          if (result.success) {
            out.push(`✓ ${result.output || 'OK'}`)
          } else {
            out.push(`✗ Error: ${result.error}`)
          }
          out.push('')
        }

        setLog(out.join('\n'))
        setStats({
          total: res.executed || 0,
          success: res.success_count || 0,
          error: res.error_count || 0
        })

        if (res.ok) {
          push('Script ejecutado exitosamente', 'success')
        } else {
          push(`Script completado con ${res.error_count} errores`, 'error')
        }
      }
    } catch (e: any) {
      setLog(`Error de red: ${e.message}`)
      push(e.message || 'Error al ejecutar script', 'error')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">Ejecutor de Scripts</h3>
        <div className="flex gap-2">
          <label className="px-3 py-2 rounded-lg border bg-white hover:bg-gray-50 cursor-pointer text-sm">
            📁 Cargar .smia
            <input
              ref={inputRef}
              type="file"
              accept=".smia,.txt"
              onChange={handleFile}
              className="hidden"
            />
          </label>
          <button
            onClick={() => runScript()}
            disabled={busy || !script}
            className="px-3 py-2 rounded-lg bg-black text-white hover:opacity-90 disabled:opacity-50 text-sm"
          >
            {busy ? 'Ejecutando...' : '▶️ Ejecutar'}
          </button>
        </div>
      </div>

      {/* Editor de script */}
      <div>
        <label className="block text-sm font-medium mb-1">Script</label>
        <textarea
          value={script}
          onChange={(e) => setScript(e.target.value)}
          placeholder="Escribe o carga un script aquí..."
          className="w-full h-32 px-3 py-2 rounded-lg border outline-none focus:ring-2 ring-black/10 font-mono text-sm"
          disabled={busy}
        />
      </div>

      {/* Estadísticas */}
      {stats && (
        <div className="flex gap-4 text-sm">
          <div className="px-3 py-2 rounded-lg bg-blue-50 border border-blue-200">
            <span className="font-medium">Total:</span> {stats.total}
          </div>
          <div className="px-3 py-2 rounded-lg bg-green-50 border border-green-200">
            <span className="font-medium">✓ Exitosos:</span> {stats.success}
          </div>
          <div className="px-3 py-2 rounded-lg bg-red-50 border border-red-200">
            <span className="font-medium">✗ Errores:</span> {stats.error}
          </div>
        </div>
      )}

      {/* Log de salida */}
      <div>
        <label className="block text-sm font-medium mb-1">Salida</label>
        <pre className="h-64 overflow-auto scroll-slim bg-[var(--muted)] rounded-xl p-3 text-sm font-mono">
          {log || 'Carga un archivo .smia o escribe un script para ejecutar.'}
        </pre>
      </div>

      <p className="text-xs text-slate-500">
        💡 Puedes cargar un archivo .smia o escribir comandos directamente. Los comandos se ejecutan línea por línea.
      </p>

      <Toasts />
    </div>
  )
}