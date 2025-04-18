import { useState } from 'react'

function App() {
  const [devices, setDevices] = useState([])

  return (
    <div className="app">
      <header>
        <h1>Smart Home</h1>
      </header>
      <main>
        <div className="dashboard">
          {devices.length === 0 ? (
            <p>Нет подключенных устройств</p>
          ) : (
            <div className="devices-grid">
              {/* Устройства будут отображаться здесь */}
            </div>
          )}
        </div>
      </main>
    </div>
  )
}

export default App 