import { useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { ArrowLeft, GitCompare } from 'lucide-react';
import { experimentsApi } from '../api/experiments';
import Badge from '../components/ui/Badge';
import Card from '../components/ui/Card';
import Button from '../components/ui/Button';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorMessage from '../components/ui/ErrorMessage';

export default function ExperimentDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [selectedRuns, setSelectedRuns] = useState<Set<string>>(new Set());

  const isPolling = (status?: string) =>
    status === 'PENDING' || status === 'RUNNING';

  const { data: exp, isLoading, error } = useQuery({
    queryKey: ['experiments', id],
    queryFn: () => experimentsApi.get(id!),
    refetchInterval: (q) => isPolling(q.state.data?.status) ? 3000 : false,
    enabled: !!id,
  });

  const { data: runsData, isLoading: runsLoading } = useQuery({
    queryKey: ['experiments', id, 'runs'],
    queryFn: () => experimentsApi.getRuns(id!),
    refetchInterval: isPolling(exp?.status) ? 3000 : false,
    enabled: !!id,
  });

  if (isLoading) return <div className="flex justify-center py-16"><LoadingSpinner size="lg" /></div>;
  if (error || !exp) return <ErrorMessage message="Experiment not found" />;

  const runs = runsData?.items ?? [];

  const toggleRun = (runId: string) => {
    setSelectedRuns(prev => {
      const next = new Set(prev);
      if (next.has(runId)) next.delete(runId);
      else next.add(runId);
      return next;
    });
  };

  const handleCompare = () => {
    if (selectedRuns.size < 2) return;
    navigate(`/compare?runs=${Array.from(selectedRuns).join(',')}`);
  };

  return (
    <div>
      <div className="flex items-center gap-4 mb-6">
        <button onClick={() => navigate('/experiments')} className="text-gray-400 hover:text-white">
          <ArrowLeft size={20} />
        </button>
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold text-white">{exp.name || 'Experiment'}</h1>
            <Badge status={exp.status} />
          </div>
          <p className="text-sm text-gray-400">
            Start: ({exp.start.x},{exp.start.y}) → Goal: ({exp.goal.x},{exp.goal.y})
          </p>
        </div>
      </div>

      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-white">Runs ({runs.length})</h2>
        {selectedRuns.size >= 2 && (
          <Button size="sm" onClick={handleCompare}>
            <GitCompare size={14} className="mr-1 inline" />
            Compare Selected ({selectedRuns.size})
          </Button>
        )}
      </div>

      {runsLoading ? (
        <LoadingSpinner />
      ) : runs.length === 0 ? (
        <p className="text-gray-400">No runs yet.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-700 text-gray-400 text-left">
                <th className="pb-3 pr-4 w-8"></th>
                <th className="pb-3 pr-4">Algorithm</th>
                <th className="pb-3 pr-4">Status</th>
                <th className="pb-3 pr-4">Path Length</th>
                <th className="pb-3 pr-4">Expanded</th>
                <th className="pb-3 pr-4">Time (ms)</th>
                <th className="pb-3">Actions</th>
              </tr>
            </thead>
            <tbody>
              {runs.map((run) => (
                <tr key={run.id} className="border-b border-gray-700/50 hover:bg-gray-700/30">
                  <td className="py-3 pr-4">
                    {run.status === 'SUCCEEDED' && (
                      <input
                        type="checkbox"
                        checked={selectedRuns.has(run.id)}
                        onChange={() => toggleRun(run.id)}
                        className="accent-indigo-500"
                      />
                    )}
                  </td>
                  <td className="py-3 pr-4 font-mono text-indigo-300">{run.algorithm_key}</td>
                  <td className="py-3 pr-4"><Badge status={run.status} /></td>
                  <td className="py-3 pr-4 text-gray-300">
                    {run.metrics?.find(m => m.key === 'path_length')?.value ?? '—'}
                  </td>
                  <td className="py-3 pr-4 text-gray-300">
                    {run.metrics?.find(m => m.key === 'expanded_nodes')?.value ?? '—'}
                  </td>
                  <td className="py-3 pr-4 text-gray-300">
                    {run.metrics?.find(m => m.key === 'time_ms')?.value?.toFixed(2) ?? '—'}
                  </td>
                  <td className="py-3">
                    <Link to={`/runs/${run.id}`} className="text-indigo-400 hover:text-indigo-300 text-xs">
                      View →
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
