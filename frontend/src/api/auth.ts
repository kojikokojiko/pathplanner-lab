import { apiClient } from './client';
import type { User } from '../types';

export interface AuthResponse {
  token: string;
  user: User;
}

export const authApi = {
  register: (email: string, password: string) =>
    apiClient.post<AuthResponse>('/v1/auth/register', { email, password }).then(r => r.data),

  login: (email: string, password: string) =>
    apiClient.post<AuthResponse>('/v1/auth/login', { email, password }).then(r => r.data),

  me: () =>
    apiClient.get<User>('/v1/me').then(r => r.data),
};
