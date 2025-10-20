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
    setErr('')
    setLoading(true)
    try{
      await API.login(id, user, pass)
      setSession({id, user})
      navigate('/visualizer')
    }catch(e){
      setErr(e.message)
    }finally{
      setLoading(false)
    }
  }

  return (
    <div className="login-container">
      <a href="/" className="login-back" onClick={(e)=>{e.preventDefault();navigate('/')}}>
        ← Volver
      </a>

      <form className="login-form" onSubmit={handleLogin} autoComplete="off">
        <div className="form-control">
          <h1 className="login-title">Sign In</h1>
        </div>

        {/* ID de Montaje */}
        <div className="form-control block-cube block-input">
          <input
            name="mount_id"
            type="text"
            placeholder="Mount ID (ej: 841A)"
            value={id}
            onChange={(e)=>setId(e.target.value)}
            required
            autoFocus
          />
          <div className="bg-top">
            <div className="bg-inner"></div>
          </div>
          <div className="bg-right">
            <div className="bg-inner"></div>
          </div>
          <div className="bg">
            <div className="bg-inner"></div>
          </div>
        </div>

        {/* Usuario */}
        <div className="form-control block-cube block-input">
          <input
            name="username"
            type="text"
            placeholder="Username (ej: root)"
            value={user}
            onChange={(e)=>setUser(e.target.value)}
            required
          />
          <div className="bg-top">
            <div className="bg-inner"></div>
          </div>
          <div className="bg-right">
            <div className="bg-inner"></div>
          </div>
          <div className="bg">
            <div className="bg-inner"></div>
          </div>
        </div>

        {/* Contraseña */}
        <div className="form-control block-cube block-input">
          <input
            name="password"
            type="password"
            placeholder="Password (ej: 123)"
            value={pass}
            onChange={(e)=>setPass(e.target.value)}
            required
          />
          <div className="bg-top">
            <div className="bg-inner"></div>
          </div>
          <div className="bg-right">
            <div className="bg-inner"></div>
          </div>
          <div className="bg">
            <div className="bg-inner"></div>
          </div>
        </div>

        {/* Botón de Login */}
        <button
          className="login-btn block-cube block-cube-hover"
          type="submit"
          disabled={loading}
        >
          <div className="bg-top">
            <div className="bg-inner"></div>
          </div>
          <div className="bg-right">
            <div className="bg-inner"></div>
          </div>
          <div className="bg">
            <div className="bg-inner"></div>
          </div>
          <div className="text">
            {loading ? 'Logging In...' : 'Log In'}
          </div>
        </button>

        {/* Error Message */}
        {err && (
          <div className="login-error">
            <b>Error:</b> {err}
          </div>
        )}

        {/* Info adicional */}
        <div className="login-info">
          <small>
            <b>Nota:</b> El login se realiza únicamente por GUI.
            Los comandos que requieren sesión incluyen: <b>mkgrp</b>, <b>rmgrp</b>, <b>mkusr</b>, <b>rmusr</b>,
            <b>chmod</b>, <b>mkfile</b>, <b>cat</b>, <b>remove</b>, <b>edit</b>, <b>rename</b>, <b>mkdir</b>, <b>copy</b>, <b>move</b>, <b>find</b>, <b>chgrp</b>, <b>chown</b>.
          </small>
        </div>
      </form>
    </div>
  )
}
