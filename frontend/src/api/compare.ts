import { apiClient } from './client';
import type { CompareResponse, LLMReviewResponse } from '../types';

export const compareApi = {
  compare: (runIds: string[]) =>
    apiClient.post<CompareResponse>('/v1/compare', { runIds }).then(r => r.data),

  llmReview: (runIds: string[], mode: 'teacher' | 'concise' = 'teacher') =>
    apiClient.post<LLMReviewResponse>('/v1/compare/llm-review', { runIds, mode, promptVersion: 'v1' }).then(r => r.data),
};
