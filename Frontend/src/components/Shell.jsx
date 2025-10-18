import { useRef, useState, useEffect } from 'react'
import { API } from '../lib/api'

export default function Shell({session}){
  const [lines,setLines]=useState([{t:'Bienvenido a MIA - Sistema de Archivos EXT2/EXT3',k:'sys'},{t:'Ejecute comandos para crear discos, particiones y sistemas de archivos.',k:'sys'}])
  const [v,setV]=useState(''); const [busy,setBusy]=useState(false)
  const [hist,setHist]=useState([]); const [idx,setIdx]=useState(0)
  const bodyRef=useRef(null)
  useEffect(()=>{bodyRef.current?.scrollTo({top:1e8})},[lines])

  async function send(){
    const line=v.trim(); if(!line) return

    setLines(p=>[...p,{t:`$ ${line}`,k:'cmd'}]); setV(''); setBusy(true)
    setHist(h=>[...h,line]); setIdx(0)
    try{
      const out=await API.run(line)
      out.split('\n').forEach(s=> setLines(p=>[...p,{t:s,k:'ok'}]))
    }catch(e){ setLines(p=>[...p,{t:'ERR: '+e.message,k:'err'}]) } finally{ setBusy(false) }
  }
  function key(e){
    if(e.key==='Enter' && !e.shiftKey){ e.preventDefault(); send() }
    if(e.key==='ArrowUp'){ e.preventDefault(); const i=Math.min(hist.length, idx+1); setIdx(i); setV(hist[hist.length-i]??'') }
    if(e.key==='ArrowDown'){ e.preventDefault(); const i=Math.max(0, idx-1); setIdx(i); setV(i===0?'':(hist[hist.length-i]??'')) }
    if(e.ctrlKey && e.key.toLowerCase()==='l'){ e.preventDefault(); setLines([]) }
  }
  return (
    <div className="shell">
      <div className="body mono" ref={bodyRef}>
        {lines.map((ln,i)=><div key={i} className={`line ${ln.k}`}>{ln.t}</div>)}
      </div>
      <div className="inputRow">
        <span className="prompt mono">user@mia:~$</span>
        <textarea
          className="input mono"
          rows={2}
          value={v}
          onChange={e=>setV(e.target.value)}
          onKeyDown={key}
          placeholder='mkdisk -size=10 -unit=M -path="/tmp/Disco1.mia"'
          disabled={busy}
        />
        <button className="btn" disabled={busy} onClick={send} style={{whiteSpace:'nowrap'}}>
          {busy ? 'Ejecutando...' : 'Ejecutar'}
        </button>
      </div>
    </div>
  )
}
