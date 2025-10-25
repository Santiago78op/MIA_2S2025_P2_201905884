import { useState, useEffect } from 'react'
import { API } from '../lib/api'
import { LockIcon, UserIcon, ClockIcon, SizeIcon } from './Icons'
import folderImg from '../assets/icons/folder.png'
import fileImg from '../assets/icons/file.png'

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

  // Cargar automáticamente cuando cambia el path
  useEffect(() => {
    load()
  }, [path, id])

  const crumbs = path.split('/').filter(Boolean)
  function go(i){ const segs=['', ...crumbs.slice(0,i+1)]; setPath(segs.join('/')||'/') }

  // Detectar si el directorio actual es de solo lectura
  const currentDir = items.find(x => x.name === '.') || {}
  const isReadOnly = currentDir.perm ? !currentDir.perm.includes('2') : true

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
        <span className="badge" style={{background: isReadOnly ? 'var(--danger)' : 'var(--success)'}}>
          {isReadOnly ? '🔒 Solo Lectura' : '✓ Lectura/Escritura'}
        </span>
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

        {/* Mensaje inicial cuando no hay archivos */}
        {!loading && items.length === 0 && !viewFile && !err && (
          <div style={{
            textAlign:'center',
            padding:'40px 20px',
            background:'var(--panel2)',
            borderRadius:'10px'
          }}>
            <div className="muted" style={{marginBottom:'8px'}}>
              El directorio está vacío
            </div>
            <small className="muted">
              No hay archivos o carpetas en esta ubicación
            </small>
          </div>
        )}

        {/* File/Directory List */}
        {!loading && !viewFile && items.length > 0 && (
          <div className="list">
              {items.map(x=>{
                // Determinar icono según tipo
                const iconSrc = x.type === 'dir' ? folderImg : fileImg

                return (
                <div key={x.name} className="item">
                  <div style={{display:'flex', gap:'12px', alignItems:'start', marginBottom:'10px'}}>
                    <div style={{paddingTop:'2px'}}>
                      <img
                        src={iconSrc}
                        alt={x.type === 'dir' ? 'Carpeta' : 'Archivo'}
                        style={{
                          width: '40px',
                          height: '40px',
                          objectFit: 'contain',
                          filter: x.type === 'dir'
                            ? 'drop-shadow(0 0 4px rgba(255, 193, 7, 0.5))'
                            : 'drop-shadow(0 0 4px rgba(87, 182, 255, 0.5))'
                        }}
                      />
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
                            <span className="mono" style={{fontSize:'10px'}}>{new Date(x.mtime * 1000).toLocaleString('es-ES', {
                              year: 'numeric',
                              month: '2-digit',
                              day: '2-digit',
                              hour: '2-digit',
                              minute: '2-digit',
                              second: '2-digit'
                            })}</span>
                          </>
                        )}
                      </div>

                      <div style={{display:'flex', gap:6}}>
                        {x.type==='dir' ? (
                          <button className="btn alt" onClick={()=>{
                            const newPath = path==='/'?`/${x.name}`:`${path}/${x.name}`
                            setPath(newPath)
                          }} style={{width:'100%', display:'flex', alignItems:'center', justifyContent:'center', gap:'6px'}}>
                            <img src={folderImg} alt="Carpeta" style={{width: '16px', height: '16px'}} />
                            Abrir Carpeta
                          </button>
                        ) : (
                          <button className="btn" onClick={()=>viewFileContent(x.name)} style={{width:'100%', display:'flex', alignItems:'center', justifyContent:'center', gap:'6px'}}>
                            <img src={fileImg} alt="Archivo" style={{width: '16px', height: '16px'}} />
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
