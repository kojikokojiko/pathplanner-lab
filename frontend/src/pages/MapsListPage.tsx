import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate } from 'react-router-dom';
import { Plus, Edit, Trash2, Map } from 'lucide-react';
import { mapsApi } from '../api/maps';
import Button from '../components/ui/Button';
import Card from '../components/ui/Card';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import ErrorMessage from '../components/ui/ErrorMessage';

export default function MapsListPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data, isLoading, error } = useQuery({
    queryKey: ['maps'],
    queryFn: () => mapsApi.list(),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => mapsApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['maps'] });
    },
  });

  if (isLoading) {
    return <div className="flex justify-center py-16"><LoadingSpinner size="lg" /></div>;
  }

  if (error) {
    return <ErrorMessage message="Failed to load maps" />;
  }

  const maps = data?.items ?? [];

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-white">Maps</h1>
        <Button onClick={() => navigate('/maps/new')}>
          <Plus size={16} className="mr-1 inline" />
          New Map
        </Button>
      </div>

      {maps.length === 0 ? (
        <div className="text-center py-16 text-gray-400">
          <Map size={48} className="mx-auto mb-4 opacity-30" />
          <p>No maps yet. Create your first map!</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {maps.map((m) => (
            <Card key={m.id} className="flex flex-col gap-3">
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="font-semibold text-white">{m.name}</h3>
                  <p className="text-sm text-gray-400">{m.width} × {m.height}</p>
                </div>
                <span className="text-xs text-gray-500 bg-gray-700 px-2 py-0.5 rounded">
                  {(m.obstacle_ratio * 100).toFixed(1)}% obstacles
                </span>
              </div>
              <p className="text-xs text-gray-500">
                {new Date(m.created_at).toLocaleDateString()}
              </p>
              <div className="flex gap-2 mt-auto pt-2 border-t border-gray-700">
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => navigate(`/maps/${m.id}`)}
                  className="flex-1"
                >
                  <Edit size={14} className="mr-1 inline" />
                  Edit
                </Button>
                <Button
                  variant="danger"
                  size="sm"
                  loading={deleteMutation.isPending}
                  onClick={() => {
                    if (confirm(`Delete "${m.name}"?`)) {
                      deleteMutation.mutate(m.id);
                    }
                  }}
                >
                  <Trash2 size={14} />
                </Button>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
