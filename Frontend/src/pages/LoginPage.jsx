import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { API } from '../lib/api'

export default function LoginPage({setSession}){
  const navigate = useNavigate()
  const [id,setId]=useState('')
  const [user,setUser]=useState('')
  const [pass,setPass]=useState('')
  const [err,setErr]=useState('')
  const [loading,setLoading]=useState(false)

  async function handleLogin(e){
    e.preventDefault()
    setErr(''); setLoading(true)
    try{
      await API.login(id, user, pass)
      setSession({id, user})
      navigate('/')
    }catch(e){
      setErr(e.message)
    }finally{
      setLoading(false)
    }
  }

  return (
    <div style={{
      display:'grid',
      placeItems:'center',
      minHeight:'calc(100vh - 60px)',
      padding:'20px',
      background:'var(--bg)'
    }}>
      <div className="card" style={{width:'100%', maxWidth:'480px'}}>
        <div className="head">
          <b>Iniciar Sesión</b>
          <span className="badge">Login GUI</span>
        </div>
        <div className="body">
          <form onSubmit={handleLogin} className="grid">
            <div className="field">
              <label><b>ID de Montaje</b></label>
              <input
                className="input"
                placeholder="Ejemplo: 841A"
                value={id}
                onChange={e=>setId(e.target.value)}
                required
                autoFocus
              />
              <small className="muted">ID retornado por el comando mount</small>
            </div>

            <div className="field">
              <label><b>Usuario</b></label>
              <input
                className="input"
                placeholder="root"
                value={user}
                onChange={e=>setUser(e.target.value)}
                required
              />
            </div>

            <div className="field">
              <label><b>Contraseña</b></label>
              <input
                className="input"
                type="password"
                placeholder="123"
                value={pass}
                onChange={e=>setPass(e.target.value)}
                required
              />
            </div>

            {err && (
              <div style={{
                padding:'10px',
                background:'var(--panel2)',
                border:'1px solid var(--danger)',
                borderRadius:'10px',
                color:' var(--danger)'
              }}>
                <b>Error:</b> {err}
              </div>
            )}

            <div style={{display:'flex', gap:10, marginTop:10}}>
              <button type="button" className="btn alt" onClick={()=>navigate('/')}>
                Cancelar
              </button>
              <button type="submit" className="btn" disabled={loading} style={{flex:1}}>
                {loading ? 'Iniciando...' : 'Iniciar Sesión'}
              </button>
            </div>

            <div style={{marginTop:12, padding:'10px', background:'var(--panel2)', borderRadius:'10px'}}>
              <small className="muted">
                <b>Nota:</b> El login ahora se realiza únicamente por GUI.
                Los comandos que requieren sesión incluyen: mkgrp, rmgrp, mkusr, rmusr,
                chmod, mkfile, cat, remove, edit, rename, mkdir, copy, move, find, chgrp, chown.
              </small>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}
