import './styles.css'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { useState } from 'react'
import Topbar from './components/Topbar'
import LoginPage from './pages/LoginPage'
import Home from './pages/Home'
import Visualizer from './pages/Visualizer'
import { API } from './lib/api'

export default function App(){
  const [session,setSession]=useState(null)

  async function logout(){
    await API.logout();
    setSession(null)
  }

  return (
    <BrowserRouter>
      <div className="app">
        <Topbar session={session} onLogout={logout}/>
        <Routes>
          <Route path="/" element={<Home session={session}/>}/>
          <Route path="/login" element={<LoginPage setSession={setSession}/>}/>
          <Route path="/visualizer" element={<Visualizer session={session}/>}/>
        </Routes>
      </div>
    </BrowserRouter>
  )
}
