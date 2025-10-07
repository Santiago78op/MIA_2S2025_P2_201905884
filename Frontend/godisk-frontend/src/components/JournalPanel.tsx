import { useEffect, useState } from 'react'
import { runCmd } from '@/lib/api'

interface JournalEntry {
  Op: string
  Path: string
  Content: string
  Timestamp: string
}

export function JournalPanel() {
  const [id, setId] = useState('')
  const [entries, setEntries] = useState<JournalEntry[]>([])
  const [rawText, setRawText] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [viewMode, setViewMode] = useState<'table' | 'raw'>('table')

  async function load() {
    if (!id) return
    setBusy(true)
    setError('')
    try {
      const res = await runCmd(`journaling -id=${id}`)
      if (res.ok) {
        setRawText(res.output || '')
        // Intentar parsear como JSON
        try {
          const parsed = JSON.parse(res.output || '[]')
          if (Array.isArray(parsed)) {
            setEntries(parsed)
          } else {
            setEntries([])
          }
        } catch {
          // Si no es JSON, solo mostrar raw
          setEntries([])
        }
      } else {
        setError(res.error || 'Error desconocido')
        setRawText('')
        setEntries([])
      }
    } catch (e: any) {
      setError(e.message || 'Error de red')
      setRawText('')
      setEntries([])
    } finally {
      setBusy(false)
    }
  }

  async function recovery() {
    if (!id) return
    if (!confirm('¿Estás seguro de ejecutar recovery? Esto aplicará las operaciones del journal.')) return
    setBusy(true)
    try {
      const res = await runCmd(`recovery -id=${id}`)
      alert(res.ok ? '✅ Recovery ejecutado exitosamente' : `❌ Error: ${res.error || 'error'}`)
      await load()
    } finally {
      setBusy(false)
    }
  }

  async function loss() {
    if (!id) return
    if (!confirm('⚠️ ¿Estás seguro de ejecutar loss? Esto BORRARÁ los datos del filesystem (excepto superblock y journal).')) return
    setBusy(true)
    try {
      const res = await runCmd(`loss -id=${id}`)
      alert(res.ok ? '✅ Loss ejecutado (datos eliminados)' : `❌ Error: ${res.error || 'error'}`)
      await load()
    } finally {
      setBusy(false)
    }
  }

  useEffect(() => {
    // Auto-load cuando cambia el ID (opcional)
  }, [id])

  return (
    <div className="space-y-4">
      {/* Header con controles */}
      <div className="flex gap-2 items-center">
        <input
          className="flex-1 px-3 py-2 rounded-lg border border-gray-300 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          placeholder="ID de partición montada (ej: vd12ab34)"
          value={id}
          onChange={e => setId(e.target.value)}
        />
        <button
          className="px-4 py-2 rounded-lg bg-blue-600 text-white hover:bg-blue-700 disabled:bg-gray-400"
          onClick={load}
          disabled={busy || !id}
        >
          {busy ? 'Cargando...' : 'Cargar Journal'}
        </button>
      </div>

      {/* Botones de acción */}
      <div className="flex gap-2">
        <button
          className="px-4 py-2 rounded-lg bg-green-600 text-white hover:bg-green-700 disabled:bg-gray-400"
          onClick={recovery}
          disabled={busy || !id}
        >
          🔄 Recovery
        </button>
        <button
          className="px-4 py-2 rounded-lg bg-red-600 text-white hover:bg-red-700 disabled:bg-gray-400"
          onClick={loss}
          disabled={busy || !id}
        >
          ⚠️ Loss
        </button>
        <div className="flex-1" />
        <div className="flex gap-1 border rounded-lg">
          <button
            className={`px-3 py-2 rounded-l-lg ${viewMode === 'table' ? 'bg-gray-200' : 'hover:bg-gray-100'}`}
            onClick={() => setViewMode('table')}
          >
            📋 Tabla
          </button>
          <button
            className={`px-3 py-2 rounded-r-lg ${viewMode === 'raw' ? 'bg-gray-200' : 'hover:bg-gray-100'}`}
            onClick={() => setViewMode('raw')}
          >
            📄 Raw
          </button>
        </div>
      </div>

      {/* Error */}
      {error && (
        <div className="p-3 rounded-lg bg-red-50 border border-red-200 text-red-800">
          ❌ {error}
        </div>
      )}

      {/* Tabla de entradas */}
      {viewMode === 'table' && entries.length > 0 && (
        <div className="border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-100 border-b">
              <tr>
                <th className="px-4 py-2 text-left">#</th>
                <th className="px-4 py-2 text-left">Operación</th>
                <th className="px-4 py-2 text-left">Ruta</th>
                <th className="px-4 py-2 text-left">Contenido</th>
                <th className="px-4 py-2 text-left">Timestamp</th>
              </tr>
            </thead>
            <tbody>
              {entries.map((entry, idx) => (
                <tr key={idx} className="border-b hover:bg-gray-50">
                  <td className="px-4 py-2 text-gray-500">{idx}</td>
                  <td className="px-4 py-2 font-mono font-semibold text-blue-600">{entry.Op}</td>
                  <td className="px-4 py-2 font-mono">{entry.Path}</td>
                  <td className="px-4 py-2 font-mono text-xs text-gray-600 max-w-xs truncate">
                    {entry.Content || '-'}
                  </td>
                  <td className="px-4 py-2 text-xs text-gray-500">
                    {entry.Timestamp ? new Date(entry.Timestamp).toLocaleString() : '-'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Vista Raw */}
      {viewMode === 'raw' && (
        <textarea
          className="w-full h-96 rounded-lg border border-gray-300 p-3 font-mono text-sm resize-none"
          value={rawText}
          readOnly
        />
      )}

      {/* Empty state */}
      {!error && entries.length === 0 && rawText === '' && (
        <div className="text-center py-12 text-gray-400">
          Ingresa un ID de partición y presiona "Cargar Journal"
        </div>
      )}

      {!error && entries.length === 0 && rawText !== '' && viewMode === 'table' && (
        <div className="text-center py-12 text-gray-400">
          Journal vacío o sin formato JSON. Usa vista "Raw" para ver el contenido.
        </div>
      )}
    </div>
  )
}