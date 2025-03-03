import '@mantine/core/styles.css';

import { RouterProvider } from 'react-router-dom';
import { router } from './Router';

import { SettingsProvider } from './context/SettingsContext';

function App() {
  return (
    <SettingsProvider>
      <RouterProvider router={router} />
    </SettingsProvider>
  );
}

export default App;