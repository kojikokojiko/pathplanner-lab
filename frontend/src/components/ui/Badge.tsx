import clsx from 'clsx';
import type { ExperimentStatus, RunStatus } from '../../types';

type Status = ExperimentStatus | RunStatus | string;

const statusColors: Record<string, string> = {
  PENDING: 'bg-yellow-900 text-yellow-300',
  RUNNING: 'bg-blue-900 text-blue-300',
  COMPLETED: 'bg-green-900 text-green-300',
  SUCCEEDED: 'bg-green-900 text-green-300',
  FAILED: 'bg-red-900 text-red-300',
  CANCELLED: 'bg-gray-700 text-gray-400',
};

interface Props {
  status: Status;
  className?: string;
}

export default function Badge({ status, className }: Props) {
  return (
    <span className={clsx(
      'inline-flex items-center px-2 py-0.5 rounded text-xs font-medium',
      statusColors[status] ?? 'bg-gray-700 text-gray-300',
      className,
    )}>
      {status}
    </span>
  );
}
