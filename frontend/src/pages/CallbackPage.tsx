import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth0 } from '@auth0/auth0-react';

export default function CallbackPage() {
  const navigate = useNavigate();
  const { isAuthenticated, isLoading, error } = useAuth0();

  useEffect(() => {
    if (isLoading) return;
    if (error || !isAuthenticated) {
      navigate('/login', { replace: true });
    } else {
      navigate('/maps', { replace: true });
    }
  }, [isAuthenticated, isLoading, error, navigate]);

  return (
    <div className="min-h-screen bg-gray-900 flex items-center justify-center">
      <p className="text-gray-400">Signing in…</p>
    </div>
  );
}
