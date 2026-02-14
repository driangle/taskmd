import { useState } from "react";

const STATUS_ITEMS = [
  { label: "Pending", bg: "bg-yellow-50", border: "border-yellow-300" },
  { label: "In Progress", bg: "bg-blue-50", border: "border-blue-300" },
  { label: "Completed", bg: "bg-green-50", border: "border-green-300" },
  { label: "Blocked", bg: "bg-red-50", border: "border-red-300" },
  { label: "Cancelled", bg: "bg-gray-50", border: "border-gray-300" },
];

const PRIORITY_ITEMS = [
  { label: "Critical", color: "bg-red-500" },
  { label: "High", color: "bg-orange-400" },
];

export function GraphLegend() {
  const [open, setOpen] = useState(false);

  return (
    <div className="absolute bottom-3 left-3 z-10">
      {open ? (
        <div className="bg-white/95 backdrop-blur-sm border border-gray-200 rounded-lg shadow-sm p-3 text-xs space-y-3 w-44">
          <div className="flex items-center justify-between">
            <span className="font-medium text-gray-700">Legend</span>
            <button
              onClick={() => setOpen(false)}
              className="text-gray-400 hover:text-gray-600 text-sm leading-none"
              aria-label="Close legend"
            >
              &times;
            </button>
          </div>

          <div className="space-y-1.5">
            <span className="text-[10px] font-medium text-gray-500 uppercase tracking-wide">Status</span>
            {STATUS_ITEMS.map((s) => (
              <div key={s.label} className="flex items-center gap-2">
                <div className={`w-4 h-3 rounded-sm border ${s.bg} ${s.border}`} />
                <span className="text-gray-600">{s.label}</span>
              </div>
            ))}
          </div>

          <div className="space-y-1.5">
            <span className="text-[10px] font-medium text-gray-500 uppercase tracking-wide">Priority</span>
            {PRIORITY_ITEMS.map((p) => (
              <div key={p.label} className="flex items-center gap-2">
                <div className="w-4 h-3 rounded-sm border border-gray-200 relative overflow-hidden">
                  <div className={`absolute left-0 top-0 bottom-0 w-1 ${p.color}`} />
                </div>
                <span className="text-gray-600">{p.label}</span>
              </div>
            ))}
          </div>

          <div className="space-y-1.5">
            <span className="text-[10px] font-medium text-gray-500 uppercase tracking-wide">Edges</span>
            <div className="flex items-center gap-2">
              <svg width="16" height="12" viewBox="0 0 16 12" className="text-gray-400">
                <line x1="0" y1="6" x2="10" y2="6" stroke="currentColor" strokeWidth="1.5" />
                <polygon points="10,3 16,6 10,9" fill="currentColor" />
              </svg>
              <span className="text-gray-600">Depends on</span>
            </div>
          </div>
        </div>
      ) : (
        <button
          onClick={() => setOpen(true)}
          className="bg-white/95 backdrop-blur-sm border border-gray-200 rounded-lg shadow-sm px-2.5 py-1.5 text-xs text-gray-500 hover:text-gray-700 hover:border-gray-300 transition-colors"
        >
          Legend
        </button>
      )}
    </div>
  );
}
