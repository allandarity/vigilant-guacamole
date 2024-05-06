import { Link } from 'react-router-dom'
import './App.css'
function App() {

  return (
    <>
      <div className="App">
        <div className='button-container'>
          <Link to="/path1" className="btn">
            Path 1
          </Link>
          <Link to="/path2" className="btn">
            Path 2
          </Link>
        </div>
      </div>
    </>
  )
}

export default App
