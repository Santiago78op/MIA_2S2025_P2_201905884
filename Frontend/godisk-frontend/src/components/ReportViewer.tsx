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
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="text-2xl font-bold text-blue-700">Visualizador de Reportes</h3>
        <button
          onClick={loadMountedList}
          className="px-4 py-2 text-base rounded-lg border bg-blue-700 text-white font-semibold shadow hover:bg-blue-800 transition"
        >
          🔄 Cargar montajes
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Controles */}
        <div className="space-y-4">
          <div>
            <label className="block text-base font-semibold mb-2 text-blue-700">Tipo de reporte</label>
            <select
              value={reportType}
              onChange={(e) => setReportType(e.target.value as ReportType)}
              className="w-full px-4 py-2 rounded-lg border outline-none focus:ring-2 ring-blue-700/20 bg-white text-base shadow"
            >
              <option value="mbr">MBR - Estructura del disco</option>
              <option value="disk">Disk - Uso del disco</option>
              <option value="sb">Superblock - Información del FS</option>
              <option value="tree">Tree - Árbol de archivos</option>
              <option value="journal">Journal - Journaling (EXT3)</option>
            </select>
          </div>

          <div>
            <label className="block text-base font-semibold mb-2 text-blue-700">ID de montaje</label>
            <input
              type="text"
              value={mountId}
              onChange={(e) => setMountId(e.target.value)}
              placeholder="Ej: A1, B2, etc."
              className="w-full px-4 py-2 rounded-lg border outline-none focus:ring-2 ring-blue-700/20 bg-white text-base shadow"
            />
            {mounted.length > 0 && (
              <div className="mt-2 flex flex-wrap gap-2">
                {mounted.map(id => (
                  <button
                    key={id}
                    onClick={() => setMountId(id)}
                    className="px-3 py-1 text-xs rounded-lg border bg-white hover:bg-blue-50 font-semibold text-blue-700 shadow-sm"
                  >
                    {id}
                  </button>
                ))}
              </div>
            )}
          </div>

          {reportType === 'tree' && (
            <div>
              <label className="block text-base font-semibold mb-2 text-blue-700">Ruta</label>
              <input
                type="text"
                value={path}
                onChange={(e) => setPath(e.target.value)}
                placeholder="/"
                className="w-full px-4 py-2 rounded-lg border outline-none focus:ring-2 ring-blue-700/20 bg-white text-base shadow"
              />
            </div>
          )}

          <button
            onClick={generateReport}
            disabled={loading}
            className="w-full px-5 py-2 rounded-lg bg-blue-700 text-white font-semibold shadow hover:bg-blue-800 transition disabled:opacity-50"
          >
            {loading ? 'Generando...' : 'Generar Reporte'}
          </button>

          {/* Info */}
          <div className="p-4 rounded-2xl bg-blue-50 border border-blue-200 text-base">
            <div className="font-bold mb-2 text-blue-700">💡 Ayuda</div>
            <ul className="text-sm space-y-1 text-slate-600">
              <li><strong>MBR:</strong> Muestra particiones del disco</li>
              <li><strong>Disk:</strong> Visualización del uso del espacio</li>
              <li><strong>Superblock:</strong> Metadata del sistema de archivos</li>
              <li><strong>Tree:</strong> Estructura de directorios y archivos</li>
              <li><strong>Journal:</strong> Operaciones registradas (solo EXT3)</li>
            </ul>
          </div>
        </div>

        {/* Visualización */}
        <div className="bg-gray-100 rounded-2xl p-6 border shadow-inner flex items-center justify-center min-h-[20rem]">
          {dotContent ? (
            <DotViewer dot={dotContent} />
          ) : (
            <div className="text-slate-500 text-base text-center">
              Selecciona un tipo de reporte y genera la visualización
            </div>
          )}
        </div>
      </div>

      <Toasts />
    </div>
  )
}
