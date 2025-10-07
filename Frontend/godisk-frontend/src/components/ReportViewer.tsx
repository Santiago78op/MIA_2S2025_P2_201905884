import { useState } from 'react'
import { getReportMBR, getReportDisk, getReportSuperblock, getReportFSTree, getReportJournal, listMounted } from '@/lib/api'
import { useToast } from '@/lib/useToast'
import DotViewer from './DotViewer'

type ReportType = 'mbr' | 'disk' | 'sb' | 'tree' | 'journal'

export function ReportViewer() {
  const [reportType, setReportType] = useState<ReportType>('mbr')
  const [mountId, setMountId] = useState('')
  const [path, setPath] = useState('/')
  const [dotContent, setDotContent] = useState('')
  const [loading, setLoading] = useState(false)
  const [mounted, setMounted] = useState<string[]>([])
  const { push, View: Toasts } = useToast()

  async function loadMountedList() {
    try {
      const res = await listMounted()
      if (res.ok && res.partitions) {
        setMounted(res.partitions.map(p => p.mount_id))
      }
    } catch (e: any) {
      console.error('Error loading mounted:', e)
    }
  }

  async function generateReport() {
    if (!mountId) {
      push('Debes ingresar un ID de montaje', 'error')
      return
    }

    setLoading(true)
    setDotContent('')

    try {
      let dot = ''

      switch (reportType) {
        case 'mbr':
          dot = await getReportMBR(mountId)
          break
        case 'disk':
          dot = await getReportDisk(mountId)
          break
        case 'sb':
          dot = await getReportSuperblock(mountId)
          break
        case 'tree':
          dot = await getReportFSTree(mountId, path)
          break
        case 'journal':
          dot = await getReportJournal(mountId)
          break
      }

      setDotContent(dot)
      push('Reporte generado correctamente', 'success')
    } catch (e: any) {
      push(e.message || 'Error al generar reporte', 'error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">Visualizador de Reportes</h3>
        <button
          onClick={loadMountedList}
          className="px-3 py-1 text-sm rounded-lg border hover:bg-gray-50"
        >
          🔄 Cargar montajes
        </button>
      </div>

      <div className="grid md:grid-cols-2 gap-4">
        {/* Controles */}
        <div className="space-y-3">
          <div>
            <label className="block text-sm font-medium mb-1">Tipo de reporte</label>
            <select
              value={reportType}
              onChange={(e) => setReportType(e.target.value as ReportType)}
              className="w-full px-3 py-2 rounded-lg border outline-none focus:ring-2 ring-black/10"
            >
              <option value="mbr">MBR - Estructura del disco</option>
              <option value="disk">Disk - Uso del disco</option>
              <option value="sb">Superblock - Información del FS</option>
              <option value="tree">Tree - Árbol de archivos</option>
              <option value="journal">Journal - Journaling (EXT3)</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">ID de montaje</label>
            <input
              type="text"
              value={mountId}
              onChange={(e) => setMountId(e.target.value)}
              placeholder="Ej: A1, B2, etc."
              className="w-full px-3 py-2 rounded-lg border outline-none focus:ring-2 ring-black/10"
            />
            {mounted.length > 0 && (
              <div className="mt-2 flex flex-wrap gap-1">
                {mounted.map(id => (
                  <button
                    key={id}
                    onClick={() => setMountId(id)}
                    className="px-2 py-1 text-xs rounded border bg-white hover:bg-blue-50"
                  >
                    {id}
                  </button>
                ))}
              </div>
            )}
          </div>

          {reportType === 'tree' && (
            <div>
              <label className="block text-sm font-medium mb-1">Ruta</label>
              <input
                type="text"
                value={path}
                onChange={(e) => setPath(e.target.value)}
                placeholder="/"
                className="w-full px-3 py-2 rounded-lg border outline-none focus:ring-2 ring-black/10"
              />
            </div>
          )}

          <button
            onClick={generateReport}
            disabled={loading}
            className="w-full px-4 py-2 rounded-lg bg-black text-white hover:opacity-90 disabled:opacity-50"
          >
            {loading ? 'Generando...' : 'Generar Reporte'}
          </button>

          {/* Info */}
          <div className="p-3 rounded-lg bg-blue-50 border border-blue-200 text-sm">
            <div className="font-medium mb-1">💡 Ayuda</div>
            <ul className="text-xs space-y-1 text-slate-600">
              <li><strong>MBR:</strong> Muestra particiones del disco</li>
              <li><strong>Disk:</strong> Visualización del uso del espacio</li>
              <li><strong>Superblock:</strong> Metadata del sistema de archivos</li>
              <li><strong>Tree:</strong> Estructura de directorios y archivos</li>
              <li><strong>Journal:</strong> Operaciones registradas (solo EXT3)</li>
            </ul>
          </div>
        </div>

        {/* Visualización */}
        <div className="bg-[var(--muted)] rounded-xl p-4">
          {dotContent ? (
            <DotViewer dot={dotContent} />
          ) : (
            <div className="flex items-center justify-center h-64 text-slate-500 text-sm">
              Selecciona un tipo de reporte y genera la visualización
            </div>
          )}
        </div>
      </div>

      <Toasts />
    </div>
  )
}
