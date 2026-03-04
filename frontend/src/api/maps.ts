import { apiClient } from './client';
import type { Map, PaginatedResponse } from '../types';

export interface CreateMapInput {
  name: string;
  width: number;
  height: number;
  grid: string;
}

export const mapsApi = {
  create: (data: CreateMapInput) =>
    apiClient.post<Map>('/v1/maps', data).then(r => r.data),

  list: (limit = 20, offset = 0) =>
    apiClient.get<PaginatedResponse<Map>>(`/v1/maps?limit=${limit}&offset=${offset}`).then(r => r.data),

  get: (id: string) =>
    apiClient.get<Map>(`/v1/maps/${id}`).then(r => r.data),

  update: (id: string, data: CreateMapInput) =>
    apiClient.put<Map>(`/v1/maps/${id}`, data).then(r => r.data),

  delete: (id: string) =>
    apiClient.delete(`/v1/maps/${id}`),
};
