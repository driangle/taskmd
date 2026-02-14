import type { ReactNode } from "react";
import { Link, NavLink } from "react-router-dom";
import { useConfig } from "../../hooks/use-config.ts";
import { useTheme } from "../../hooks/use-theme.ts";

const tabs = [
  { path: "/tasks", label: "Tasks" },
  { path: "/next", label: "Next Up" },
  { path: "/board", label: "Board" },
  { path: "/graph", label: "Graph" },
  { path: "/stats", label: "Stats" },
  { path: "/validate", label: "Validate" },
];

interface ShellProps {
  children: ReactNode;
}

export function Shell({ children }: ShellProps) {
  const { readonly, version } = useConfig();
  const { theme, toggle } = useTheme();

  return (
    <div className="min-h-screen bg-gray-50 text-gray-900 dark:bg-gray-900 dark:text-gray-100">
      <header className="bg-white border-b border-gray-200 dark:bg-gray-800 dark:border-gray-700">
        <div className="max-w-7xl mx-auto px-4 sm:px-6">
          <div className="flex items-center justify-between h-14">
            <div className="flex items-center gap-2">
              <Link
                to="/tasks"
                className="text-lg font-semibold tracking-tight"
              >
                taskmd
              </Link>
              {version && (
                <span className="text-xs text-gray-400 dark:text-gray-500">
                  {version}
                </span>
              )}
              {readonly && (
                <span className="px-2 py-0.5 text-xs font-medium rounded-full bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300">
                  Read Only
                </span>
              )}
            </div>
            <nav className="flex items-center gap-1">
              {tabs.map((tab) => (
                <NavLink
                  key={tab.path}
                  to={tab.path}
                  className={({ isActive }) =>
                    `px-3 py-1.5 text-sm rounded-md transition-colors ${
                      isActive
                        ? "bg-gray-900 text-white dark:bg-white dark:text-gray-900"
                        : "text-gray-600 hover:text-gray-900 hover:bg-gray-100 dark:text-gray-400 dark:hover:text-gray-100 dark:hover:bg-gray-700"
                    }`
                  }
                >
                  {tab.label}
                </NavLink>
              ))}
              <a
                href="https://driangle.github.io/taskmd/"
                target="_blank"
                rel="noopener noreferrer"
                className="px-3 py-1.5 text-sm rounded-md transition-colors text-gray-600 hover:text-gray-900 hover:bg-gray-100 dark:text-gray-400 dark:hover:text-gray-100 dark:hover:bg-gray-700"
              >
                Docs ↗
              </a>
              <button
                onClick={toggle}
                className="ml-1 p-1.5 rounded-md text-gray-600 hover:text-gray-900 hover:bg-gray-100 dark:text-gray-400 dark:hover:text-gray-100 dark:hover:bg-gray-700 transition-colors"
                aria-label={`Switch to ${theme === "dark" ? "light" : "dark"} mode`}
              >
                {theme === "dark" ? (
                  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
                  </svg>
                ) : (
                  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
                  </svg>
                )}
              </button>
            </nav>
          </div>
        </div>
      </header>
      <main className="max-w-7xl mx-auto px-4 sm:px-6 py-6">{children}</main>
    </div>
  );
}
