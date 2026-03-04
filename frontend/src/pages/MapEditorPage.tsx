import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Save, ArrowLeft } from 'lucide-react';
import { mapsApi } from '../api/maps';
import GridEditor from '../components/map/GridEditor';
import Button from '../components/ui/Button';
import Input from '../components/ui/Input';
import LoadingSpinner from '../components/ui/LoadingSpinner';

export default function MapEditorPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const isNew = !id;

  const [name, setName] = useState('');
  const [width, setWidth] = useState(20);
  const [height, setHeight] = useState(15);
  const [grid, setGrid] = useState('');
  const [start, setStart] = useState<{ x: number; y: number } | null>(null);
  const [goal, setGoal] = useState<{ x: number; y: number } | null>(null);
  const [configDone, setConfigDone] = useState(!isNew);
  const [error, setError] = useState('');

  const { data: existingMap, isLoading } = useQuery({
    queryKey: ['maps', id],
    queryFn: () => mapsApi.get(id!),
    enabled: !!id,
  });

  useEffect(() => {
    if (existingMap) {
      setName(existingMap.name);
      setWidth(existingMap.width);
      setHeight(existingMap.height);
      setGrid(existingMap.grid);
    }
  }, [existingMap]);

  const saveMutation = useMutation({
    mutationFn: () => {
      const payload = { name, width, height, grid };
      return id ? mapsApi.update(id, payload) : mapsApi.create(payload);
    },
    onSuccess: () => navigate('/maps'),
    onError: () => setError('Failed to save map'),
  });

  if (!isNew && isLoading) {
    return <div className="flex justify-center py-16"><LoadingSpinner size="lg" /></div>;
  }

  const handleGridChange = (g: string, s: { x: number; y: number } | null, gl: { x: number; y: number } | null) => {
    setGrid(g);
    setStart(s);
    setGoal(gl);
  };

  const handleSave = () => {
    if (!name.trim()) { setError('Name is required'); return; }
    if (!start) { setError('Please set a start point (S)'); return; }
    if (!goal) { setError('Please set a goal point (G)'); return; }
    setError('');
    saveMutation.mutate();
  };

  return (
    <div>
      <div className="flex items-center gap-4 mb-6">
        <button onClick={() => navigate('/maps')} className="text-gray-400 hover:text-white">
          <ArrowLeft size={20} />
        </button>
        <h1 className="text-2xl font-bold text-white">{isNew ? 'New Map' : 'Edit Map'}</h1>
      </div>

      <div className="flex flex-col gap-6">
        <div className="bg-gray-800 rounded-lg border border-gray-700 p-4 flex flex-wrap gap-4 items-end">
          <div className="flex-1 min-w-40">
            <Input
              label="Map Name"
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder="e.g. maze-01"
            />
          </div>
          {isNew && !configDone && (
            <>
              <div className="w-28">
                <Input
                  label="Width"
                  type="number"
                  min={2}
                  max={200}
                  value={width}
                  onChange={e => setWidth(parseInt(e.target.value) || 20)}
                />
              </div>
              <div className="w-28">
                <Input
                  label="Height"
                  type="number"
                  min={2}
                  max={200}
                  value={height}
                  onChange={e => setHeight(parseInt(e.target.value) || 15)}
                />
              </div>
              <Button onClick={() => setConfigDone(true)}>Create Grid</Button>
            </>
          )}
          {configDone && (
            <div className="text-sm text-gray-400">
              {width} × {height}
              {start && ` | S: (${start.x},${start.y})`}
              {goal && ` | G: (${goal.x},${goal.y})`}
            </div>
          )}
        </div>

        {configDone && (
          <div className="bg-gray-800 rounded-lg border border-gray-700 p-4">
            <GridEditor
              width={width}
              height={height}
              initialGrid={grid || undefined}
              onGridChange={handleGridChange}
            />
          </div>
        )}

        {error && <p className="text-sm text-red-400">{error}</p>}

        <div className="flex gap-3">
          <Button onClick={handleSave} loading={saveMutation.isPending} disabled={!configDone}>
            <Save size={16} className="mr-1 inline" />
            Save Map
          </Button>
          <Button variant="secondary" onClick={() => navigate('/maps')}>
            Cancel
          </Button>
        </div>
      </div>
    </div>
  );
}
