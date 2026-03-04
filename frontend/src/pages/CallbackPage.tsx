import { useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { handleCallback } from '../lib/auth0';

export default function CallbackPage() {
  const navigate = useNavigate();
  const called = useRef(false);

  useEffect(() => {
    if (called.current) return;
    called.current = true;

    const params = new URLSearchParams(window.location.search);
    const code = params.get('code');
    const state = params.get('state') ?? '';
    const error = params.get('error');

    if (error || !code) {
      navigate('/login');
      return;
    }

    handleCallback(code, state)
      .then((idToken) => {
        localStorage.setItem('token', idToken);
        navigate('/maps');
      })
      .catch(() => {
        navigate('/login');
      });
  }, [navigate]);

  return (
    <div className="min-h-screen bg-gray-900 flex items-center justify-center">
      <p className="text-gray-400">Signing in…</p>
    </div>
  );
}
