const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export type CmdResponse = {
  ok: boolean
  output?: string
  error?: string
  input?: string
  params?: Record<string, string>
  usage?: string
}

export async function runCmd(line: string): Promise<CmdResponse> {
  const res = await fetch(`${API_URL}/api/cmd/run`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ line }),
  })
  return res.json()
}

// ==== Reports (DOT) ====
export async function getReportMBR(id: string): Promise<string> {
  const res = await fetch(`${API_URL}/api/reports/mbr?id=${encodeURIComponent(id)}`)
  if (!res.ok) throw new Error('Error al obtener reporte MBR')
  return res.text()
}
export async function getReportFSTree(id: string, path: string): Promise<string> {
  const res = await fetch(`${API_URL}/api/reports/tree?id=${encodeURIComponent(id)}&path=${encodeURIComponent(path)}`)
  if (!res.ok) throw new Error('Error al obtener reporte Tree')
  return res.text()
}
export async function getReportJournal(id: string): Promise<string> {
  const res = await fetch(`${API_URL}/api/reports/journal?id=${encodeURIComponent(id)}`)
  if (!res.ok) throw new Error('Error al obtener reporte Journal')
  return res.text()
}