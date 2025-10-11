import { useEffect, useState } from 'react'
import { listDisks, listMounted, getDiskInfo, type MountedPartition } from '@/lib/api'
import { useToast } from '@/lib/useToast'

export function DiskExplorer() {
  const [disks, setDisks] = useState<any[]>([])
  const [mounted, setMounted] = useState<MountedPartition[]>([])
  const [selectedDisk, setSelectedDisk] = useState<any>(null)
  const [loading, setLoading] = useState(false)
  const { push, View: Toasts } = useToast()

  useEffect(() => {
    loadDisks()
    loadMounted()
  }, [])

  async function loadDisks() {
    setLoading(true)
    try {
      const res = await listDisks()
      if (res.ok) {
        setDisks(res.disks || [])
      } else {
        push(res.error || 'Error al cargar discos', 'error')
      }
    } catch (e: any) {
      push(e.message || 'Error de red', 'error')
    } finally {
      setLoading(false)
    }
  }

  async function loadMounted() {
    try {
      const res = await listMounted()
      if (res.ok) {
        setMounted(res.partitions || [])
      }
    } catch (e: any) {
      console.error('Error loading mounted partitions:', e)
    }
  }

  async function selectDisk(diskPath: string) {
    setLoading(true)
    try {
      const info = await getDiskInfo(diskPath)
      if (info.ok) {
        setSelectedDisk(info)
      } else {
        push(info.error || 'Error al obtener info del disco', 'error')
      }
    } catch (e: any) {
      push(e.message || 'Error de red', 'error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="text-2xl font-bold text-blue-700">Explorador de Discos</h3>
        <button
          onClick={() => { loadDisks(); loadMounted(); }}
          disabled={loading}
          className="px-4 py-2 text-base rounded-lg border bg-blue-700 text-white font-semibold shadow hover:bg-blue-800 transition disabled:opacity-50"
        >
          🔄 Recargar
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Lista de discos */}
        <div>
          <h4 className="font-semibold mb-3 text-blue-700">Discos disponibles (.mia)</h4>
          <div className="space-y-2 h-64 overflow-auto scroll-slim bg-gray-100 rounded-2xl p-4 border shadow-inner">
            {loading && <div className="text-base text-slate-500">Cargando...</div>}
            {!loading && disks.length === 0 && (
              <div className="text-base text-slate-500">No hay discos .mia en el directorio actual</div>
            )}
            {disks.map((disk, i) => (
              <div
                key={i}
                onClick={() => selectDisk(disk.path)}
                className="p-3 rounded-xl border bg-white hover:bg-blue-50 cursor-pointer flex flex-col gap-1 shadow-sm transition"
              >
                <div className="font-bold text-base text-blue-700">{disk.name}</div>
                <div className="text-xs text-slate-500">
                  {(disk.size / (1024*1024)).toFixed(2)} MB
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Particiones montadas */}
        <div>
          <h4 className="font-semibold mb-3 text-blue-700">Particiones montadas</h4>
          <div className="space-y-2 h-64 overflow-auto scroll-slim bg-gray-100 rounded-2xl p-4 border shadow-inner">
            {mounted.length === 0 && (
              <div className="text-base text-slate-500">No hay particiones montadas</div>
            )}
            {mounted.map((m, i) => (
              <div key={i} className="p-3 rounded-xl border bg-white flex flex-col gap-1 shadow-sm">
                <div className="font-bold text-base font-mono text-blue-700">{m.mount_id}</div>
                <div className="text-xs text-slate-500">{m.disk_path}</div>
                <div className="text-xs text-slate-500">Partición: {m.partition_id}</div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Información del disco seleccionado */}
      {selectedDisk && (
        <div className="mt-2 p-6 rounded-2xl border bg-white shadow-md">
          <h4 className="font-bold mb-4 text-blue-700 text-lg">Información del disco</h4>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-base">
            <div>
              <div className="text-slate-500">Path:</div>
              <div className="font-mono text-xs">{selectedDisk.path}</div>
            </div>
            <div>
              <div className="text-slate-500">Tamaño:</div>
              <div>{(selectedDisk.size / (1024*1024)).toFixed(2)} MB</div>
            </div>
            <div>
              <div className="text-slate-500">Creado:</div>
              <div>{new Date(selectedDisk.created_at).toLocaleString()}</div>
            </div>
            <div>
              <div className="text-slate-500">Fit:</div>
              <div>{selectedDisk.fit}</div>
            </div>
          </div>

          {selectedDisk.partitions && selectedDisk.partitions.length > 0 && (
            <div className="mt-6">
              <h5 className="font-bold mb-3 text-blue-700">Particiones</h5>
              <div className="space-y-2">
                {selectedDisk.partitions.map((p: any, i: number) => (
                  <div key={i} className="p-3 rounded-xl border bg-gray-50 text-base flex flex-col gap-1 shadow-sm">
                    <div className="flex justify-between">
                      <span className="font-bold text-blue-700">{p.name}</span>
                      <span className="text-slate-500">{p.type}</span>
                    </div>
                    <div className="text-xs text-slate-500 mt-1">
                      Inicio: {p.start} | Tamaño: {(p.size / (1024*1024)).toFixed(2)} MB
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      <Toasts />
    </div>
  )
}