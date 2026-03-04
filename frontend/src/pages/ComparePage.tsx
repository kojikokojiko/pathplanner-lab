import { useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Sparkles } from 'lucide-react';
import { compareApi } from '../api/compare';
import Button from '../components/ui/Button';
import Card from '../components/ui/Card';
import Badge from '../components/ui/Badge';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorMessage from '../components/ui/ErrorMessage';
import type { LLMReviewResponse } from '../types';

type SortKey = 'algorithm' | 'path_length' | 'expanded_nodes' | 'time_ms' | 'turns' | 'explored_ratio';
type SortDir = 'asc' | 'desc';

export default function ComparePage() {
  const [searchParams] = useSearchParams();
  const runIdsParam = searchParams.get('runs') ?? '';
  const runIds = runIdsParam ? runIdsParam.split(',').filter(Boolean) : [];

  const [manualIds, setManualIds] = useState('');
  const [activeIds, setActiveIds] = useState(runIds);
  const [sortKey, setSortKey] = useState<SortKey>('time_ms');
  const [sortDir, setSortDir] = useState<SortDir>('asc');
  const [llmReport, setLlmReport] = useState<LLMReviewResponse | null>(null);
  const [llmError, setLlmError] = useState('');

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['compare', activeIds],
    queryFn: () => compareApi.compare(activeIds),
    enabled: activeIds.length > 0,
  });

  const llmMutation = useMutation({
    mutationFn: () => compareApi.llmReview(activeIds, 'teacher'),
    onSuccess: (d) => { setLlmReport(d); setLlmError(''); },
    onError: () => setLlmError('LLM review failed'),
  });

  const handleSort = (key: SortKey) => {
    if (sortKey === key) setSortDir(d => d === 'asc' ? 'desc' : 'asc');
    else { setSortKey(key); setSortDir('asc'); }
  };

  const sorted = [...(data?.table ?? [])].sort((a, b) => {
    const av = a[sortKey] as number | string;
    const bv = b[sortKey] as number | string;
    const dir = sortDir === 'asc' ? 1 : -1;
    if (typeof av === 'string') return av.localeCompare(bv as string) * dir;
    return ((av as number) - (bv as number)) * dir;
  });

  const SortHeader = ({ k, label }: { k: SortKey; label: string }) => (
    <th
      className="pb-3 pr-4 cursor-pointer hover:text-white select-none"
      onClick={() => handleSort(k)}
    >
      {label} {sortKey === k ? (sortDir === 'asc' ? '↑' : '↓') : ''}
    </th>
  );

  return (
    <div>
      <h1 className="text-2xl font-bold text-white mb-6">Compare Runs</h1>

      {activeIds.length === 0 && (
        <Card className="mb-6">
          <p className="text-sm text-gray-400 mb-3">Enter run IDs to compare (comma-separated):</p>
          <div className="flex gap-3">
            <input
              className="flex-1 bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-gray-100 text-sm"
              value={manualIds}
              onChange={e => setManualIds(e.target.value)}
              placeholder="uuid1,uuid2,..."
            />
            <Button
              size="sm"
              onClick={() => {
                const ids = manualIds.split(',').map(s => s.trim()).filter(Boolean);
                setActiveIds(ids);
              }}
            >
              Compare
            </Button>
          </div>
        </Card>
      )}

      {activeIds.length > 0 && (
        <p className="text-sm text-gray-400 mb-4">Comparing {activeIds.length} runs</p>
      )}

      {isLoading && <div className="flex justify-center py-16"><LoadingSpinner size="lg" /></div>}
      {error && <ErrorMessage message="Failed to load comparison" />}

      {data && sorted.length > 0 && (
        <Card className="mb-6 overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-700 text-gray-400 text-left">
                <th className="pb-3 pr-4">Run ID</th>
                <SortHeader k="algorithm" label="Algorithm" />
                <th className="pb-3 pr-4">Status</th>
                <th className="pb-3 pr-4">Found</th>
                <SortHeader k="path_length" label="Path Len" />
                <SortHeader k="expanded_nodes" label="Expanded" />
                <SortHeader k="time_ms" label="Time (ms)" />
                <SortHeader k="turns" label="Turns" />
                <SortHeader k="explored_ratio" label="Explored%" />
              </tr>
            </thead>
            <tbody>
              {sorted.map((row) => (
                <tr key={row.run_id} className="border-b border-gray-700/50 hover:bg-gray-700/30">
                  <td className="py-3 pr-4 font-mono text-xs text-gray-400">{row.run_id.slice(0, 8)}…</td>
                  <td className="py-3 pr-4 font-mono text-indigo-300">{row.algorithm}</td>
                  <td className="py-3 pr-4"><Badge status={row.status} /></td>
                  <td className="py-3 pr-4">{row.found ? '✓' : '✗'}</td>
                  <td className="py-3 pr-4 text-gray-300">{row.path_length || '—'}</td>
                  <td className="py-3 pr-4 text-gray-300">{row.expanded_nodes || '—'}</td>
                  <td className="py-3 pr-4 text-gray-300">{row.time_ms?.toFixed(3) || '—'}</td>
                  <td className="py-3 pr-4 text-gray-300">{row.turns || '—'}</td>
                  <td className="py-3 pr-4 text-gray-300">
                    {row.explored_ratio ? `${(row.explored_ratio * 100).toFixed(1)}%` : '—'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}

      {data && (
        <Card>
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-lg font-semibold text-white">LLM Analysis</h2>
            <Button
              size="sm"
              onClick={() => llmMutation.mutate()}
              loading={llmMutation.isPending}
            >
              <Sparkles size={14} className="mr-1 inline" />
              Analyze
            </Button>
          </div>
          {llmError && <ErrorMessage message={llmError} />}
          {llmReport && (
            <div className="flex flex-col gap-4">
              <p className="text-gray-300">{llmReport.summary}</p>
              {llmReport.recommendations?.length > 0 && (
                <ul className="list-disc list-inside text-sm text-gray-300 space-y-1">
                  {llmReport.recommendations.map((r, i) => (
                    <li key={i}>{r}</li>
                  ))}
                </ul>
              )}
            </div>
          )}
          {!llmReport && !llmError && (
            <p className="text-gray-500 text-sm">Click "Analyze" for AI-powered insights.</p>
          )}
        </Card>
      )}
    </div>
  );
}
