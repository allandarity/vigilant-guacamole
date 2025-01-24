import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App.tsx'
import './index.css'
import { RouterProvider, createBrowserRouter } from 'react-router-dom'
import AllRecommendations from './pages/allReccomendations/AllReccomendations.tsx'
import LetterboxdRecommendations from "./pages/letterboxdRecommendations/LetterboxdReccomendations.tsx";


const router = createBrowserRouter([
  {
    path: "/",
    element: <App />
  },
  {
    path: "/all",
    element: <AllRecommendations />
  },
  {
    path: "/watchlist",
    element: <LetterboxdRecommendations />
  }

])

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>,
)
