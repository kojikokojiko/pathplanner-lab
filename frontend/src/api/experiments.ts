import { apiClient } from './client';
import type { Experiment, Run, RunInput, PaginatedResponse } from '../types';

export interface CreateExperimentInput {
  mapId: string;
  name: string;
  start: { x: number; y: number };
  goal: { x: number; y: number };
  runs: RunInput[];
  seed?: number;
}

export const experimentsApi = {
  create: (data: CreateExperimentInput, idempotencyKey?: string) =>
    apiClient.post<Experiment>('/v1/experiments', data, {
      headers: idempotencyKey ? { 'Idempotency-Key': idempotencyKey } : {},
    }).then(r => r.data),

  list: (limit = 20, offset = 0) =>
    apiClient.get<PaginatedResponse<Experiment>>(`/v1/experiments?limit=${limit}&offset=${offset}`).then(r => r.data),

  get: (id: string) =>
    apiClient.get<Experiment>(`/v1/experiments/${id}`).then(r => r.data),

  getRuns: (id: string) =>
    apiClient.get<{ items: Run[] }>(`/v1/experiments/${id}/runs`).then(r => r.data),

  cancel: (id: string) =>
    apiClient.post(`/v1/experiments/${id}/cancel`).then(r => r.data),
};
