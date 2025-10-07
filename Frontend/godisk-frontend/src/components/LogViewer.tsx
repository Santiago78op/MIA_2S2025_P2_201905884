import { useEffect, useState } from 'react'
import { getLogs, clearLogs, getLogStats, type LogEntry, type LogLevel } from '@/lib/api'
import { useToast } from '@/lib/useToast'

export function LogViewer() {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [stats, setStats] = useState<Record<string, number>>({})
  const [filter, setFilter] = useState<LogLevel | 'ALL'>('ALL')
  const [limit, setLimit] = useState(100)
  const [autoRefresh, setAutoRefresh] = useState(false)
  const [loading, setLoading] = useState(false)
  const { push, View: Toasts } = useToast()

  useEffect(() => {
    loadLogs()
    loadStats()
  }, [filter, limit])

  useEffect(() => {
    if (!autoRefresh) return

    const interval = setInterval(() => {
      loadLogs()
      loadStats()
    }, 2000)

    return () => clearInterval(interval)
  }, [autoRefresh, filter, limit])

  async function loadLogs() {
    setLoading(true)
    try {
      const levelFilter = filter === 'ALL' ? undefined : filter
      const res = await getLogs(levelFilter, limit)
      if (res.ok && res.logs) {
        setLogs(res.logs)
      } else {
        push(res.error || 'Error al cargar logs', 'error')
      }
    } catch (e: any) {
      push(e.message || 'Error de red', 'error')
    } finally {
      setLoading(false)
    }
  }

  async function loadStats() {
    try {
      const res = await getLogStats()
      if (res.ok && res.stats) {
        setStats(res.stats)
      }
    } catch (e: any) {
      console.error('Error loading stats:', e)
    }
  }

  async function handleClearLogs() {
    if (!confirm('¿Estás seguro de limpiar todos los logs?')) return

    try {
      const res = await clearLogs()
      if (res.ok) {
        push('Logs limpiados correctamente', 'success')
        loadLogs()
        loadStats()
      } else {
        push(res.error || 'Error al limpiar logs', 'error')
      }
    } catch (e: any) {
      push(e.message || 'Error de red', 'error')
    }
  }

  function getLevelColor(level: LogLevel): string {
    switch (level) {
      case 'DEBUG': return 'text-gray-600 bg-gray-100'
      case 'INFO': return 'text-blue-600 bg-blue-100'
      case 'WARN': return 'text-yellow-600 bg-yellow-100'
      case 'ERROR': return 'text-red-600 bg-red-100'
      default: return 'text-gray-600 bg-gray-100'
    }
  }

  function formatTimestamp(timestamp: string): string {
    const date = new Date(timestamp)
    return date.toLocaleString()
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">Visor de Logs del Sistema</h3>
        <div className="flex items-center gap-2">
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
              className="rounded"
            />
            Auto-refresh (2s)
          </label>
          <button
            onClick={loadLogs}
            disabled={loading}
            className="px-3 py-1 text-sm rounded-lg border hover:bg-gray-50 disabled:opacity-50"
          >
            🔄 Recargar
          </button>
          <button
            onClick={handleClearLogs}
            className="px-3 py-1 text-sm rounded-lg border hover:bg-red-50 text-red-600"
          >
            🗑️ Limpiar
          </button>
        </div>
      </div>

      {/* Stats */}
      <div className="flex gap-2 text-sm">
        <div className="px-3 py-2 rounded-lg bg-gray-100 border">
          Total: <strong>{stats.total || 0}</strong>
        </div>
        <div className="px-3 py-2 rounded-lg bg-blue-100 border border-blue-200">
          Info: <strong>{stats.info || 0}</strong>
        </div>
        <div className="px-3 py-2 rounded-lg bg-yellow-100 border border-yellow-200">
          Warn: <strong>{stats.warn || 0}</strong>
        </div>
        <div className="px-3 py-2 rounded-lg bg-red-100 border border-red-200">
          Error: <strong>{stats.error || 0}</strong>
        </div>
        <div className="px-3 py-2 rounded-lg bg-gray-100 border">
          Debug: <strong>{stats.debug || 0}</strong>
        </div>
      </div>

      {/* Filters */}
      <div className="flex gap-2 items-center">
        <label className="text-sm font-medium">Nivel:</label>
        <select
          value={filter}
          onChange={(e) => setFilter(e.target.value as LogLevel | 'ALL')}
          className="px-3 py-1 rounded-lg border outline-none focus:ring-2 ring-black/10"
        >
          <option value="ALL">Todos</option>
          <option value="DEBUG">Debug</option>
          <option value="INFO">Info</option>
          <option value="WARN">Warn</option>
          <option value="ERROR">Error</option>
        </select>

        <label className="text-sm font-medium ml-4">Límite:</label>
        <select
          value={limit}
          onChange={(e) => setLimit(Number(e.target.value))}
          className="px-3 py-1 rounded-lg border outline-none focus:ring-2 ring-black/10"
        >
          <option value={50}>50</option>
          <option value={100}>100</option>
          <option value={200}>200</option>
          <option value={500}>500</option>
        </select>
      </div>

      {/* Logs Table */}
      <div className="border rounded-xl overflow-hidden">
        <div className="h-96 overflow-auto">
          <table className="w-full text-sm">
            <thead className="bg-gray-100 sticky top-0">
              <tr>
                <th className="px-3 py-2 text-left font-medium">Timestamp</th>
                <th className="px-3 py-2 text-left font-medium">Nivel</th>
                <th className="px-3 py-2 text-left font-medium">Mensaje</th>
                <th className="px-3 py-2 text-left font-medium">Contexto</th>
              </tr>
            </thead>
            <tbody>
              {loading && logs.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-3 py-8 text-center text-slate-500">
                    Cargando logs...
                  </td>
                </tr>
              )}
              {!loading && logs.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-3 py-8 text-center text-slate-500">
                    No hay logs disponibles
                  </td>
                </tr>
              )}
              {logs.map((log, i) => (
                <tr key={i} className="border-t hover:bg-gray-50">
                  <td className="px-3 py-2 text-xs text-slate-600 font-mono whitespace-nowrap">
                    {formatTimestamp(log.timestamp)}
                  </td>
                  <td className="px-3 py-2">
                    <span className={`px-2 py-1 rounded text-xs font-medium ${getLevelColor(log.level)}`}>
                      {log.level}
                    </span>
                  </td>
                  <td className="px-3 py-2">{log.message}</td>
                  <td className="px-3 py-2">
                    {log.context && Object.keys(log.context).length > 0 && (
                      <details className="cursor-pointer">
                        <summary className="text-xs text-blue-600">Ver contexto</summary>
                        <pre className="mt-1 text-xs bg-gray-100 p-2 rounded overflow-auto max-w-md">
                          {JSON.stringify(log.context, null, 2)}
                        </pre>
                      </details>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <Toasts />
    </div>
  )
}
