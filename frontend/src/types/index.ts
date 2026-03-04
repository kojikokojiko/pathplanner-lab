export interface User {
  id: string;
  email: string;
  created_at: string;
}

export interface Point {
  x: number;
  y: number;
}

export interface Map {
  id: string;
  owner_user_id: string;
  name: string;
  width: number;
  height: number;
  grid: string;
  obstacle_ratio: number;
  created_at: string;
  updated_at: string;
}

export type ExperimentStatus = 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED' | 'CANCELLED';
export type RunStatus = 'PENDING' | 'RUNNING' | 'SUCCEEDED' | 'FAILED';
export type AlgorithmKey = 'BFS' | 'DIJKSTRA' | 'ASTAR' | 'WEIGHTED_ASTAR' | 'GREEDY';

export interface RunInput {
  algorithmKey: AlgorithmKey;
  params: Record<string, unknown>;
}

export interface Experiment {
  id: string;
  owner_user_id: string;
  map_id: string;
  name: string;
  start: Point;
  goal: Point;
  seed: number;
  status: ExperimentStatus;
  runs?: Run[];
  created_at: string;
  updated_at: string;
}

export interface Run {
  id: string;
  experiment_id: string;
  algorithm_key: AlgorithmKey;
  params: Record<string, unknown>;
  status: RunStatus;
  metrics?: RunMetric[];
  artifacts?: RunArtifact[];
  error_msg?: string;
  created_at: string;
  updated_at: string;
}

export interface RunMetric {
  run_id: string;
  key: string;
  value: number;
}

export interface RunArtifact {
  run_id: string;
  kind: 'PATH' | 'EXPANSION_HEATMAP' | 'TIMELINE';
  data: unknown;
}

export interface LLMReport {
  id: string;
  run_id?: string;
  compare_hash?: string;
  model: string;
  prompt_version: string;
  summary: string;
  recommendations: string[];
  created_at: string;
}

export interface CompareRow {
  run_id: string;
  algorithm: string;
  status: string;
  found: boolean;
  path_length: number;
  expanded_nodes: number;
  time_ms: number;
  turns: number;
  explored_ratio: number;
}

export interface CompareResponse {
  table: CompareRow[];
  insights?: string;
}

export interface LLMReviewResponse {
  reportId: string;
  summary: string;
  recommendations: string[];
}

export interface PaginatedResponse<T> {
  items: T[];
  limit: number;
  offset: number;
}
