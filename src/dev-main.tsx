import {StrictMode} from 'react';
import {createRoot} from 'react-dom/client';
import AppDev from './AppDev.tsx';
import './index.css';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AppDev />
  </StrictMode>,
);
