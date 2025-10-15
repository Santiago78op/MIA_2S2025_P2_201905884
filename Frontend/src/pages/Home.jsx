import { useNavigate } from 'react-router-dom'
import Terminal from '../components/Terminal'

export default function Home({session}){
  const navigate = useNavigate()

  return (
    <div style={{padding:'12px', minHeight:'calc(100vh - 60px)', display:'flex', flexDirection:'column', gap:'12px'}}>
      {/* Action Bar */}
      <div className="card">
        <div className="head" style={{flexWrap:'wrap'}}>
          <b>MIA - Manejo e Implementación de Archivos</b>
          <span className="badge">Proyecto 2</span>
          <div style={{marginLeft:'auto', display:'flex', gap:8, flexWrap:'wrap'}}>
            {!session && (
              <button className="btn" onClick={()=>navigate('/login')}>
                🔑 Iniciar Sesión
              </button>
            )}
            <button className="btn alt" onClick={()=>navigate('/visualizer')}>
              📁 Visualizador
            </button>
          </div>
        </div>
        {session && (
          <div className="body" style={{padding:'8px 12px', background:'var(--panel2)'}}>
            <div style={{display:'flex', gap:'12px', alignItems:'center', flexWrap:'wrap'}}>
              <span className="badge">✓ Sesión: {session.user}</span>
              <span className="badge">ID: {session.id}</span>
              <small className="muted">
                Los comandos que requieren sesión están habilitados
              </small>
            </div>
          </div>
        )}
      </div>

      {/* Terminal */}
      <div className="card" style={{flex:1, minHeight:'600px'}}>
        <div className="head">
          <b>Terminal de Comandos</b>
          <span className="badge">Entrada y Salida</span>
        </div>
        <div className="body" style={{height:'calc(100% - 50px)'}}>
          <Terminal session={session}/>
        </div>
      </div>

      {/* Info Card */}
      <div className="card">
        <div className="head">
          <b>Información y Comandos Disponibles</b>
        </div>
        <div className="body">
          <div className="grid cols-2" style={{gap:'12px'}}>
            <div>
              <h4 style={{margin:'0 0 8px 0', color:'var(--neon)'}}>Gestión de Discos y Particiones</h4>
              <div className="mono" style={{fontSize:'12px', lineHeight:'1.8'}}>
                mkdisk, rmdisk, fdisk, mount, unmount
              </div>
            </div>
            <div>
              <h4 style={{margin:'0 0 8px 0', color:'var(--neon)'}}>Sistema de Archivos</h4>
              <div className="mono" style={{fontSize:'12px', lineHeight:'1.8'}}>
                mkfs, login, logout, mkgrp, rmgrp, mkusr, rmusr
              </div>
            </div>
            <div>
              <h4 style={{margin:'0 0 8px 0', color:'var(--neo2)'}}>Operaciones de Archivos</h4>
              <div className="mono" style={{fontSize:'12px', lineHeight:'1.8'}}>
                mkdir, mkfile, remove, edit, rename, copy, move, cat
              </div>
            </div>
            <div>
              <h4 style={{margin:'0 0 8px 0', color:'var(--neo2)'}}>Permisos y Reportes</h4>
              <div className="mono" style={{fontSize:'12px', lineHeight:'1.8'}}>
                chmod, chown, chgrp, find, recovery, loss, rep, exec
              </div>
            </div>
          </div>
          <hr style={{border:'none', borderTop:'1px solid var(--border)', margin:'12px 0'}}/>
          <small className="muted">
            💡 <b>Tip:</b> Use el botón "Cargar Archivo" para ejecutar scripts .smia |
            Presione Ctrl+Enter en el área de entrada para ejecutar todos los comandos
          </small>
        </div>
      </div>
    </div>
  )
}
