import { useState } from 'react'
import { API } from '../lib/api'
import { ServerIcon } from './Icons'
import discoImg from '../assets/icons/disco.png'

export default function DiskPicker({onSelect}){
  const [disks,setDisks]=useState([])
  const [err,setErr]=useState('')
  const [loading,setLoading]=useState(false)

  async function loadDisks(){
    setLoading(true)
    setErr('')
    try {
      const data = await API.disks()
      setDisks(data)
    } catch(e) {
      setErr(e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="card">
      <div className="head">
        <b>Paso 1: Selección de Disco</b>
        <span className="badge">Discos Disponibles</span>
        <button
          className="btn"
          onClick={loadDisks}
          disabled={loading}
          style={{marginLeft:'auto', fontSize:'12px', padding:'6px 12px'}}
        >
          {loading ? 'Cargando...' : 'Cargar Discos'}
        </button>
      </div>
      <div className="body">
        {err && (
          <div style={{
            padding:'10px',
            background:'var(--panel2)',
            border:'1px solid var(--danger)',
            borderRadius:'10px',
            color:'var(--danger)',
            marginBottom:'12px'
          }}>
            <b>Error:</b> {err}
          </div>
        )}

        {!loading && disks.length === 0 && !err && (
          <div style={{
            textAlign:'center',
            padding:'40px 20px',
            background:'var(--panel2)',
            borderRadius:'10px'
          }}>
            <div className="muted" style={{marginBottom:'8px'}}>
              Haz clic en "Cargar Discos" para ver los discos disponibles
            </div>
            <small className="muted">
              Asegúrate de haber creado y montado discos usando la terminal
            </small>
          </div>
        )}

        {loading && (
          <div className="muted" style={{textAlign:'center', padding:'20px'}}>
            Cargando discos...
          </div>
        )}

        {!loading && disks.length > 0 && (
          <div className="list">
            {disks.map(d=>(
              <div key={d.path} className="disk-card-wrapper">
                <div className="disk-top-section">
                  <div className="disk-border"></div>
                  <div className="disk-logo-center">
                    <img src={discoImg} alt="Disco" />
                  </div>
                </div>
                <div className="disk-bottom-section">
                  <span className="disk-title">{d.name || 'Disco sin nombre'}</span>
                  <div className="disk-row">
                    <div className="disk-item">
                      <span className="disk-big-text">{d.size}</span>
                      <span className="disk-regular-text">Capacidad</span>
                    </div>
                    <div className="disk-item">
                      <span className="disk-big-text">{d.fit}</span>
                      <span className="disk-regular-text">Ajuste</span>
                    </div>
                    <div className="disk-item">
                      <span className="disk-big-text">{d.mounted?.length || 0}</span>
                      <span className="disk-regular-text">Montadas</span>
                    </div>
                  </div>
                  <button
                    className="btn"
                    onClick={()=>onSelect(d)}
                    style={{
                      width:'100%',
                      marginTop: '15px',
                      padding: '12px',
                      fontSize: '14px',
                      fontWeight: '600',
                      textTransform: 'uppercase',
                      letterSpacing: '1px'
                    }}
                  >
                    Seleccionar Disco
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
