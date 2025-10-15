import { useEffect, useState } from 'react'
import { API } from '../lib/api'

export default function JournalPanel({id}){
  const [rows,setRows]=useState([])
  useEffect(()=>{
    let timer
    const fetchIt = async ()=>{ try{ setRows(await API.journaling(id)) }catch{} }
    fetchIt(); timer = setInterval(fetchIt, 3000) // polling 3s
    return ()=>clearInterval(timer)
  },[id])
  return (
    <div className="card">
      <div className="head"><b>Journaling</b><span className="badge">EXT3</span></div>
      <div className="body">
        {rows.length===0 && <div className="muted">Sin transacciones.</div>}
        {rows.map((j,i)=>(
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
