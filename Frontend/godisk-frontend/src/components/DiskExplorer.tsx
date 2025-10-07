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
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">Explorador de Discos</h3>
        <button
          onClick={() => { loadDisks(); loadMounted(); }}
          disabled={loading}
          className="px-3 py-1 text-sm rounded-lg border hover:bg-gray-50 disabled:opacity-50"
        >
          🔄 Recargar
        </button>
      </div>

      <div className="grid md:grid-cols-2 gap-4">
        {/* Lista de discos */}
        <div>
          <h4 className="font-medium mb-2">Discos disponibles (.mia)</h4>
          <div className="space-y-2 h-64 overflow-auto scroll-slim bg-[var(--muted)] rounded-xl p-3">
            {loading && <div className="text-sm text-slate-500">Cargando...</div>}
            {!loading && disks.length === 0 && (
              <div className="text-sm text-slate-500">No hay discos .mia en el directorio actual</div>
            )}
            {disks.map((disk, i) => (
              <div
                key={i}
                onClick={() => selectDisk(disk.path)}
                className="p-2 rounded border bg-white hover:bg-blue-50 cursor-pointer"
              >
                <div className="font-medium text-sm">{disk.name}</div>
                <div className="text-xs text-slate-500">
                  {(disk.size / (1024*1024)).toFixed(2)} MB
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Particiones montadas */}
        <div>
          <h4 className="font-medium mb-2">Particiones montadas</h4>
          <div className="space-y-2 h-64 overflow-auto scroll-slim bg-[var(--muted)] rounded-xl p-3">
            {mounted.length === 0 && (
              <div className="text-sm text-slate-500">No hay particiones montadas</div>
            )}
            {mounted.map((m, i) => (
              <div key={i} className="p-2 rounded border bg-white">
                <div className="font-medium text-sm font-mono">{m.mount_id}</div>
                <div className="text-xs text-slate-500">{m.disk_path}</div>
                <div className="text-xs text-slate-500">Partición: {m.partition_id}</div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Información del disco seleccionado */}
      {selectedDisk && (
        <div className="mt-4 p-4 rounded-xl border bg-white">
          <h4 className="font-medium mb-3">Información del disco</h4>
          <div className="grid md:grid-cols-2 gap-4 text-sm">
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
            <div className="mt-4">
              <h5 className="font-medium mb-2">Particiones</h5>
              <div className="space-y-2">
                {selectedDisk.partitions.map((p: any, i: number) => (
                  <div key={i} className="p-2 rounded border bg-gray-50 text-sm">
                    <div className="flex justify-between">
                      <span className="font-medium">{p.name}</span>
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