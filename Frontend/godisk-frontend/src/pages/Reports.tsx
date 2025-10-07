import { useState } from 'react'
import DotViewer from '@/components/DotViewer'
import { getReportMBR, getReportDisk, getReportSuperblock, getReportFSTree, getReportJournal } from '@/lib/api'

type ReportTab = 'mbr' | 'disk' | 'sb' | 'tree' | 'journal'

export default function ReportsPage() {
  const [id, setId] = useState('')
  const [path, setPath] = useState('/')
  const [tab, setTab] = useState<ReportTab>('mbr')
  const [dot, setDot] = useState('')
  const [busy, setBusy] = useState(false)

  async function load() {
    if (!id) return
    setBusy(true)
    try {
      let d = ''
      if (tab === 'mbr') d = await getReportMBR(id)
      if (tab === 'disk') d = await getReportDisk(id)
      if (tab === 'sb') d = await getReportSuperblock(id)
      if (tab === 'tree') d = await getReportFSTree(id, path)
      if (tab === 'journal') d = await getReportJournal(id)
      setDot(d)
    } catch (e: any) {
      setDot(`digraph G { label="Error: ${(e?.message||'falló')}" }`)
    } finally { setBusy(false) }
  }

  return (
    <section className="bg-white rounded-2xl shadow-sm border p-3">
      <h2 className="font-medium mb-2">Reportes (DOT → Viz.js)</h2>
      <div className="flex flex-wrap items-center gap-2 mb-3">
        <input className="px-3 py-2 rounded-lg border" placeholder="ID montado (ej: A1, B2)" value={id} onChange={e=>setId(e.target.value)} />
        {tab==='tree' && (
          <input className="px-3 py-2 rounded-lg border" placeholder="/" value={path} onChange={e=>setPath(e.target.value)} />
        )}
        <div className="flex items-center gap-2 ml-auto">
          <TabBtn active={tab==='mbr'} onClick={()=>setTab('mbr')}>MBR</TabBtn>
          <TabBtn active={tab==='disk'} onClick={()=>setTab('disk')}>Disk</TabBtn>
          <TabBtn active={tab==='sb'} onClick={()=>setTab('sb')}>SuperBlock</TabBtn>
          <TabBtn active={tab==='tree'} onClick={()=>setTab('tree')}>Tree</TabBtn>
          <TabBtn active={tab==='journal'} onClick={()=>setTab('journal')}>Journal</TabBtn>
          <button className="px-3 py-2 rounded-lg bg-black text-white disabled:opacity-50" onClick={load} disabled={busy || !id}>{busy? 'Cargando…':'Cargar'}</button>
        </div>
      </div>
      <DotViewer dot={dot} />
      <details className="mt-3">
        <summary className="cursor-pointer text-sm text-slate-600">Ver DOT</summary>
        <pre className="text-xs bg-[var(--muted)] p-2 rounded-lg overflow-auto max-h-64">{dot}</pre>
      </details>
    </section>
  )
}

function TabBtn({active, onClick, children}:{active:boolean; onClick:()=>void; children:React.ReactNode}){
  return (
    <button onClick={onClick} className={`px-3 py-2 rounded-lg border ${active? 'bg-black text-white border-black':'bg-white text-black'}`}>{children}</button>
  )
}