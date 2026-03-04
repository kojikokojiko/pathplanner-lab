import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useAuthSetup } from './hooks/useAuthSetup';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import CallbackPage from './pages/CallbackPage';
import MapsListPage from './pages/MapsListPage';
import MapEditorPage from './pages/MapEditorPage';
import ExperimentsListPage from './pages/ExperimentsListPage';
import ExperimentDetailPage from './pages/ExperimentDetailPage';
import RunDetailPage from './pages/RunDetailPage';
import ComparePage from './pages/ComparePage';
import { PrivateRoute } from './components/layout/PrivateRoute';

export default function App() {
  useAuthSetup();
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/auth/callback" element={<CallbackPage />} />
        <Route path="/" element={<Navigate to="/maps" replace />} />
        <Route element={<PrivateRoute />}>
          <Route path="/maps" element={<MapsListPage />} />
          <Route path="/maps/new" element={<MapEditorPage />} />
          <Route path="/maps/:id" element={<MapEditorPage />} />
          <Route path="/experiments" element={<ExperimentsListPage />} />
          <Route path="/experiments/:id" element={<ExperimentDetailPage />} />
          <Route path="/runs/:id" element={<RunDetailPage />} />
          <Route path="/compare" element={<ComparePage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
