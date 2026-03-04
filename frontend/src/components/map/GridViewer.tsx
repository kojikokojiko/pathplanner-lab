import { useMemo } from 'react';
import type { RunArtifact } from '../../types';

interface Props {
  width: number;
  height: number;
  grid: string;
  artifacts?: RunArtifact[];
}

const CELL_SIZE = 16;
const MAX_WIDTH = 800;

export default function GridViewer({ width, height, grid, artifacts }: Props) {
  const cellSize = Math.min(CELL_SIZE, Math.floor(MAX_WIDTH / width));

  const { path, heatmap, start, goal } = useMemo(() => {
    const pathArtifact = artifacts?.find(a => a.kind === 'PATH');
    const heatmapArtifact = artifacts?.find(a => a.kind === 'EXPANSION_HEATMAP');

    type PathPoint = { x: number; y: number };
    const path = pathArtifact ? (pathArtifact.data as PathPoint[]) : [];
    const heatmapRaw = heatmapArtifact ? (heatmapArtifact.data as Record<string, number>) : {};

    const pathSet = new Set(path.map((p: PathPoint) => `${p.x},${p.y}`));
    const maxOrder = Math.max(0, ...Object.values(heatmapRaw));

    // Find start/goal from grid
    let start: { x: number; y: number } | null = null;
    let goal: { x: number; y: number } | null = null;
    const lines = grid.split('\n');
    for (let y = 0; y < lines.length; y++) {
      for (let x = 0; x < lines[y].length; x++) {
        if (lines[y][x] === 'S') start = { x, y };
        if (lines[y][x] === 'G') goal = { x, y };
      }
    }

    return { path: pathSet, heatmap: heatmapRaw, maxOrder, start, goal };
  }, [artifacts, grid]);

  const lines = grid.split('\n');

  const getCellColor = (x: number, y: number, ch: string) => {
    const key = `${x},${y}`;
    if (ch === '#') return '#334155';
    if (ch === 'S' || (start && x === start.x && y === start.y)) return '#22c55e';
    if (ch === 'G' || (goal && x === goal.x && y === goal.y)) return '#ef4444';
    if (path.has(key)) return '#3b82f6';
    const order = heatmap[key];
    if (order !== undefined) {
      const { maxOrder } = { maxOrder: Math.max(0, ...Object.values(heatmap)) };
      const intensity = maxOrder > 0 ? order / maxOrder : 0;
      const r = Math.round(255 * intensity);
      const g = Math.round(100 * (1 - intensity));
      return `rgb(${r}, ${g}, 50)`;
    }
    return '#f3f4f6';
  };

  return (
    <div className="overflow-auto border border-gray-600 rounded">
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: `repeat(${width}, ${cellSize}px)`,
          gap: '1px',
          backgroundColor: '#374151',
          padding: '1px',
        }}
      >
        {Array.from({ length: height }, (_, y) =>
          Array.from({ length: width }, (_, x) => {
            const ch = lines[y]?.[x] ?? '.';
            return (
              <div
                key={`${x}-${y}`}
                style={{
                  width: cellSize,
                  height: cellSize,
                  backgroundColor: getCellColor(x, y, ch),
                }}
              />
            );
          })
        )}
      </div>
    </div>
  );
}
