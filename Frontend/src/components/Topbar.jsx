import { useEffect, useState } from 'react'
import { API } from '../lib/api'

export default function Topbar({session, onLogout}) {
  const [h,setH]=useState(null)
  const [theme,setTheme]=useState(localStorage.getItem('theme')||'neo')

  useEffect(()=>{API.health().then(setH).catch(()=>{})},[])

  useEffect(()=>{
    document.documentElement.setAttribute('data-theme', theme==='aurora'?'aurora':'')
    localStorage.setItem('theme', theme)
  },[theme])

  function toggleTheme(){
    setTheme(t => t==='neo'?'aurora':'neo')
  }

  return (
    <div className="topbar">
      <div className="brand">
        <div className="logo"></div>
        <div>
          <div style={{fontWeight:800, letterSpacing:1}}>MIA - Proyecto 2</div>
          <small className="muted">Sistema de Archivos EXT2/EXT3</small>
        </div>
      </div>
      <div style={{display:'flex', gap:10, alignItems:'center', flexWrap:'wrap'}}>
        <button className="theme-toggle" onClick={toggleTheme}>
          {theme==='neo'?'🟢 Neo':'🟣 Aurora'}
        </button>
        {h && <span className="badge mono">Backend: {h.status}</span>}
        {session?.user ? (
          <>
            <span className="badge">📂 ID: {session.id}</span>
            <span className="badge">👤 {session.user}</span>
            <button className="btn" onClick={onLogout} style={{fontSize:'12px', padding:'6px 12px'}}>
              Cerrar Sesión
            </button>
          </>
        ) : (
          <span className="badge" style={{background:'var(--panel2)', borderColor:'var(--muted)', color:'var(--muted)'}}>
            Sin sesión activa
          </span>
        )}
      </div>
    </div>
  )
}
