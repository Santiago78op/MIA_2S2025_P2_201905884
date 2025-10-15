import { useEffect, useState } from 'react'
import { API } from '../lib/api'

export default function Explorer({id}){
  const [path,setPath]=useState('/')
  const [items,setItems]=useState([])
  const [viewFile,setViewFile]=useState(null) // {name, content}
  const [err,setErr]=useState('')
  const [loading,setLoading]=useState(false)

  useEffect(()=>{ load(path) },[id, path])

  async function load(p){
    setErr(''); setViewFile(null); setLoading(true)
    try{
      const d=await API.list(id, p)
      setItems(d.items||[])
    }catch(e){
      setErr(e.message)
    }finally{
      setLoading(false)
    }
  }

  const crumbs = path.split('/').filter(Boolean)
  function go(i){ const segs=['',...crumbs.slice(0,i+1)]; setPath(segs.join('/')||'/') }

  async function viewFileContent(fileName){
    try{
      const fullPath = path==='/' ? `/${fileName}` : `${path}/${fileName}`
      const txt = await API.file(id, fullPath)
      setViewFile({name:fileName, content:txt})
    }catch(e){
      alert('Error al leer archivo: ' + e.message)
    }
  }

  return (
    <div className="card">
      <div className="head">
        <b>Explorador de Archivos</b>
        <span className="badge mono">ID: {id}</span>
        <span className="badge">Solo Lectura</span>
      </div>
      <div className="body explorer">
        {/* Breadcrumb */}
        <div style={{marginBottom:10}}>
          <small className="muted" style={{display:'block', marginBottom:6}}>Ruta actual:</small>
          <div className="breadcrumb">
            <span className="cr" onClick={()=>setPath('/')}>/</span>
            {crumbs.map((c,i)=><span key={i} className="cr" onClick={()=>go(i)}>{c}</span>)}
          </div>
        </div>

        {err && <div className="line err">Error: {err}</div>}
        {loading && <div className="line sys">Cargando...</div>}

        {/* File/Directory List */}
        {!loading && !viewFile && (
          <>
            {items.length === 0 && <div className="muted">Carpeta vacía</div>}
            <div className="list">
              {items.map(x=>(
                <div key={x.name} className="item">
                  <div className="nm">{x.type==='dir' ? '📁' : '📄'} {x.name}</div>
                  <div className="perm mono">
                    {x.perm} · uid:{x.uid} · gid:{x.gid}
                    {x.size && ` · ${x.size} bytes`}
                  </div>
                  <div style={{marginTop:8, display:'flex', gap:6}}>
                    {x.type==='dir' ? (
                      <button className="btn alt" onClick={()=>setPath(path==='/'?`/${x.name}`:`${path}/${x.name}`)}>
                        Abrir Carpeta
                      </button>
                    ) : (
                      <button className="btn" onClick={()=>viewFileContent(x.name)}>
                        Ver Contenido
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </>
        )}

        {/* File Content Viewer */}
        {viewFile && (
          <div>
            <div style={{display:'flex', justifyContent:'space-between', alignItems:'center', marginBottom:10}}>
              <div>
                <b>Archivo: </b><span className="mono">{viewFile.name}</span>
              </div>
              <button className="btn alt" onClick={()=>setViewFile(null)}>Volver a lista</button>
            </div>
            <pre className="code mono" style={{maxHeight:'400px', overflow:'auto'}}>
              {viewFile.content || '(archivo vacío)'}
            </pre>
          </div>
        )}
      </div>
    </div>
  )
}
