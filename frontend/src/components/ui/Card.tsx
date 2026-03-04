import { ReactNode } from 'react';
import clsx from 'clsx';

interface Props {
  children: ReactNode;
  className?: string;
}

export default function Card({ children, className }: Props) {
  return (
    <div className={clsx('bg-gray-800 rounded-lg border border-gray-700 p-4', className)}>
      {children}
    </div>
  );
}
