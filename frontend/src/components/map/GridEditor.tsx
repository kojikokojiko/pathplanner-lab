import { useState, useCallback, useEffect } from 'react';
import clsx from 'clsx';

type EditMode = 'obstacle' | 'start' | 'goal';

interface Props {
  width: number;
  height: number;
  initialGrid?: string;
  onGridChange: (grid: string, start: { x: number; y: number } | null, goal: { x: number; y: number } | null) => void;
  readOnly?: boolean;
}

type CellType = 'empty' | 'obstacle' | 'start' | 'goal';

function parseGrid(gridText: string, width: number, height: number): { cells: CellType[][], start: { x: number; y: number } | null, goal: { x: number; y: number } | null } {
  const cells: CellType[][] = Array.from({ length: height }, () =>
    Array.from({ length: width }, () => 'empty' as CellType)
  );
  let start: { x: number; y: number } | null = null;
  let goal: { x: number; y: number } | null = null;

  const lines = gridText.split('\n');
  for (let y = 0; y < Math.min(height, lines.length); y++) {
    for (let x = 0; x < Math.min(width, lines[y].length); x++) {
      const ch = lines[y][x];
      if (ch === '#') cells[y][x] = 'obstacle';
      else if (ch === 'S') { cells[y][x] = 'start'; start = { x, y }; }
      else if (ch === 'G') { cells[y][x] = 'goal'; goal = { x, y }; }
    }
  }
  return { cells, start, goal };
}

function cellsToGrid(cells: CellType[][], width: number, height: number): string {
  const lines: string[] = [];
  for (let y = 0; y < height; y++) {
    let line = '';
    for (let x = 0; x < width; x++) {
      const c = cells[y]?.[x] ?? 'empty';
      if (c === 'obstacle') line += '#';
      else if (c === 'start') line += 'S';
      else if (c === 'goal') line += 'G';
      else line += '.';
    }
    lines.push(line);
  }
  return lines.join('\n');
}

const cellColors: Record<CellType, string> = {
  empty: 'bg-gray-100 hover:bg-gray-200',
  obstacle: 'bg-slate-700 hover:bg-slate-600',
  start: 'bg-green-500 hover:bg-green-400',
  goal: 'bg-red-500 hover:bg-red-400',
};

const CELL_SIZE = 18;
const MAX_DISPLAY_WIDTH = 900;

export default function GridEditor({ width, height, initialGrid, onGridChange, readOnly = false }: Props) {
  const [cells, setCells] = useState<CellType[][]>(() => {
    if (initialGrid) {
      return parseGrid(initialGrid, width, height).cells;
    }
    return Array.from({ length: height }, () => Array.from({ length: width }, () => 'empty' as CellType));
  });
  const [start, setStart] = useState<{ x: number; y: number } | null>(() => {
    if (initialGrid) return parseGrid(initialGrid, width, height).start;
    return null;
  });
  const [goal, setGoal] = useState<{ x: number; y: number } | null>(() => {
    if (initialGrid) return parseGrid(initialGrid, width, height).goal;
    return null;
  });
  const [mode, setMode] = useState<EditMode>('obstacle');
  const [isDrawing, setIsDrawing] = useState(false);
  const [drawValue, setDrawValue] = useState<CellType>('obstacle');

  useEffect(() => {
    if (initialGrid) {
      const parsed = parseGrid(initialGrid, width, height);
      setCells(parsed.cells);
      setStart(parsed.start);
      setGoal(parsed.goal);
    }
  }, [initialGrid, width, height]);

  const handleCellInteract = useCallback((x: number, y: number, isStart: boolean) => {
    if (readOnly) return;
    setCells(prev => {
      const next = prev.map(row => [...row]);
      if (mode === 'obstacle') {
        const newVal: CellType = isStart ? 'obstacle' : (next[y][x] === 'obstacle' ? 'empty' : 'empty');
        if (isStart) {
          next[y][x] = 'obstacle';
        } else {
          next[y][x] = drawValue;
        }
        const grid = cellsToGrid(next, width, height);
        onGridChange(grid, start, goal);
        return next;
      } else if (mode === 'start') {
        // Clear previous start
        if (start) next[start.y][start.x] = 'empty';
        next[y][x] = 'start';
        const newStart = { x, y };
        setStart(newStart);
        const grid = cellsToGrid(next, width, height);
        onGridChange(grid, newStart, goal);
        return next;
      } else if (mode === 'goal') {
        if (goal) next[goal.y][goal.x] = 'empty';
        next[y][x] = 'goal';
        const newGoal = { x, y };
        setGoal(newGoal);
        const grid = cellsToGrid(next, width, height);
        onGridChange(grid, start, newGoal);
        return next;
      }
      return next;
    });
  }, [mode, start, goal, width, height, onGridChange, readOnly, drawValue]);

  const handleMouseDown = (x: number, y: number) => {
    if (readOnly || mode !== 'obstacle') return;
    const newVal: CellType = cells[y][x] === 'obstacle' ? 'empty' : 'obstacle';
    setDrawValue(newVal);
    setIsDrawing(true);
    handleCellInteract(x, y, true);
  };

  const handleMouseEnter = (x: number, y: number) => {
    if (!isDrawing || readOnly || mode !== 'obstacle') return;
    setCells(prev => {
      const next = prev.map(row => [...row]);
      if (next[y][x] !== 'start' && next[y][x] !== 'goal') {
        next[y][x] = drawValue;
      }
      const grid = cellsToGrid(next, width, height);
      onGridChange(grid, start, goal);
      return next;
    });
  };

  const cellSize = Math.min(CELL_SIZE, Math.floor(MAX_DISPLAY_WIDTH / width));

  return (
    <div className="flex flex-col gap-4">
      {!readOnly && (
        <div className="flex gap-2">
          {(['obstacle', 'start', 'goal'] as EditMode[]).map((m) => (
            <button
              key={m}
              onClick={() => setMode(m)}
              className={clsx(
                'px-3 py-1.5 rounded text-sm font-medium capitalize transition-colors',
                mode === m
                  ? 'bg-indigo-600 text-white'
                  : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
              )}
            >
              {m === 'obstacle' ? 'Draw Obstacle' : m === 'start' ? 'Set Start' : 'Set Goal'}
            </button>
          ))}
          <span className="ml-4 text-sm text-gray-400 self-center">
            {start ? `S: (${start.x},${start.y})` : 'No Start'}
            {' | '}
            {goal ? `G: (${goal.x},${goal.y})` : 'No Goal'}
          </span>
        </div>
      )}
      <div
        className="overflow-auto border border-gray-600 rounded"
        onMouseUp={() => setIsDrawing(false)}
        onMouseLeave={() => setIsDrawing(false)}
      >
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: `repeat(${width}, ${cellSize}px)`,
            gap: '1px',
            backgroundColor: '#374151',
            padding: '1px',
          }}
        >
          {cells.map((row, y) =>
            row.map((cell, x) => (
              <div
                key={`${x}-${y}`}
                style={{ width: cellSize, height: cellSize }}
                className={clsx(
                  'cursor-pointer transition-colors',
                  cellColors[cell],
                  readOnly && 'cursor-default',
                )}
                onMouseDown={() => mode === 'obstacle' ? handleMouseDown(x, y) : handleCellInteract(x, y, false)}
                onMouseEnter={() => handleMouseEnter(x, y)}
              />
            ))
          )}
        </div>
      </div>
    </div>
  );
}
