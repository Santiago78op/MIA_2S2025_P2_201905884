import { useState, useEffect, useCallback } from 'react'
import { API } from '../lib/api'
import ImageLightbox from './ImageLightbox'

export default function ReportsGallery(){
  const [files, setFiles] = useState([])
  const [loading, setLoading] = useState(false)
  const [filter, setFilter] = useState('all') // all | images | text
  const [active, setActive] = useState(0)
  const [lb, setLb] = useState({open:false, src:'', title:''})

  async function loadFiles(){
    setLoading(true)
    try{
      const data = await API.listReports()
      setFiles(data.files || [])
    }catch(e){
      console.error('Error cargando archivos:', e)
    }finally{
      setLoading(false)
    }
  }

  // Filtrar archivos según el filtro activo
  const filteredFiles = files.filter(f => {
    if (filter === 'all') return true
    if (filter === 'images') return f.is_image
    if (filter === 'text') return !f.is_image
    return true
  })

  // Ajustar índice activo si el filtro reduce la lista
  useEffect(()=>{
    if (active > filteredFiles.length-1 && filteredFiles.length > 0) setActive(0)
  }, [filter, active, filteredFiles.length])

  // Funciones de navegación con useCallback
  const next = useCallback(() => {
    if(filteredFiles.length) setActive(i => (i+1)%filteredFiles.length)
  }, [filteredFiles.length])

  const prev = useCallback(() => {
    if(filteredFiles.length) setActive(i => (i-1+filteredFiles.length)%filteredFiles.length)
  }, [filteredFiles.length])

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
  }, [next, prev])

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

  function formatBytes(bytes) {
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024*1024) return (bytes/1024).toFixed(1) + ' KB'
    return (bytes/(1024*1024)).toFixed(1) + ' MB'
  }

  function formatDate(dateStr) {
    const d = new Date(dateStr)
    return d.toLocaleString('es-ES', {
      year: 'numeric', month: '2-digit', day: '2-digit',
      hour: '2-digit', minute: '2-digit'
    })
  }

  return (
    <div className="card">
      <div className="head">
        <b>Galería de Reportes</b>
        <span className="badge">{filteredFiles.length} archivo{filteredFiles.length !== 1 ? 's' : ''}</span>
        <div style={{marginLeft:'auto'}} className="toolbar">
          <div className="pills">
            <span className={`pill ${filter==='all'?'active':''}`}   onClick={()=>setFilter('all')}>Todos ({files.length})</span>
            <span className={`pill ${filter==='images'?'active':''}`} onClick={()=>setFilter('images')}>Imágenes ({files.filter(f=>f.is_image).length})</span>
            <span className={`pill ${filter==='text'?'active':''}`}  onClick={()=>setFilter('text')}>Texto ({files.filter(f=>!f.is_image).length})</span>
          </div>
          <button className="btn" onClick={loadFiles} disabled={loading}>
            {loading ? 'Cargando...' : 'Recargar'}
          </button>
        </div>
      </div>

      {filteredFiles.length === 0 && (
        <div className="body" style={{textAlign:'center', padding:'40px'}}>
          <div className="muted">
            {loading ? 'Cargando archivos...' : files.length === 0 ? 'Haz clic en "Recargar" para cargar los reportes' : 'No hay archivos que coincidan con el filtro'}
          </div>
          {!loading && files.length === 0 && (
            <small className="muted" style={{display:'block', marginTop:'8px'}}>
              Genera reportes desde la terminal o desde la pestaña de reportería
            </small>
          )}
        </div>
      )}

      {filteredFiles.length > 0 && (
        <div className="body cflow" onWheel={(e)=>{ if(Math.abs(e.deltaY)>Math.abs(e.deltaX)) (e.deltaY>0?next():prev()) }}>
          <div className="nav3d">
            <button className="btn3d left" onClick={prev} title="Anterior">&lt;</button>
            <button className="btn3d right" onClick={next} title="Siguiente">&gt;</button>
          </div>

          <div className="rail">
            <div className="cards">
              {filteredFiles.map((file, i) => (
                <div key={file.name} className="card3d" style={posStyle(i)}>
                  <div className="hd">
                    <b>{file.name}</b>
                    <span className="badge">{file.ext.toUpperCase().substring(1)}</span>
                  </div>
                  <div className="bd">
                    <div style={{display:'grid', gridTemplateColumns:'80px 1fr', gap:'6px 10px', fontSize:'13px', marginBottom:'12px'}}>
                      <div className="muted">Tamaño:</div>
                      <div className="mono">{formatBytes(file.size)}</div>
                      <div className="muted">Modificado:</div>
                      <div style={{fontSize:'11px'}}>{formatDate(file.mod_time)}</div>
                    </div>

                    <div className="preview3d">
                      {file.is_image ? (
                        <img
                          alt={file.name}
                          src={file.url + '?t=' + new Date(file.mod_time).getTime()}
                          onClick={()=>setLb({
                            open:true,
                            src: file.url + '?t=' + new Date(file.mod_time).getTime(),
                            title: file.name
                          })}
                          style={{cursor:'pointer'}}
                        />
                      ) : (
                        <div style={{
                          display:'flex',
                          flexDirection:'column',
                          gap:'8px',
                          justifyContent:'center',
                          alignItems:'center',
                          height:'100%',
                          padding:'20px'
                        }}>
                          <div style={{fontSize:'48px', opacity:0.3}}>📄</div>
                          <div className="mono" style={{fontSize:'11px', wordBreak:'break-all', textAlign:'center'}}>
                            {file.name}
                          </div>
                          <small className="muted" style={{fontSize:'11px'}}>
                            Archivo de texto - Abre desde la carpeta Reports
                          </small>
                          <a
                            href={file.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="btn alt"
                            style={{fontSize:'12px', padding:'6px 12px', marginTop:'8px'}}
                          >
                            Descargar
                          </a>
                        </div>
                      )}
                    </div>

                    {file.is_image && (
                      <div className="footer">
                        <a
                          href={file.url}
                          download={file.name}
                          className="btn alt"
                          style={{width:'100%', fontSize:'13px', padding:'8px 12px', textAlign:'center', display:'block', textDecoration:'none'}}
                        >
                          Descargar Imagen
                        </a>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>

            <div className="dots" style={{marginTop:8}}>
              {filteredFiles.map((_,i)=><span key={i} className={`dot3d ${i===active?'active':''}`} onClick={()=>setActive(i)} />)}
            </div>
          </div>
        </div>
      )}

      <ImageLightbox
        open={lb.open} src={lb.src} title={lb.title}
        onClose={()=>setLb({open:false, src:'', title:''})}
      />
    </div>
  )
}
