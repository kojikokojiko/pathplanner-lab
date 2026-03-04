import { useEffect } from 'react';
import { useAuth0 } from '@auth0/auth0-react';
import { apiClient } from '../api/client';

export function useAuthSetup() {
  const { getAccessTokenSilently, logout } = useAuth0();

  useEffect(() => {
    const reqId = apiClient.interceptors.request.use(async (config) => {
      try {
        const token = await getAccessTokenSilently();
        config.headers.Authorization = `Bearer ${token}`;
      } catch {
        // not authenticated — let the request go through unauthenticated
      }
      return config;
    });

    const resId = apiClient.interceptors.response.use(
      (res) => res,
      (err) => {
        if (err.response?.status === 401) {
          logout({ logoutParams: { returnTo: `${window.location.origin}/login` } });
        }
        return Promise.reject(err);
      },
    );

    return () => {
      apiClient.interceptors.request.eject(reqId);
      apiClient.interceptors.response.eject(resId);
    };
  }, [getAccessTokenSilently, logout]);
}
