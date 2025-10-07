const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

// ===== Types =====
export type CmdResponse = {
  ok: boolean
  output?: string
  error?: string
  input?: string
  command?: string
  params?: Record<string, string>
  usage?: string
}

export type ScriptResponse = {
  ok: boolean
  error?: string
  results?: CommandResult[]
  total_lines?: number
  executed?: number
  success_count?: number
  error_count?: number
}

export type CommandResult = {
  line: number
  input: string
  output: string
  success: boolean
  error?: string
}

export type DiskInfo = {
  ok: boolean
  path?: string
  size?: number
  modified?: string
  mbr_size?: number
  created_at?: string
  signature?: number
  fit?: string
  partitions?: PartitionInfo[]
  error?: string
}

export type PartitionInfo = {
  index: number
  name: string
  type: string
  fit: string
  start: number
  size: number
}

export type MountedPartition = {
  disk_path: string
  partition_id: string
  mount_id: string
}

// ===== Commands =====
export async function runCmd(line: string): Promise<CmdResponse> {
  const res = await fetch(`${API_URL}/api/cmd/run`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ line }),
  })
  return res.json()
}

export async function executeCommand(line: string): Promise<CmdResponse> {
  const res = await fetch(`${API_URL}/api/cmd/execute`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ line }),
  })
  return res.json()
}

export async function executeScript(script: string): Promise<ScriptResponse> {
  const res = await fetch(`${API_URL}/api/cmd/script`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ script }),
  })
  return res.json()
}

export async function validateCommand(line: string): Promise<CmdResponse> {
  const res = await fetch(`${API_URL}/api/cmd/validate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ line }),
  })
  return res.json()
}

// ===== Disk Management =====
export async function listDisks(path?: string): Promise<any> {
  const url = path
    ? `${API_URL}/api/disks?path=${encodeURIComponent(path)}`
    : `${API_URL}/api/disks`
  const res = await fetch(url)
  return res.json()
}

export async function getDiskInfo(path: string): Promise<DiskInfo> {
  const res = await fetch(`${API_URL}/api/disks/info?path=${encodeURIComponent(path)}`)
  return res.json()
}

export async function listMounted(): Promise<{ ok: boolean; partitions?: MountedPartition[]; count?: number; error?: string }> {
  const res = await fetch(`${API_URL}/api/mounted`)
  return res.json()
}

// ===== Reports (DOT) =====
export async function getReportMBR(id: string): Promise<string> {
  const res = await fetch(`${API_URL}/api/reports/mbr?id=${encodeURIComponent(id)}`)
  if (!res.ok) throw new Error('Error al obtener reporte MBR')
  return res.text()
}

export async function getReportDisk(id: string): Promise<string> {
  const res = await fetch(`${API_URL}/api/reports/disk?id=${encodeURIComponent(id)}`)
  if (!res.ok) throw new Error('Error al obtener reporte Disk')
  return res.text()
}

export async function getReportSuperblock(id: string): Promise<string> {
  const res = await fetch(`${API_URL}/api/reports/sb?id=${encodeURIComponent(id)}`)
  if (!res.ok) throw new Error('Error al obtener reporte Superblock')
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

// ===== Health =====
export async function checkHealth(): Promise<boolean> {
  try {
    const res = await fetch(`${API_URL}/healthz`)
    return res.ok
  } catch {
    return false
  }
}