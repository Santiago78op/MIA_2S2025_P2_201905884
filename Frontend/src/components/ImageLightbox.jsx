import { useState } from 'react'

export default function ImageLightbox({open, src, title, onClose}) {
  const [scale,setScale] = useState(1)
  if(!open) return null
  return (
    <div className="lb" onClick={onClose}>
      <div className="box" onClick={e=>e.stopPropagation()}>
        <div className="top">
          <b>{title||'Reporte'}</b>
          <div className="zoom">
            <button className="btn alt" onClick={()=>setScale(s=>Math.max(0.25,s-0.25))}>-</button>
            <button className="btn alt" onClick={()=>setScale(1)}>100%</button>
            <button className="btn alt" onClick={()=>setScale(s=>Math.min(5,s+0.25))}>+</button>
            <button className="btn" onClick={onClose}>Cerrar</button>
          </div>
        </div>
        <div className="area">
          <img
            alt={title} src={src}
            style={{ transform:`scale(${scale})`, transformOrigin:'center center' }}
          />
        </div>
      </div>
    </div>
  )
}
