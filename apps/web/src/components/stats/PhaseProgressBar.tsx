export interface PhaseProgress {
  id: string;
  name: string;
  total: number;
  completed: number;
}

interface PhaseProgressBarProps {
  phase: PhaseProgress;
}

function barColor(pct: number): string {
  if (pct >= 75) return "bg-green-500 dark:bg-green-400";
  if (pct >= 25) return "bg-yellow-500 dark:bg-yellow-400";
  return "bg-gray-400 dark:bg-gray-500";
}

export function PhaseProgressBar({ phase }: PhaseProgressBarProps) {
  const pct = phase.total > 0 ? Math.round((phase.completed / phase.total) * 100) : 0;

  return (
    <div className="space-y-1">
      <div className="flex justify-between items-center text-sm">
        <span className="font-medium text-gray-700 dark:text-gray-200">{phase.name}</span>
        <span className="text-gray-500 dark:text-gray-400">
          {phase.completed} / {phase.total} tasks ({pct}%)
        </span>
      </div>
      <div className="h-2 rounded-full bg-gray-200 dark:bg-gray-700 overflow-hidden">
        <div
          role="progressbar"
          aria-label={`${phase.name} progress`}
          aria-valuenow={pct}
          aria-valuemin={0}
          aria-valuemax={100}
          className={`h-full rounded-full transition-all ${barColor(pct)}`}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}
