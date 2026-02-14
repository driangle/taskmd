import { Link } from "react-router-dom";
import type { ValidationResult, ValidationIssue } from "../../api/types.ts";

interface ValidateViewProps {
  result: ValidationResult;
}

export function ValidateView({ result }: ValidateViewProps) {
  const { issues, errors, warnings } = result;

  const grouped = groupByFile(issues);

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <SummaryCard label="Errors" value={errors} color="red" />
        <SummaryCard label="Warnings" value={warnings} color="yellow" />
      </div>

      {issues.length === 0 ? (
        <div className="bg-white rounded-lg border border-gray-200 p-8 text-center dark:bg-gray-800 dark:border-gray-700">
          <p className="text-sm text-green-600 dark:text-green-400 font-medium">
            All tasks are valid
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          {Object.entries(grouped).map(([filePath, fileIssues]) => (
            <div
              key={filePath}
              className="bg-white rounded-lg border border-gray-200 p-4 dark:bg-gray-800 dark:border-gray-700"
            >
              <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-200 mb-3 font-mono">
                {filePath}
              </h3>
              <div className="space-y-2">
                {fileIssues.map((issue, i) => (
                  <IssueRow key={i} issue={issue} />
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function SummaryCard({
  label,
  value,
  color,
}: {
  label: string;
  value: number;
  color: "red" | "yellow";
}) {
  const textColor = color === "red"
    ? "text-red-600 dark:text-red-400"
    : "text-yellow-600 dark:text-yellow-400";
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4 dark:bg-gray-800 dark:border-gray-700">
      <p className="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wider">{label}</p>
      <p className={`mt-1 text-2xl font-semibold ${textColor}`}>{value}</p>
    </div>
  );
}

function IssueRow({ issue }: { issue: ValidationIssue }) {
  const dotColor =
    issue.level === "error" ? "bg-red-500" : "bg-yellow-400";

  return (
    <div className="flex items-start gap-2 text-sm">
      <span
        className={`mt-1.5 h-2 w-2 rounded-full flex-shrink-0 ${dotColor}`}
      />
      {issue.task_id && (
        <Link
          to={`/tasks/${issue.task_id}`}
          className="font-mono text-blue-600 hover:underline dark:text-blue-400 flex-shrink-0"
        >
          {issue.task_id}
        </Link>
      )}
      <span className="text-gray-700 dark:text-gray-300">{issue.message}</span>
    </div>
  );
}

function groupByFile(
  issues: ValidationIssue[],
): Record<string, ValidationIssue[]> {
  const groups: Record<string, ValidationIssue[]> = {};
  for (const issue of issues) {
    const key = issue.file_path ?? "(general)";
    if (!groups[key]) groups[key] = [];
    groups[key].push(issue);
  }
  return groups;
}
