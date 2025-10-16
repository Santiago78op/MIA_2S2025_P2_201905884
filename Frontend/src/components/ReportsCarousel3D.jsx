import { useMemo, useState, useEffect } from 'react'
import { API } from '../lib/api'
import ImageLightbox from './ImageLightbox'

const REPORTS = [
  { key:'mbr',      title:'MBR',        kind:'image', desc:'Estructura del MBR y particiones.' },
  { key:'disk',     title:'Disk',       kind:'image', desc:'Mapa visual de ocupación.' },
  { key:'inode',    title:'Inode',      kind:'image', desc:'Detalle de un inodo y apuntadores.' },
  { key:'block',    title:'Block',      kind:'image', desc:'Bloques de archivo/carpeta.' },
  { key:'tree',     title:'Tree',       kind:'image', desc:'Árbol de directorios y archivos.' },
  { key:'file',     title:'File',       kind:'image', desc:'Vista de un archivo.', needsExtra:true, extraLabel:'path_file', placeholder:'/home/a.txt' },
  { key:'ls',       title:'LS',         kind:'image', desc:'Listado de un directorio.', needsExtra:true, extraLabel:'path_file_ls', placeholder:'/home' },
  { key:'bm_inode', title:'BM Inode',   kind:'text',  desc:'Bitmap de inodos (20 bits por línea).' },
  { key:'bm_block', title:'BM Block',   kind:'text',  desc:'Bitmap de bloques (20 bits por línea).' },
  { key:'sb',       title:'SuperBlock', kind:'text',  desc:'Campos y offsets del SuperBlock.' },
]

const IMG_EXT = new Set(['.png','.jpg','.jpeg','.svg','.gif'])
const isImg = (p) => IMG_EXT.has((p?.slice(p.lastIndexOf('.'))||'').toLowerCase())

export default function ReportsCarousel3D(){
  const [id, setId] = useState('841A')
  const [filter, setFilter] = useState('all') // all | image | text
  const [active, setActive] = useState(0)
  const [busy, setBusy] = useState(false)
  const [result, setResult] = useState({})  // key -> { path, ts, error }
  const [extras, setExtras] = useState(() => (JSON.parse(localStorage.getItem('rep_extras')||'{}')))
  const [lb, setLb] = useState({open:false, src:'', title:''})
  const [progress, setProgress] = useState(0)

  const defaults = useMemo(()=>({
    out(name, id) {
      const base = './Reports'
      if (name==='bm_inode') return `${base}/bm_inode_${id}.txt`
      if (name==='bm_block') return `${base}/bm_block_${id}.txt`
      if (name==='sb')       return `${base}/sb_${id}.txt`
      return `${base}/${name}_${id}.png`
    }
  }),[])

  const items = REPORTS.filter(r => filter==='all' ? true : r.kind===filter)
  useEffect(()=>{ if (active > items.length-1) setActive(0) }, [filter, active, items.length]) // reajusta índice

  // Navegación por teclado
  useEffect(() => {
    const handleKeyDown = (e) => {
      if (e.key === 'ArrowLeft') {
        e.preventDefault()
        prev()
      } else if (e.key === 'ArrowRight') {
        e.preventDefault()
        next()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [active, items.length])

  function posStyle(i){
    const off = i - active

    // Card activa (center)
    if (off === 0) {
      return {
        transform: 'translate(-50%, -50%) translateX(0) scale(1)',
        opacity: 1,
        zIndex: 100,
        visibility: 'visible'
      }
    }

    // Card siguiente (right)
    if (off === 1) {
      return {
        transform: 'translate(-50%, -50%) translateX(100%) scale(0.85)',
        opacity: 0.7,
        zIndex: 90,
        visibility: 'visible'
      }
    }

    // Card anterior (left)
    if (off === -1) {
      return {
        transform: 'translate(-50%, -50%) translateX(-100%) scale(0.85)',
        opacity: 0.7,
        zIndex: 90,
        visibility: 'visible'
      }
    }

    // Cards muy alejadas (ocultas)
    return {
      transform: `translate(-50%, -50%) translateX(${off > 0 ? '200%' : '-200%'}) scale(0.7)`,
      opacity: 0,
      zIndex: 50,
      visibility: 'hidden'
    }
  }

  function next(){ setActive(i => (i+1)%items.length) }
  function prev(){ setActive(i => (i-1+items.length)%items.length) }

  async function generateOne(repKey){
    const meta = REPORTS.find(r=>r.key===repKey); if(!meta) return
    const out = defaults.out(repKey, id)
    try{
      const p = await API.genReport(repKey, id, out, meta.needsExtra ? (extras[repKey]||'') : '')
      setResult(s=>({...s, [repKey]: { path:p, ts: Date.now() }}))
      return true
    }catch(e){
      setResult(s=>({...s, [repKey]: { error: e.message }}))
      return false
    }
  }

  async function generateAll(){
    setBusy(true); setProgress(0)
    const keys = items.map(x=>x.key)
    for(let i=0; i<keys.length; i++){
      // si requiere extra y no lo tiene, lo salta con error
      const meta = REPORTS.find(r=>r.key===keys[i])
      if (meta?.needsExtra && !(extras[keys[i]]||'').trim()) {
        setResult(s=>({...s, [keys[i]]:{error:'Falta extra'}}))
      } else {
        await generateOne(keys[i])
      }
      setProgress((i+1)/keys.length)
    }
    setBusy(false)
  }

  function setExtraFor(k, v){
    setExtras(s=>{
      const n = {...s, [k]: v}
      localStorage.setItem('rep_extras', JSON.stringify(n))
      return n
    })
  }

  return (
    <div className="card">
      <div className="head">
        <b>Reportería Coverflow 3D</b>
        <span className="badge">Navegación interactiva</span>
        <div style={{marginLeft:'auto'}} className="toolbar">
          <div className="pills">
            <span className={`pill ${filter==='all'?'active':''}`}   onClick={()=>setFilter('all')}>Todos</span>
            <span className={`pill ${filter==='image'?'active':''}`} onClick={()=>setFilter('image')}>Imágenes</span>
            <span className={`pill ${filter==='text'?'active':''}`}  onClick={()=>setFilter('text')}>Texto</span>
          </div>
          <input className="input mono" style={{maxWidth:200}} value={id} onChange={e=>setId(e.target.value)} placeholder="ID (ej. 841A)"/>
          <button className="btn" onClick={generateAll} disabled={busy}>Generar todos</button>
        </div>
      </div>

      {busy && (
        <div style={{padding:'8px 12px'}}>
          <div className="progress"><span style={{width:`${Math.round(progress*100)}%`}}/></div>
          <small className="muted">{Math.round(progress*100)}%</small>
        </div>
      )}

      <div className="body cflow" onWheel={(e)=>{ if(Math.abs(e.deltaY)>Math.abs(e.deltaX)) (e.deltaY>0?next():prev()) }}>
        <div className="nav3d">
          <button className="btn3d left" onClick={prev} title="Anterior">&lt;</button>
          <button className="btn3d right" onClick={next} title="Siguiente">&gt;</button>
        </div>

        <div className="rail">
          <div className="cards">
            {items.map((meta, i) => {
              const res = result[meta.key]
              const out = defaults.out(meta.key, id)
              const filename = res?.path ? res.path.split('/').pop() : out.split('/').pop()
              const needsExtra = meta.needsExtra && !(extras[meta.key]||'').trim()
              return (
                <div key={meta.key} className="card3d" style={posStyle(i)}>
                  <div className="hd">
                    <b>{meta.title}</b>
                    <span className="badge">{meta.kind}</span>
                    {meta.needsExtra && <span className={`badge ${needsExtra?'warn':''}`}>{meta.extraLabel}</span>}
                  </div>
                  <div className="bd">
                    <div className="muted" style={{fontSize:'13px', lineHeight:'1.4'}}>{meta.desc}</div>

                    <div style={{display:'grid', gridTemplateColumns:'80px 1fr', gap:'6px 10px', fontSize:'13px'}}>
                      <div className="muted">ID:</div>
                      <div className="mono">{id}</div>
                      <div className="muted">Output:</div>
                      <div className="mono" style={{fontSize:'11px', wordBreak:'break-all'}}>{filename}</div>
                    </div>

                    {meta.needsExtra && (
                      <div style={{display:'flex', flexDirection:'column', gap:'4px'}}>
                        <label className="muted" style={{fontSize:'12px'}}>{meta.extraLabel}:</label>
                        <input
                          className="input mono"
                          placeholder={meta.placeholder||'/ruta'}
                          value={extras[meta.key]||''}
                          onChange={e=>setExtraFor(meta.key, e.target.value)}
                          style={{fontSize:'12px', padding:'6px 8px'}}
                        />
                      </div>
                    )}

                    <div className="preview3d">
                      {!res?.path && !res?.error && (
                        <div style={{display:'flex', alignItems:'center', justifyContent:'center', height:'100%'}}>
                          <small className="muted">Presiona "Generar" para crear el reporte</small>
                        </div>
                      )}
                      {res?.error && (
                        <div style={{display:'flex', alignItems:'center', justifyContent:'center', height:'100%', padding:'10px'}}>
                          <div className="line err" style={{textAlign:'center'}}>ERROR: {res.error}</div>
                        </div>
                      )}
                      {res?.path && (isImg(res.path)
                        ? <img
                            alt={meta.title}
                            src={`/reports/static/${encodeURIComponent(res.path.split('/').pop())}?t=${res.ts}`}
                            onClick={()=>setLb({open:true, src:`/reports/static/${encodeURIComponent(res.path.split('/').pop())}?t=${res.ts}`, title: meta.title})}
                          />
                        : <div style={{display:'flex', flexDirection:'column', gap:'6px', justifyContent:'center', height:'100%', padding:'10px'}}>
                            <div className="line ok">Archivo generado exitosamente</div>
                            <div className="mono" style={{fontSize:'11px', wordBreak:'break-all'}}>{res.path}</div>
                            <small className="muted" style={{fontSize:'11px'}}>Abre el archivo desde la carpeta Reports del backend</small>
                          </div>
                      )}
                    </div>

                    <div className="footer">
                      <button
                        className="btn alt"
                        disabled={busy || needsExtra}
                        onClick={()=>generateOne(meta.key)}
                        style={{width:'100%', fontSize:'13px', padding:'8px 12px'}}
                      >
                        {busy ? 'Generando…' : 'Generar Reporte'}
                      </button>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>

          <div className="dots" style={{marginTop:8}}>
            {items.map((_,i)=><span key={i} className={`dot3d ${i===active?'active':''}`} onClick={()=>setActive(i)} />)}
          </div>
        </div>
      </div>

      <ImageLightbox
        open={lb.open} src={lb.src} title={lb.title}
        onClose={()=>setLb({open:false, src:'', title:''})}
      />
    </div>
  )
}
