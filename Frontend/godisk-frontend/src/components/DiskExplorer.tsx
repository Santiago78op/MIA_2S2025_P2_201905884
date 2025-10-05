import { useEffect, useState } from 'react'
import { runCmd } from '@/lib/api'

// Este componente funciona en modo "solo lectura" usando comandos que ya tienes (cuando existan):
// - list mounts (si implementas) o ingresa manualmente el id
// - tree -id=... -path=/
// - cat -id=... -path=/ruta/archivo.txt (si implementas)

export function DiskExplorer() {
  const [id, setId] = useState('')
  const [path, setPath] = useState('/')
  const [tree, setTree] = useState<string>('')
  const [file, setFile] = useState<string>('')

  async function loadTree() {
    if (!id) return
    const res = await runCmd(`tree -id=${id} -path=${JSON.stringify(path)}`)
    setTree(res.ok ? (res.output||'') : (res.error||'error'))
  }

  async function openFile(p: string) {
    if (!id) return
    const res = await runCmd(`cat -id=${id} -path=${JSON.stringify(p)}`)
    setFile(res.ok ? (res.output||'') : (res.error||'error'))
  }

  useEffect(() => { /* opcional: auto cargar */ }, [])

  return (
    <div className="grid md:grid-cols-2 gap-3">
      <div>
        <div className="flex gap-2 mb-2">
          <input className="flex-1 px-3 py-2 rounded-lg border" placeholder="ID de partición montada" value={id} onChange={e=>setId(e.target.value)} />
          <input className="flex-1 px-3 py-2 rounded-lg border" placeholder="/" value={path} onChange={e=>setPath(e.target.value)} />
          <button className="px-3 py-2 rounded-lg border" onClick={loadTree}>Cargar</button>
        </div>
        <pre className="h-64 overflow-auto scroll-slim bg-[var(--muted)] rounded-xl p-3 text-sm">{tree || 'Árbol vacío / implementa comando tree en backend.'}</pre>
        <p className="text-xs text-slate-500 mt-1">Consejo: puedes implementar endpoints /api/fs/tree y /api/fs/file para no depender de comandos.</p>
      </div>
      <div>
        <h4 className="font-medium mb-1">Visor de archivo</h4>
        <pre className="h-80 overflow-auto scroll-slim bg-[var(--muted)] rounded-xl p-3 text-sm">{file || 'Selecciona un archivo (usa el comando cat en backend).'}
        </pre>
      </div>
    </div>
  )
}