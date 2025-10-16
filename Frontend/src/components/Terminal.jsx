import { useRef, useState, useEffect } from 'react'
import { API } from '../lib/api'

const SESSION_REQUIRED_CMDS = ['mkgrp', 'rmgrp', 'mkusr', 'rmusr', 'chmod', 'mkfile', 'cat', 'remove', 'edit', 'rename', 'mkdir', 'copy', 'move', 'find', 'chgrp', 'chown']

// Persistent storage keys
const TERMINAL_HISTORY_KEY = 'mia_terminal_history'
const TERMINAL_LINES_KEY = 'mia_terminal_lines'
const TERMINAL_STATS_KEY = 'mia_terminal_stats'

export default function Terminal({session}){
  // Load persisted state
  const [lines,setLines]=useState(() => {
    const saved = localStorage.getItem(TERMINAL_LINES_KEY)
    if(saved) {
      try { return JSON.parse(saved) }
      catch(e) { return [{t:'Bienvenido a MIA - Sistema de Archivos EXT2/EXT3',k:'sys'}] }
    }
    return [{t:'Bienvenido a MIA - Sistema de Archivos EXT2/EXT3',k:'sys'}]
  })

  const [commandArea,setCommandArea]=useState('')
  const [busy,setBusy]=useState(false)
  const [hist,setHist]=useState(() => {
    const saved = localStorage.getItem(TERMINAL_HISTORY_KEY)
    if(saved) {
      try { return JSON.parse(saved) }
      catch(e) { return [] }
    }
    return []
  })

  // Stats: {total, success, errors}
  const [stats,setStats]=useState(() => {
    const saved = localStorage.getItem(TERMINAL_STATS_KEY)
    if(saved) {
      try { return JSON.parse(saved) }
      catch(e) { return {total:0, success:0, errors:0} }
    }
    return {total:0, success:0, errors:0}
  })

  const outputRef=useRef(null)
  const inputRef=useRef(null)
  const fileInputRef=useRef(null)

  // Auto-scroll output
  useEffect(()=>{
    outputRef.current?.scrollTo({top:1e8, behavior:'smooth'})
  },[lines])

  // Persist state
  useEffect(()=>{
    localStorage.setItem(TERMINAL_LINES_KEY, JSON.stringify(lines))
  },[lines])

  useEffect(()=>{
    localStorage.setItem(TERMINAL_HISTORY_KEY, JSON.stringify(hist))
  },[hist])

  useEffect(()=>{
    localStorage.setItem(TERMINAL_STATS_KEY, JSON.stringify(stats))
  },[stats])

  function requiresSession(line){
    const cmd = line.trim().split(/\s+/)[0].toLowerCase()
    return SESSION_REQUIRED_CMDS.includes(cmd)
  }

  async function executeCommand(line){
    if(!line.trim()) return { success: false, skipped: true }

    // Check if command requires session
    if(requiresSession(line) && !session){
      setLines(p=>[...p,{t:`$ ${line}`,k:'cmd'},{t:'ERROR: Este comando requiere iniciar sesión primero',k:'err'}])
      return { success: false, skipped: false }
    }

    setLines(p=>[...p,{t:`$ ${line}`,k:'cmd'}])
    try{
      const out=await API.run(line)
      out.split('\n').forEach(s=> setLines(p=>[...p,{t:s,k:'ok'}]))
      return { success: true, skipped: false }
    }catch(e){
      setLines(p=>[...p,{t:'ERROR: '+e.message,k:'err'}])
      return { success: false, skipped: false }
    }
  }

  async function executeAll(){
    const commands = commandArea.trim().split('\n').filter(l => l.trim() && !l.trim().startsWith('#'))
    if(commands.length === 0) return

    setBusy(true)
    setLines(p=>[...p,{t:'=== Ejecutando comandos del área de entrada ===',k:'sys'}])

    let successCount = 0
    let errorCount = 0

    for(const cmd of commands){
      const result = await executeCommand(cmd)
      if(!result.skipped) {
        if(result.success) successCount++
        else errorCount++
      }
      setHist(h=>[...h, cmd])
    }

    const totalExecuted = successCount + errorCount
    setStats(prev => ({
      total: prev.total + totalExecuted,
      success: prev.success + successCount,
      errors: prev.errors + errorCount
    }))

    setLines(p=>[...p,{
      t:`=== Ejecución completada: ${totalExecuted} comando(s) - ✓ ${successCount} exitoso(s) - ✗ ${errorCount} error(es) ===`,
      k:'sys'
    }])

    setBusy(false)
    setCommandArea('')
  }

  function handleFileUpload(e){
    const file = e.target.files?.[0]
    if(!file) return

    const reader = new FileReader()
    reader.onload = (evt) => {
      const content = evt.target?.result
      if(typeof content === 'string'){
        setCommandArea(content)
        setLines(p=>[...p,{t:`Archivo cargado: ${file.name} (${content.split('\n').length} líneas)`,k:'sys'}])
      }
    }
    reader.readAsText(file)
    e.target.value = '' // Reset input
  }

  function handleKeyDown(e){
    // Ctrl+Enter: ejecutar todo
    if(e.ctrlKey && e.key === 'Enter'){
      e.preventDefault()
      executeAll()
    }
  }

  function clearTerminal(){
    setLines([{t:'Terminal limpiada',k:'sys'}])
  }

  function resetStats(){
    setStats({total:0, success:0, errors:0})
    setLines(p=>[...p,{t:'Estadísticas reseteadas',k:'sys'}])
  }

  return (
    <div style={{display:'grid', gridTemplateRows:'auto 1fr auto', height:'100%', gap:'12px'}}>
      {/* Output Area */}
      <div style={{
        background:'var(--panel)',
        border:'1px solid var(--border)',
        borderRadius:'10px',
        padding:'12px',
        minHeight:'300px',
        maxHeight:'400px',
        overflow:'auto',
        fontFamily:'monospace',
        fontSize:'13px',
        lineHeight:'1.5'
      }} ref={outputRef}>
        <div style={{
          marginBottom:'8px',
          padding:'8px',
          background:'var(--panel2)',
          borderRadius:'6px',
          display:'flex',
          justifyContent:'space-between',
          alignItems:'center',
          flexWrap:'wrap',
          gap:'8px'
        }}>
          <div style={{display:'flex', alignItems:'center', gap:'8px'}}>
            <span style={{color:'var(--neon)', fontWeight:'700'}}>julian@pop-os</span>
            <span style={{color:'var(--muted)'}}>:</span>
            <span style={{color:'var(--neo2)', fontWeight:'700'}}>~/home</span>
            <span style={{color:'var(--neon)', fontSize:'16px'}}>$</span>
          </div>
          <div style={{display:'flex', gap:'12px', fontSize:'11px'}}>
            <span className="badge" style={{background:'var(--panel)', borderColor:'var(--muted)'}}>
              Total: {stats.total}
            </span>
            <span className="badge" style={{background:'#001f14', borderColor:'#00c77a'}}>
              OK: {stats.success}
            </span>
            <span className="badge" style={{background:'#1f0010', borderColor:'#ff5c7c'}}>
              ERR: {stats.errors}
            </span>
            <button
              onClick={resetStats}
              style={{
                background:'transparent',
                border:'none',
                color:'var(--muted)',
                cursor:'pointer',
                fontSize:'10px',
                padding:'0 4px'
              }}
              title="Resetear estadísticas"
            >
              Reset
            </button>
          </div>
        </div>
        {lines.map((ln,i)=><div key={i} className={`line ${ln.k}`}>{ln.t}</div>)}
        {busy && <div className="line sys">Ejecutando comandos...</div>}
      </div>

      {/* Input Area */}
      <div style={{
        background:'var(--panel)',
        border:'1px solid var(--border)',
        borderRadius:'10px',
        padding:'12px',
        display:'flex',
        flexDirection:'column',
        gap:'10px'
      }}>
        <div style={{display:'flex', justifyContent:'space-between', alignItems:'center', flexWrap:'wrap', gap:'8px'}}>
          <label style={{fontSize:'14px', fontWeight:'600'}}>
            Área de Entrada de Comandos
          </label>
          <div style={{display:'flex', gap:'8px', flexWrap:'wrap'}}>
            <input
              type="file"
              ref={fileInputRef}
              onChange={handleFileUpload}
              accept=".mia,.smia,.txt"
              style={{display:'none'}}
            />
            <button
              className="btn alt"
              onClick={()=>fileInputRef.current?.click()}
              disabled={busy}
              style={{fontSize:'12px', padding:'6px 12px'}}
            >
              Cargar Archivo
            </button>
            <button
              className="btn"
              onClick={executeAll}
              disabled={busy || !commandArea.trim()}
              style={{fontSize:'12px', padding:'6px 12px'}}
            >
              Ejecutar
            </button>
            <button
              className="btn alt"
              onClick={clearTerminal}
              disabled={busy}
              style={{fontSize:'12px', padding:'6px 12px'}}
            >
              Limpiar
            </button>
          </div>
        </div>

        <textarea
          ref={inputRef}
          className="input mono"
          value={commandArea}
          onChange={e=>setCommandArea(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={busy}
          rows={8}
          placeholder={`# Ingrese comandos aquí (uno por línea)
mkdisk -size=10 -unit=M -path="/tmp/Disco1.mia"
fdisk -size=1024 -unit=K -type=P -path="/tmp/Disco1.mia" -name=Part1
mount -path="/tmp/Disco1.mia" -name=Part1

# Presione Ctrl+Enter para ejecutar todos los comandos
# O use el botón "Ejecutar"`}
          style={{minHeight:'180px', resize:'vertical'}}
        />

        <small className="muted" style={{fontSize:'11px'}}>
          <b>Tip:</b> Use # para comentarios | Ctrl+Enter para ejecutar | Las estadísticas se mantienen entre sesiones
        </small>
      </div>
    </div>
  )
}
