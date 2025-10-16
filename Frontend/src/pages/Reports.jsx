import { useNavigate } from 'react-router-dom'
import ReportsCarousel3D from '../components/ReportsCarousel3D'

export default function Reports(){
  const navigate = useNavigate()

  return (
    <div style={{padding:'12px', minHeight:'calc(100vh - 60px)', display:'flex', flexDirection:'column'}}>
      <div className="card">
        <div className="head">
          <b>Centro de Reportes</b>
          <span className="badge">3D Coverflow</span>
          <div style={{marginLeft:'auto', display:'flex', gap:8, flexWrap:'wrap'}}>
            <button className="btn" onClick={()=>navigate('/visualizer')}>Visualizador</button>
            <button className="btn alt" onClick={()=>navigate('/')}>Volver a Terminal</button>
          </div>
        </div>
        <div className="body">
          <p className="muted" style={{marginBottom: '8px', fontSize:'13px', lineHeight:'1.6'}}>
            Genera reportes visuales y textuales del sistema de archivos con un carrusel 3D interactivo.
            Usa las flechas del teclado, botones de navegación, rueda del mouse o los indicadores para navegar.
            Haz clic en las imágenes para verlas en pantalla completa con zoom.
          </p>
        </div>
      </div>
      <ReportsCarousel3D/>
    </div>
  )
}
