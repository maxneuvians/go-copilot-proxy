import { createBrowserRouter, Navigate } from 'react-router-dom';

import Navbar from './pages/Navbar/Navbar';
import Chat from './pages/Chat/Chat';
import AboutChat from './pages/AboutChat/AboutChat';

export const router = createBrowserRouter([
  {
    path: '/',
    element: <Navbar />,
    children: [
      {
        index: true,
        element: <Navigate to="/chat" replace />
      },
      {
        path: 'chat',
        element: <Chat />
      },
      {
        path: 'about',
        element: <AboutChat />
      }
    ]
  },
]);