import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import Layout from './Layout'
import Home from './pages/Home'
import TorrentDetail from './pages/TorrentDetail'
import AddTorrent from './pages/AddTorrent'

const router = createBrowserRouter([
  {
    path: '/',
    element: <Layout />,
    children: [
      {
        index: true,
        element: <Home />,
      },
      {
        path: 'torrent/:id',
        element: <TorrentDetail />,
      },
      {
        path: 'add',
        element: <AddTorrent />,
      },
    ],
  },
])

function App() {
  return <RouterProvider router={router} />
}

export default App
