import { useEffect, useState } from 'react'
import { API } from '../lib/api'

export default function DiskPicker({onSelect}){
  const [disks,setDisks]=useState([])
  const [err,setErr]=useState('')
  useEffect(()=>{ API.disks().then(setDisks).catch(e=>setErr(e.message)) },[])
  return (
    <div className="card">
      <div className="head"><b>Discos</b><span className="badge">selección</span></div>
      <div className="body">
        {err && <div className="line err">{err}</div>}
        <div className="list">
          {disks.map(d=>(
            <div key={d.path} className="item">
              <div className="nm mono">{d.path}</div>
              <div className="tag">Capacidad: {d.size}</div>
              <div className="tag">Fit: {d.fit} · Montadas: {d.mounted?.length||0}</div>
              <div style={{marginTop:8, display:'flex', gap:8}}>
                <button className="btn" onClick={()=>onSelect(d)}>Explorar</button>
              </div>
            </div>
          ))}
          {disks.length===0 && <div className="muted">No hay discos aún. Crea uno en la terminal.</div>}
        </div>
      </div>
    </div>
  )
}
