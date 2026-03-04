import clsx from 'clsx';

interface Props {
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

const sizeClasses = {
  sm: 'w-4 h-4 border-2',
  md: 'w-8 h-8 border-2',
  lg: 'w-12 h-12 border-4',
};

export default function LoadingSpinner({ size = 'md', className }: Props) {
  return (
    <div className={clsx(
      'rounded-full border-indigo-500 border-t-transparent animate-spin',
      sizeClasses[size],
      className,
    )} />
  );
}
