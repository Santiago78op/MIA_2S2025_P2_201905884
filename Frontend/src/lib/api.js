const BASE = '' // Vite proxy → :8080

export const API = {
  async health(){ const r = await fetch('/health'); return r.json() },

  async login(id, user, pass){
    const r = await fetch('/api/auth/login', {method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({ id, user, pass })})
    if(!r.ok) throw new Error((await r.json()).message||'Login falló'); return r.json()
  },
  async logout(){ await fetch('/api/auth/logout', {method:'POST'}) },

  async run(line){
    const r = await fetch('/api/commands',{method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({input: line})})
    const d = await r.json(); if(!r.ok) throw new Error(d.message||'Error'); return d.output||''
  },

  async runScript(script, stopOnError = true){
    const r = await fetch('/api/script', {method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({script, stopOnError})})
    const d = await r.json(); if(!r.ok) throw new Error(d.message||'Error en script'); return d.output||''
  },

  async disks(){ const r=await fetch('/api/disks'); if(!r.ok) throw new Error('disks'); return r.json() },
  async partitions(diskPath){ const r=await fetch(`/api/disks/${encodeURIComponent(diskPath)}/partitions`); if(!r.ok) throw new Error('parts'); return r.json() },

  async list(id, path){ const r=await fetch(`/api/fs/${encodeURIComponent(id)}/tree?path=${encodeURIComponent(path)}`); if(!r.ok) throw new Error('list'); return r.json() },
  async file(id, path){ const r=await fetch(`/api/fs/${encodeURIComponent(id)}/file?path=${encodeURIComponent(path)}`); if(!r.ok) throw new Error('file'); const d = await r.json(); return d.content },

  async journaling(id){ const r=await fetch(`/api/journal/${encodeURIComponent(id)}/table`); if(!r.ok) throw new Error('journal'); return r.json() },

  async genReport(name, id, out, extra=''){
    const r = await fetch('/api/reports', {
      method:'POST', headers:{'Content-Type':'application/json'},
      body: JSON.stringify({ name, id, out, extra })
    });
    const data = await r.json();
    if(!r.ok) throw new Error(data.message || 'Error generando reporte');
    return data.path; // path creado por el backend
  },

  async listReports(){
    const r = await fetch('/api/reports/list');
    if(!r.ok) throw new Error('Error listando reportes');
    return r.json(); // { files: [...], count: N }
  }
}
