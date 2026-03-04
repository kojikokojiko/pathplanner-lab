import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Plus, FlaskConical } from 'lucide-react';
import { experimentsApi, CreateExperimentInput } from '../api/experiments';
import { mapsApi } from '../api/maps';
import type { AlgorithmKey, RunInput } from '../types';
import Button from '../components/ui/Button';
import Card from '../components/ui/Card';
import Badge from '../components/ui/Badge';
import Input from '../components/ui/Input';
import LoadingSpinner from '../components/ui/LoadingSpinner';

function parseStartGoalFromGrid(grid: string): { start: { x: number; y: number } | null; goal: { x: number; y: number } | null } {
  let start = null, goal = null;
  grid.split('\n').forEach((row, y) => {
    [...row].forEach((ch, x) => {
      if (ch === 'S') start = { x, y };
      if (ch === 'G') goal = { x, y };
    });
  });
  return { start, goal };
}

const ALGORITHMS: { key: AlgorithmKey; label: string }[] = [
  { key: 'BFS', label: 'BFS' },
  { key: 'DIJKSTRA', label: 'Dijkstra' },
  { key: 'ASTAR', label: 'A*' },
  { key: 'WEIGHTED_ASTAR', label: 'Weighted A*' },
  { key: 'GREEDY', label: 'Greedy' },
];

function NewExperimentModal({ onClose }: { onClose: () => void }) {
  const queryClient = useQueryClient();
  const [name, setName] = useState('');
  const [mapId, setMapId] = useState('');
  const [startX, setStartX] = useState(0);
  const [startY, setStartY] = useState(0);
  const [goalX, setGoalX] = useState(10);
  const [goalY, setGoalY] = useState(10);
  const [selectedAlgos, setSelectedAlgos] = useState<Set<AlgorithmKey>>(new Set(['ASTAR']));
  const [weightedW, setWeightedW] = useState(1.5);
  const [error, setError] = useState('');

  const { data: mapsData } = useQuery({ queryKey: ['maps'], queryFn: () => mapsApi.list() });

  const createMutation = useMutation({
    mutationFn: (input: CreateExperimentInput) => experimentsApi.create(input, crypto.randomUUID()),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['experiments'] });
      onClose();
    },
    onError: () => setError('Failed to create experiment'),
  });

  const toggleAlgo = (key: AlgorithmKey) => {
    setSelectedAlgos(prev => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });
  };

  const handleCreate = () => {
    if (!mapId) { setError('Select a map'); return; }
    if (selectedAlgos.size === 0) { setError('Select at least one algorithm'); return; }
    setError('');
    const runs: RunInput[] = Array.from(selectedAlgos).map(key => {
      const params: Record<string, unknown> = { neighbors: 4, heuristic: 'manhattan' };
      if (key === 'WEIGHTED_ASTAR') params.w = weightedW;
      return { algorithmKey: key, params };
    });
    createMutation.mutate({
      mapId,
      name: name || `Experiment ${new Date().toLocaleTimeString()}`,
      start: { x: startX, y: startY },
      goal: { x: goalX, y: goalY },
      runs,
      seed: 42,
    });
  };

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
      <div className="bg-gray-800 rounded-xl border border-gray-700 p-6 w-full max-w-md">
        <h2 className="text-xl font-bold text-white mb-4">New Experiment</h2>
        <div className="flex flex-col gap-3">
          <Input label="Name (optional)" value={name} onChange={e => setName(e.target.value)} placeholder="Auto-generated if empty" />
          <div>
            <label className="text-sm font-medium text-gray-300 block mb-1">Map</label>
            <select
              className="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-gray-100"
              value={mapId}
              onChange={e => {
                const id = e.target.value;
                setMapId(id);
                const map = mapsData?.items?.find(m => m.id === id);
                if (map) {
                  const { start, goal } = parseStartGoalFromGrid(map.grid);
                  if (start) { setStartX(start.x); setStartY(start.y); }
                  if (goal)  { setGoalX(goal.x);   setGoalY(goal.y); }
                }
              }}
            >
              <option value="">-- Select Map --</option>
              {mapsData?.items?.map(m => (
                <option key={m.id} value={m.id}>{m.name} ({m.width}x{m.height})</option>
              ))}
            </select>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <Input label="Start X" type="number" min={0} value={startX} onChange={e => setStartX(parseInt(e.target.value) || 0)} />
            <Input label="Start Y" type="number" min={0} value={startY} onChange={e => setStartY(parseInt(e.target.value) || 0)} />
            <Input label="Goal X" type="number" min={0} value={goalX} onChange={e => setGoalX(parseInt(e.target.value) || 0)} />
            <Input label="Goal Y" type="number" min={0} value={goalY} onChange={e => setGoalY(parseInt(e.target.value) || 0)} />
          </div>
          <div>
            <label className="text-sm font-medium text-gray-300 block mb-2">Algorithms</label>
            <div className="flex flex-wrap gap-2">
              {ALGORITHMS.map(({ key, label }) => (
                <button
                  key={key}
                  onClick={() => toggleAlgo(key)}
                  className={`px-3 py-1.5 rounded text-sm font-medium transition-colors ${
                    selectedAlgos.has(key) ? 'bg-indigo-600 text-white' : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                  }`}
                >
                  {label}
                </button>
              ))}
            </div>
          </div>
          {selectedAlgos.has('WEIGHTED_ASTAR') && (
            <div>
              <label className="text-sm font-medium text-gray-300 block mb-1">
                Weighted A* — w: {weightedW.toFixed(1)}
              </label>
              <input
                type="range" min={1.0} max={3.0} step={0.1}
                value={weightedW}
                onChange={e => setWeightedW(parseFloat(e.target.value))}
                className="w-full accent-indigo-500"
              />
            </div>
          )}
          {error && <p className="text-sm text-red-400">{error}</p>}
          <div className="flex gap-3 mt-2">
            <Button onClick={handleCreate} loading={createMutation.isPending} className="flex-1">
              Create & Run
            </Button>
            <Button variant="secondary" onClick={onClose}>Cancel</Button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function ExperimentsListPage() {
  const navigate = useNavigate();
  const [showModal, setShowModal] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ['experiments'],
    queryFn: () => experimentsApi.list(),
    refetchInterval: 5000,
  });

  if (isLoading) {
    return <div className="flex justify-center py-16"><LoadingSpinner size="lg" /></div>;
  }

  const experiments = data?.items ?? [];

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-white">Experiments</h1>
        <Button onClick={() => setShowModal(true)}>
          <Plus size={16} className="mr-1 inline" />
          New Experiment
        </Button>
      </div>

      {experiments.length === 0 ? (
        <div className="text-center py-16 text-gray-400">
          <FlaskConical size={48} className="mx-auto mb-4 opacity-30" />
          <p>No experiments yet.</p>
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          {experiments.map((exp) => (
            <Card key={exp.id} className="cursor-pointer hover:border-indigo-500 transition-colors" >
              <div className="flex items-center justify-between" onClick={() => navigate(`/experiments/${exp.id}`)}>
                <div>
                  <div className="flex items-center gap-3">
                    <h3 className="font-semibold text-white">{exp.name || exp.id.slice(0, 8)}</h3>
                    <Badge status={exp.status} />
                  </div>
                  <p className="text-sm text-gray-400 mt-1">
                    S: ({exp.start.x},{exp.start.y}) → G: ({exp.goal.x},{exp.goal.y})
                  </p>
                </div>
                <p className="text-xs text-gray-500">{new Date(exp.created_at).toLocaleString()}</p>
              </div>
            </Card>
          ))}
        </div>
      )}

      {showModal && <NewExperimentModal onClose={() => setShowModal(false)} />}
    </div>
  );
}
