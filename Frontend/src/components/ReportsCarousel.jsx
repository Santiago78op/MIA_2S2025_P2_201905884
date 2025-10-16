import { useMemo, useRef, useState } from 'react'
import { API } from '../lib/api'

const IMG = new Set(['.png','.jpg','.jpeg','.svg','.gif'])
const isImg = p => IMG.has((p?.slice(p.lastIndexOf('.'))||'').toLowerCase())

const REPORTS = [
  { key:'mbr',      title:'MBR',        kind:'image', desc:'Estructura del MBR y particiones primarias/extendida.' },
  { key:'disk',     title:'Disk',       kind:'image', desc:'Mapa visual de ocupación del disco.' },
  { key:'inode',    title:'Inode',      kind:'image', desc:'Detalle de un inodo y sus apuntadores.' },
  { key:'block',    title:'Block',      kind:'image', desc:'Bloques del archivo/carpeta objetivo.' },
  { key:'tree',     title:'Tree',       kind:'image', desc:'Árbol de directorios y archivos (DOT).' },
  { key:'file',     title:'File',       kind:'image', desc:'Vista de un archivo específico.', needsExtra:true, extraLabel:'path_file' },
  { key:'ls',       title:'LS',         kind:'image', desc:'Listado detallado de un directorio.', needsExtra:true, extraLabel:'path_file_ls' },
  { key:'bm_inode', title:'BM Inode',   kind:'text',  desc:'Bitmap de inodos (20 bits por línea).' },
  { key:'bm_block', title:'BM Block',   kind:'text',  desc:'Bitmap de bloques (20 bits por línea).' },
  { key:'sb',       title:'SuperBlock', kind:'text',  desc:'Campos y offsets del SuperBlock.' },
]

export default function ReportsCarousel(){
  const [id, setId] = useState('841A')
  const [busy, setBusy] = useState(false)
  const [result, setResult] = useState({}) // key -> {path, ts, previewExtra}

  const viewport = useRef()

  const defaults = useMemo(()=>({
    out(name, id) {
      const base = './Reports'
      if (name==='bm_inode') return `${base}/bm_inode_${id}.txt`
      if (name==='bm_block') return `${base}/bm_block_${id}.txt`
      if (name==='sb')       return `${base}/sb_${id}.txt`
      return `${base}/${name}_${id}.png`
    }
  }),[])

  function scrollBy(dir){ viewport.current?.scrollBy({left: dir*380, behavior:'smooth'}) }

  async function make(repKey, extraVal){
    const meta = REPORTS.find(r=>r.key===repKey)
    if(!meta) return
    const out = defaults.out(repKey, id)
    setBusy(true)
    try{
      const p = await API.genReport(repKey, id, out, (meta.needsExtra ? (extraVal||'') : ''))
      setResult(s=>({...s, [repKey]: { path:p, ts: Date.now(), extra: extraVal||'' }}))
    }catch(e){
      setResult(s=>({...s, [repKey]: { error: e.message }}))
    }finally{
      setBusy(false)
    }
  }

  return (
    <div className="card">
      <div className="head">
        <b>Reportería</b>
        <span className="badge">Carousel</span>
        <div style={{marginLeft:'auto', display:'flex', gap:8}}>
          <input className="input mono" style={{maxWidth:200}} value={id} onChange={e=>setId(e.target.value)} placeholder="ID (ej. 841A)"/>
        </div>
      </div>

      <div className="body carousel">
        <div className="nav">
          <button className="btn-nav left"  onClick={()=>scrollBy(-1)}>{'<'}</button>
          <button className="btn-nav right" onClick={()=>scrollBy(1)}>{'>'}</button>
        </div>

        <div className="viewport" ref={viewport}>
          {REPORTS.map(r => (
            <ReportCard
              key={r.key}
              meta={r}
              id={id}
              out={defaults.out(r.key, id)}
              busy={busy}
              onGenerate={make}
              res={result[r.key]}
            />
          ))}
        </div>
      </div>
    </div>
  )
}

function ReportCard({meta, id, out, busy, onGenerate, res}){
  const [extra, setExtra] = useState('')
  const isImage = isImg(out)
  const filename = res?.path ? res.path.split('/').pop() : out.split('/').pop()
  return (
    <div className="report-card">
      <div className="hd">
        <b>{meta.title}</b>
        <span className="tag">{meta.kind}</span>
        {meta.needsExtra && <span className="tag">extra: {meta.extraLabel}</span>}
      </div>
      <div className="bd">
        <div className="desc">{meta.desc}</div>
        <div className="grid">
          <div className="muted">id</div>
          <div className="mono">{id}</div>
          <div className="muted">out</div>
          <div className="mono">{filename}</div>
        </div>

        {meta.needsExtra && (
          <div className="grid">
            <div className="muted">{meta.extraLabel}</div>
            <input className="input mono" value={extra} onChange={e=>setExtra(e.target.value)} placeholder={meta.key==='ls'?'/home/':'/home/a.txt'}/>
          </div>
        )}

        <div className="preview">
          {!res?.path && <small className="muted">Sin generar todavía.</small>}
          {res?.error && <div className="line err">ERR: {res.error}</div>}

          {res?.path && (isImage
            ? <img alt={meta.title}
                   src={`/reports/static/${encodeURIComponent(res.path.split('/').pop())}?t=${res.ts}`}
            />
            : <div>
                <div className="line ok">Generado en servidor:</div>
                <div className="mono">{res.path}</div>
                <small className="muted">Ábrelo desde la carpeta Reports del backend.</small>
              </div>
          )}
        </div>
      </div>
      <div className="ft">
        <button className="btn alt" disabled={busy} onClick={()=>onGenerate(meta.key, extra)}>
          {busy ? 'Generando…' : 'Generar'}
        </button>
      </div>
    </div>
  )
}
