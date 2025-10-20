import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import DiskPicker from '../components/DiskPicker'
import PartitionPicker from '../components/PartitionPicker'
import Explorer from '../components/Explorer'
import JournalPanel from '../components/JournalPanel'

export default function Visualizer({session}){
  const navigate = useNavigate()
  const [disk,setDisk]=useState(null)
  const [partition,setPartition]=useState(null)
  const [step,setStep]=useState(1) // 1: disk, 2: partition, 3: filesystem

  function handleDiskSelect(d){
    setDisk(d)
    setPartition(null)
    setStep(2)
  }

  function handlePartitionSelect(p){
    setPartition(p)
    if(!session){
      alert('Debe iniciar sesión primero para explorar el sistema de archivos')
      navigate('/login')
      return
    }
    setStep(3)
  }

  function reset(){
    setDisk(null)
    setPartition(null)
    setStep(1)
  }

  function backToDisks(){
    setPartition(null)
    setStep(1)
  }

  function backToPartitions(){
    setPartition(null)
    setStep(2)
  }

  return (
    <div style={{padding:'12px', minHeight:'calc(100vh - 60px)', display:'flex', flexDirection:'column', gap:'12px'}}>
      {/* Breadcrumb / Steps */}
      <div className="card" style={{flexShrink: 0}}>
        <div className="head">
          <b>Visualizador del Sistema de Archivos</b>
          <div style={{marginLeft:'auto', display:'flex', gap:8, flexWrap:'wrap'}}>
            {step > 1 && <button className="btn alt" onClick={reset}>Reiniciar</button>}
            <button className="btn" onClick={()=>navigate('/reports')}>Reportes</button>
            <button className="btn alt" onClick={()=>navigate('/')}>Volver a Terminal</button>
          </div>
        </div>
        <div className="body" style={{minHeight:'auto'}}>
          <div className="breadcrumb" style={{justifyContent:'center'}}>
            <div className={`cr ${step>=1?'active':''}`}>1. Disco</div>
            <div className={`cr ${step>=2?'active':''}`}>2. Partición</div>
            <div className={`cr ${step>=3?'active':''}`}>3. Sistema de Archivos</div>
          </div>
        </div>
      </div>

      {/* Step 1: Disk Selection */}
      {step === 1 && <DiskPicker onSelect={handleDiskSelect}/>}

      {/* Step 2: Partition Selection */}
      {step === 2 && disk && (
        <>
          <div className="card" style={{flexShrink: 0}}>
            <div className="head">
              <b>Disco Seleccionado</b>
              <span className="badge mono">{disk.path}</span>
            </div>
            <div className="body kv" style={{minHeight:'auto'}}>
              <div className="muted">Capacidad</div><div>{disk.size}</div>
              <div className="muted">Fit</div><div>{disk.fit}</div>
              <div className="muted">Particiones Montadas</div><div>{disk.mounted?.length || 0}</div>
            </div>
          </div>
          <PartitionPicker disk={disk} onSelect={handlePartitionSelect} onBack={backToDisks}/>
        </>
      )}

      {/* Step 3: Filesystem Explorer */}
      {step === 3 && partition && session && (
        <>
          <div className="card" style={{flexShrink: 0}}>
            <div className="head">
              <b>Partición Seleccionada</b>
              <span className="badge">{partition.name}</span>
            </div>
            <div className="body kv" style={{minHeight:'auto'}}>
              <div className="muted">Disco</div><div className="mono">{disk.path}</div>
              <div className="muted">Tipo</div><div>{partition.type}</div>
              <div className="muted">Tamaño</div><div>{partition.size}</div>
              <div className="muted">Fit</div><div>{partition.fit}</div>
              <div className="muted">Sesión</div><div>{session.user} (ID: {session.id})</div>
            </div>
          </div>

          <div className="grid cols-2">
            <Explorer id={session.id} onBack={backToPartitions}/>
            <JournalPanel id={session.id}/>
          </div>
        </>
      )}
    </div>
  )
}
