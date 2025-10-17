import { useNavigate } from 'react-router-dom'
import ReportsGallery from '../components/ReportsGallery'

export default function Reports(){
  const navigate = useNavigate()

  return (
    <div style={{padding:'12px', minHeight:'calc(100vh - 60px)', display:'flex', flexDirection:'column', gap:'12px'}}>
      <div className="card" style={{flexShrink: 0}}>
        <div className="head">
          <b>Galería de Reportes</b>
          <span className="badge">Visualización 3D</span>
          <div style={{marginLeft:'auto', display:'flex', gap:8, flexWrap:'wrap'}}>
            <button className="btn" onClick={()=>navigate('/visualizer')}>Visualizador</button>
            <button className="btn alt" onClick={()=>navigate('/')}>Volver a Terminal</button>
          </div>
        </div>
        <div className="body" style={{minHeight: 'auto'}}>
          <p className="muted" style={{marginBottom: '0', fontSize:'13px', lineHeight:'1.6'}}>
            Visualiza todos los archivos generados en la carpeta <code>Backend/Reports</code>.
            La galería se actualiza automáticamente cada 5 segundos para mostrar nuevos reportes.
            Usa las flechas del teclado, botones de navegación, rueda del mouse o los indicadores para navegar.
          </p>
        </div>
      </div>

      <ReportsGallery/>
    </div>
  )
}
