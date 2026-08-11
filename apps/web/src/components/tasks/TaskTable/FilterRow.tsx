import type { CSSProperties } from "react";

export const INACTIVE_STYLE = "bg-gray-50 border border-gray-200 text-gray-400 hover:bg-gray-100 hover:text-gray-500 dark:bg-gray-800/50 dark:border-gray-700 dark:text-gray-500 dark:hover:bg-gray-700 dark:hover:text-gray-400";

const PILL_BASE = "min-h-[44px] sm:min-h-0 inline-flex items-center px-2.5 py-1 text-xs rounded-full transition-colors duration-150";

export interface FilterRowProps {
  label: string;
  items: readonly string[];
  selected: Set<string>;
  colors: Record<string, string>;
  /** Inline styles for values Tailwind cannot class ahead of time (custom effort vocabularies). */
  styles?: Record<string, CSSProperties>;
  onToggle: (item: string) => void;
  onSelectAll: () => void;
}

/** A labelled row of toggle pills with an "all" reset. */
export function FilterRow({ label, items, selected, colors, styles, onToggle, onSelectAll }: FilterRowProps) {
  const allSelected = selected.size === items.length;
  return (
    <div className="flex items-center gap-2 flex-wrap" data-arrow-nav>
      <span className="text-xs text-gray-500 dark:text-gray-400 font-medium">{label}:</span>
      <button
        onClick={onSelectAll}
        className={`${PILL_BASE} ${
          allSelected
            ? "bg-gray-200 text-gray-700 font-medium ring-1 ring-gray-300 dark:bg-gray-600 dark:text-gray-200 dark:ring-gray-500"
            : INACTIVE_STYLE
        }`}
      >
        all
      </button>
      {items.map((item) => {
        const active = selected.has(item);
        return (
          <button
            key={item}
            onClick={() => onToggle(item)}
            style={active ? styles?.[item] : undefined}
            className={`${PILL_BASE} ${active ? (colors[item] ?? "font-medium") : INACTIVE_STYLE}`}
          >
            {item}
          </button>
        );
      })}
    </div>
  );
}
