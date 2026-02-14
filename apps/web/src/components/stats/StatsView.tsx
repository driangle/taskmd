import type { Stats } from "../../api/types.ts";

interface StatsViewProps {
  stats: Stats;
}

export function StatsView({ stats }: StatsViewProps) {
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <MetricCard label="Total Tasks" value={stats.total_tasks} />
        <MetricCard label="Blocked" value={stats.blocked_tasks_count} />
        <MetricCard
          label="Critical Path"
          value={stats.critical_path_length}
        />
        <MetricCard
          label="Avg Deps/Task"
          value={stats.avg_dependencies_per_task.toFixed(1)}
        />
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <BreakdownCard title="By Status" data={stats.tasks_by_status} />
        <BreakdownCard title="By Priority" data={stats.tasks_by_priority} />
        <BreakdownCard title="By Effort" data={stats.tasks_by_effort} />
      </div>

      <div className="bg-white rounded-lg border border-gray-200 p-4">
        <h3 className="text-sm font-semibold text-gray-700 mb-3">Tags</h3>
        {!stats.tags_by_count || stats.tags_by_count.length === 0 ? (
          <p className="text-xs text-gray-400">No tags found</p>
        ) : (
          <div className="space-y-2">
            {stats.tags_by_count.map(({ tag, count }) => (
              <div key={tag} className="flex justify-between items-center">
                <span className="text-sm text-gray-600">{tag}</span>
                <span className="text-sm font-medium">{count}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function MetricCard({
  label,
  value,
}: {
  label: string;
  value: number | string;
}) {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4">
      <p className="text-xs text-gray-500 uppercase tracking-wider">{label}</p>
      <p className="mt-1 text-2xl font-semibold">{value}</p>
    </div>
  );
}

function BreakdownCard({
  title,
  data,
}: {
  title: string;
  data: Record<string, number>;
}) {
  const entries = Object.entries(data).filter(([, v]) => v > 0);
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4">
      <h3 className="text-sm font-semibold text-gray-700 mb-3">{title}</h3>
      {entries.length === 0 ? (
        <p className="text-xs text-gray-400">No data</p>
      ) : (
        <div className="space-y-2">
          {entries.map(([key, val]) => (
            <div key={key} className="flex justify-between items-center">
              <span className="text-sm text-gray-600">{key}</span>
              <span className="text-sm font-medium">{val}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
