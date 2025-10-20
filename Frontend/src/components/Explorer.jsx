import { useState } from 'react'
import { API } from '../lib/api'
import { FolderIcon, FileIcon, FileCodeIcon, FileBinaryIcon, LockIcon, UserIcon, ClockIcon, SizeIcon } from './Icons'

export default function Explorer({id, onBack}){
  const [path,setPath]=useState('/')
  const [items,setItems]=useState([])
  const [viewFile,setViewFile]=useState(null) // {name, content}
  const [err,setErr]=useState('')
  const [loading,setLoading]=useState(false)

  async function load(){
    setErr('')
    setViewFile(null)
    setLoading(true)
    try{
      const d = await API.list(id, path)
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
        <div style={{marginLeft:'auto', display:'flex', gap:'8px'}}>
          {onBack && (
            <button
              className="btn alt"
              onClick={onBack}
              style={{fontSize:'12px', padding:'6px 12px'}}
            >
              ← Volver a Particiones
            </button>
          )}
          <button
            className="btn"
            onClick={load}
            disabled={loading}
            style={{fontSize:'12px', padding:'6px 12px'}}
          >
            {loading ? 'Actualizando...' : 'Cargar / Actualizar'}
          </button>
        </div>
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

        {/* Mensaje inicial cuando no se ha cargado nada */}
        {!loading && items.length === 0 && !viewFile && !err && (
          <div style={{
            textAlign:'center',
            padding:'40px 20px',
            background:'var(--panel2)',
            borderRadius:'10px'
          }}>
            <div className="muted" style={{marginBottom:'8px'}}>
              Haz clic en "Cargar / Actualizar" para explorar el sistema de archivos
            </div>
            <small className="muted">
              Comenzarás desde la raíz (/)
            </small>
          </div>
        )}

        {/* File/Directory List */}
        {!loading && !viewFile && items.length > 0 && (
          <div className="list">
              {items.map(x=>{
                // Determinar icono según tipo y nombre
                let ItemIcon = FileIcon
                if (x.type === 'dir') {
                  ItemIcon = FolderIcon
                } else if (x.name.match(/\.(sh|py|js|jsx|ts|tsx|go|c|cpp|java)$/i)) {
                  ItemIcon = FileCodeIcon
                } else if (x.name.match(/\.(bin|exe|out|o)$/i)) {
                  ItemIcon = FileBinaryIcon
                }

                return (
                <div key={x.name} className="item">
                  <div style={{display:'flex', gap:'12px', alignItems:'start', marginBottom:'10px'}}>
                    <div style={{paddingTop:'2px'}}>
                      <ItemIcon size={36} color={x.type==='dir' ? 'var(--warning)' : 'var(--neo2)'} />
                    </div>
                    <div style={{flex:1}}>
                      <div style={{display:'flex', alignItems:'center', gap:'8px', marginBottom:'6px'}}>
                        <div className="nm">{x.name}</div>
                        <span className="badge" style={{
                          background: x.type==='dir' ? 'var(--warning)' : 'var(--info)',
                          borderColor: x.type==='dir' ? 'var(--warning)' : 'var(--info)',
                          fontSize:'10px',
                          padding:'2px 6px'
                        }}>
                          {x.type==='dir' ? 'DIR' : 'FILE'}
                        </span>
                      </div>

                      <div style={{display:'grid', gridTemplateColumns:'auto 1fr', gap:'8px 12px', fontSize:'11px', marginBottom:'8px'}}>
                        <div style={{display:'flex', alignItems:'center', gap:'4px'}}>
                          <LockIcon size={12} color="var(--muted)" />
                          <span className="muted">Permisos:</span>
                        </div>
                        <span className="mono">{x.perm || 'N/A'}</span>

                        <div style={{display:'flex', alignItems:'center', gap:'4px'}}>
                          <UserIcon size={12} color="var(--muted)" />
                          <span className="muted">Owner:</span>
                        </div>
                        <span className="mono">{x.owner || 'N/A'}:{x.group || 'N/A'}</span>

                        {x.size > 0 && (
                          <>
                            <div style={{display:'flex', alignItems:'center', gap:'4px'}}>
                              <SizeIcon size={12} color="var(--muted)" />
                              <span className="muted">Tamaño:</span>
                            </div>
                            <span className="mono">{x.size} bytes</span>
                          </>
                        )}

                        {x.mtime && (
                          <>
                            <div style={{display:'flex', alignItems:'center', gap:'4px'}}>
                              <ClockIcon size={12} color="var(--muted)" />
                              <span className="muted">Modificado:</span>
                            </div>
                            <span className="mono" style={{fontSize:'10px'}}>{new Date(x.mtime).toLocaleString()}</span>
                          </>
                        )}
                      </div>

                      <div style={{display:'flex', gap:6}}>
                        {x.type==='dir' ? (
                          <button className="btn alt" onClick={()=>setPath(path==='/'?`/${x.name}`:`${path}/${x.name}`)} style={{width:'100%', display:'flex', alignItems:'center', justifyContent:'center', gap:'6px'}}>
                            <FolderIcon size={16} color="currentColor" />
                            Abrir Carpeta
                          </button>
                        ) : (
                          <button className="btn" onClick={()=>viewFileContent(x.name)} style={{width:'100%', display:'flex', alignItems:'center', justifyContent:'center', gap:'6px'}}>
                            <FileIcon size={16} color="currentColor" />
                            Ver Contenido
                          </button>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              )})}
          </div>
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
