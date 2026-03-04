import { apiClient } from './client';
import type { Run, RunMetric, RunArtifact, LLMReviewResponse } from '../types';

export const runsApi = {
  get: (id: string) =>
    apiClient.get<Run>(`/v1/runs/${id}`).then(r => r.data),

  getMetrics: (id: string) =>
    apiClient.get<{ items: RunMetric[] }>(`/v1/runs/${id}/metrics`).then(r => r.data),

  getArtifacts: (id: string) =>
    apiClient.get<{ items: RunArtifact[] }>(`/v1/runs/${id}/artifacts`).then(r => r.data),

  llmReview: (id: string, mode: 'teacher' | 'concise' = 'teacher') =>
    apiClient.post<LLMReviewResponse>(`/v1/runs/${id}/llm-review`, { mode, promptVersion: 'v1' }).then(r => r.data),
};
