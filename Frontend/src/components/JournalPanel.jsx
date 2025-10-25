import { useState } from 'react'
import { API } from '../lib/api'

export default function JournalPanel({id}){
  const [rows,setRows]=useState([])
  const [loading,setLoading]=useState(false)
  const [err,setErr]=useState('')

  async function loadJournal(){
    setLoading(true)
    setErr('')
    try {
      const data = await API.journaling(id)
      setRows(data.rows || [])
    } catch(e) {
      // Mejorar mensaje de error para journal no disponible
      let errorMsg = e.message || 'Error cargando journal'
      if (errorMsg.includes('journal') || errorMsg.includes('EXT2') || errorMsg.includes('no disponible')) {
        errorMsg = 'Journal no disponible. Esta partición debe ser formateada con EXT3 (mkfs -type=3fs) para usar journaling.'
      }
      setErr(errorMsg)
      setRows([])
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="card">
      <div className="head">
        <b>Journaling</b>
        <span className="badge">EXT3</span>
        <button
          className="btn"
          onClick={loadJournal}
          disabled={loading}
          style={{marginLeft:'auto', fontSize:'12px', padding:'6px 12px'}}
        >
          {loading ? 'Cargando...' : 'Cargar Journal'}
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
            marginBottom:'12px',
            fontSize:'12px'
          }}>
            <b>Error:</b> {err}
          </div>
        )}
        {loading && <div className="muted">Cargando journal...</div>}
        {!loading && rows.length===0 && !err && (
          <div className="muted">
            Haz clic en "Cargar Journal" para ver las transacciones (solo EXT3)
          </div>
        )}
        {!loading && rows.length > 0 && rows.map((j,i)=>(
          <div key={i} className="journalRow">
            <div className="op">{j.operation}</div>
            <div className="mono">{j.path} {j.extra?`→ ${j.extra}`:''}</div>
            <div className="muted mono">{new Date(j.time*1000).toLocaleString()}</div>
          </div>
        ))}
      </div>
    </div>
  )
}
