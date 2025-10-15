import { useEffect, useState } from 'react'
import { API } from '../lib/api'

export default function PartitionPicker({disk, onSelect}){
  const [parts,setParts]=useState([])
  useEffect(()=>{ if(disk) API.partitions(disk.path).then(setParts) },[disk])
  if(!disk) return null
  return (
    <div className="card">
      <div className="head"><b>Particiones</b><span className="badge">{disk.path}</span></div>
      <div className="body list">
        {parts.map(p=>(
          <div key={p.name} className="item">
            <div className="nm">{p.name}</div>
            <div className="tag">Tipo: {p.type} · Fit: {p.fit}</div>
            <div className="tag">Tamaño: {p.size}</div>
            <button className="btn" onClick={()=>onSelect(p)}>Montar / Explorar</button>
          </div>
        ))}
        {parts.length===0 && <div className="muted">Sin particiones detectadas.</div>}
      </div>
    </div>
  )
}
