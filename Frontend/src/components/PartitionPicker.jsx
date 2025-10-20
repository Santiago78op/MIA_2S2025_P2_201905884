import { useState } from 'react'
import { API } from '../lib/api'
import { PartitionIcon, MountIcon, CheckIcon } from './Icons'
import particionImg from '../assets/icons/particion.png'

export default function PartitionPicker({disk, onSelect, onBack}){
  const [parts,setParts]=useState([])
  const [loading,setLoading]=useState(false)
  const [err,setErr]=useState('')

  async function loadPartitions(){
    if(!disk) return
    setLoading(true)
    setErr('')
    try {
      const data = await API.partitions(disk.path)
      setParts(data)
    } catch(e) {
      setErr(e.message)
    } finally {
      setLoading(false)
    }
  }

  if(!disk) return null

  return (
    <div className="card">
      <div className="head">
        <b>Paso 2: Selección de Partición</b>
        <span className="badge">{disk.name}</span>
        <div style={{marginLeft:'auto', display:'flex', gap:'8px'}}>
          {onBack && (
            <button
              className="btn alt"
              onClick={onBack}
              style={{fontSize:'12px', padding:'6px 12px'}}
            >
              ← Volver a Discos
            </button>
          )}
          <button
            className="btn"
            onClick={loadPartitions}
            disabled={loading}
            style={{fontSize:'12px', padding:'6px 12px'}}
          >
            {loading ? 'Cargando...' : 'Cargar Particiones'}
          </button>
        </div>
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

        {!loading && parts.length === 0 && !err && (
          <div style={{
            textAlign:'center',
            padding:'40px 20px',
            background:'var(--panel2)',
            borderRadius:'10px'
          }}>
            <div className="muted" style={{marginBottom:'8px'}}>
              Haz clic en "Cargar Particiones" para ver las particiones de este disco
            </div>
            <small className="muted">
              Monta una partición usando el comando mount en la terminal
            </small>
          </div>
        )}

        {loading && (
          <div className="muted" style={{textAlign:'center', padding:'20px'}}>
            Cargando particiones...
          </div>
        )}

        {!loading && parts.length > 0 && (
          <div className="list">
            {parts.map(p=>(
              <div key={p.id || p.name} className="partition-card">
                <div className="partition-content">
                  <div className="partition-back">
                    <div className="partition-back-content">
                      <img src={particionImg} alt="Partición" style={{width: '60px', height: '60px', filter: 'drop-shadow(0 0 10px rgba(255,255,255,0.5))'}} />
                      <strong>Hover para info</strong>
                    </div>
                  </div>
                  <div className="partition-front">
                    <div className="partition-img">
                      <div className="partition-circle"></div>
                      <div className="partition-circle partition-right"></div>
                      <div className="partition-circle partition-bottom"></div>
                    </div>
                    <div className="partition-front-content">
                      <small className="partition-badge">{p.type || 'Partición'}</small>
                      <div className="partition-description">
                        <div className="partition-title-section">
                          <p className="partition-title-text">
                            <strong>{p.name}</strong>
                          </p>
                          {p.formatted && <CheckIcon size={15} color="var(--neon)" />}
                        </div>
                        <div className="partition-details">
                          <div className="partition-detail-row">
                            <span className="partition-label">Tamaño:</span>
                            <span className="partition-value">{p.size}</span>
                          </div>
                          <div className="partition-detail-row">
                            <span className="partition-label">Ajuste:</span>
                            <span className="partition-value">{p.fit}</span>
                          </div>
                          {p.id && (
                            <div className="partition-detail-row">
                              <span className="partition-label">ID:</span>
                              <span className="partition-value">{p.id}</span>
                            </div>
                          )}
                        </div>
                        <button
                          className="partition-explore-btn"
                          onClick={()=>onSelect(p)}
                        >
                          Explorar FS
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
