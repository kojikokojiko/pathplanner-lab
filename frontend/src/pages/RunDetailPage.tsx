import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';
import { ArrowLeft, Sparkles } from 'lucide-react';
import { runsApi } from '../api/runs';
import { experimentsApi } from '../api/experiments';
import { mapsApi } from '../api/maps';
import Badge from '../components/ui/Badge';
import Card from '../components/ui/Card';
import Button from '../components/ui/Button';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorMessage from '../components/ui/ErrorMessage';
import GridViewer from '../components/map/GridViewer';
import type { LLMReviewResponse } from '../types';

export default function RunDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [llmReport, setLlmReport] = useState<LLMReviewResponse | null>(null);
  const [llmError, setLlmError] = useState('');

  const { data: run, isLoading } = useQuery({
    queryKey: ['runs', id],
    queryFn: () => runsApi.get(id!),
    refetchInterval: (q) => ['PENDING','RUNNING'].includes(q.state.data?.status ?? '') ? 3000 : false,
    enabled: !!id,
  });

  const { data: metricsData } = useQuery({
    queryKey: ['runs', id, 'metrics'],
    queryFn: () => runsApi.getMetrics(id!),
    enabled: !!id && run?.status === 'SUCCEEDED',
  });

  const { data: artifactsData } = useQuery({
    queryKey: ['runs', id, 'artifacts'],
    queryFn: () => runsApi.getArtifacts(id!),
    enabled: !!id && run?.status === 'SUCCEEDED',
  });

  const { data: exp } = useQuery({
    queryKey: ['experiments', run?.experiment_id],
    queryFn: () => experimentsApi.get(run!.experiment_id),
    enabled: !!run?.experiment_id,
  });

  const { data: mapData } = useQuery({
    queryKey: ['maps', exp?.map_id],
    queryFn: () => mapsApi.get(exp!.map_id),
    enabled: !!exp?.map_id,
  });

  const llmMutation = useMutation({
    mutationFn: () => runsApi.llmReview(id!, 'teacher'),
    onSuccess: (data) => { setLlmReport(data); setLlmError(''); },
    onError: () => setLlmError('LLM review failed'),
  });

  if (isLoading) return <div className="flex justify-center py-16"><LoadingSpinner size="lg" /></div>;
  if (!run) return <ErrorMessage message="Run not found" />;

  const metrics = metricsData?.items ?? [];
  const artifacts = artifactsData?.items ?? [];

  return (
    <div>
      <div className="flex items-center gap-4 mb-6">
        <button
          onClick={() => run.experiment_id ? navigate(`/experiments/${run.experiment_id}`) : navigate('/experiments')}
          className="text-gray-400 hover:text-white"
        >
          <ArrowLeft size={20} />
        </button>
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold text-white font-mono">{run.algorithm_key}</h1>
            <Badge status={run.status} />
          </div>
          <p className="text-sm text-gray-400">Run ID: {run.id.slice(0, 8)}...</p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Metrics */}
        {metrics.length > 0 && (
          <Card>
            <h2 className="text-lg font-semibold text-white mb-3">Metrics</h2>
            <table className="w-full text-sm">
              <tbody>
                {metrics.map((m) => (
                  <tr key={m.key} className="border-b border-gray-700/50">
                    <td className="py-2 text-gray-400 capitalize">{m.key.replace(/_/g, ' ')}</td>
                    <td className="py-2 text-right text-white font-mono">
                      {m.key === 'found' ? (m.value > 0 ? '✓ Yes' : '✗ No')
                        : m.key === 'time_ms' ? `${m.value.toFixed(3)} ms`
                        : m.key.includes('ratio') ? `${(m.value * 100).toFixed(1)}%`
                        : m.value}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
        )}

        {/* Params */}
        <Card>
          <h2 className="text-lg font-semibold text-white mb-3">Parameters</h2>
          <pre className="text-sm text-gray-300 bg-gray-900 rounded p-3 overflow-auto">
            {JSON.stringify(run.params, null, 2)}
          </pre>
        </Card>

        {/* Grid visualization */}
        {mapData && (
          <Card className="lg:col-span-2">
            <h2 className="text-lg font-semibold text-white mb-3">Visualization</h2>
            <div className="flex gap-4 text-xs text-gray-400 mb-3">
              <span className="flex items-center gap-1"><span className="w-3 h-3 bg-blue-500 rounded inline-block" /> Path</span>
              <span className="flex items-center gap-1"><span className="w-3 h-3 bg-orange-500 rounded inline-block" /> Explored</span>
              <span className="flex items-center gap-1"><span className="w-3 h-3 bg-green-500 rounded inline-block" /> Start</span>
              <span className="flex items-center gap-1"><span className="w-3 h-3 bg-purple-500 rounded inline-block" /> Goal</span>
              <span className="flex items-center gap-1"><span className="w-3 h-3 bg-slate-700 rounded inline-block" /> Obstacle</span>
            </div>
            <GridViewer
              width={mapData.width}
              height={mapData.height}
              grid={mapData.grid}
              artifacts={artifacts}
            />
          </Card>
        )}

        {/* LLM Review */}
        <Card className="lg:col-span-2">
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-lg font-semibold text-white">LLM Review</h2>
            <Button
              size="sm"
              onClick={() => llmMutation.mutate()}
              loading={llmMutation.isPending}
              disabled={run.status !== 'SUCCEEDED'}
            >
              <Sparkles size={14} className="mr-1 inline" />
              Generate Review
            </Button>
          </div>
          {llmError && <ErrorMessage message={llmError} />}
          {llmReport && (
            <div className="flex flex-col gap-4">
              <p className="text-gray-300">{llmReport.summary}</p>
              {llmReport.recommendations?.length > 0 && (
                <div>
                  <p className="text-sm font-medium text-gray-400 mb-2">Recommendations:</p>
                  <ul className="list-disc list-inside text-sm text-gray-300 space-y-1">
                    {llmReport.recommendations.map((r, i) => (
                      <li key={i}>{r}</li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}
          {!llmReport && !llmError && (
            <p className="text-gray-500 text-sm">Click "Generate Review" to get AI analysis.</p>
          )}
        </Card>
      </div>
    </div>
  );
}
