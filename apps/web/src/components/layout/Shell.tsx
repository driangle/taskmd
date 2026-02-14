import type { ReactNode } from "react";
import { Link, NavLink } from "react-router-dom";
import { useConfig } from "../../hooks/use-config.ts";

const tabs = [
  { path: "/tasks", label: "Tasks" },
  { path: "/board", label: "Board" },
  { path: "/graph", label: "Graph" },
  { path: "/stats", label: "Stats" },
];

interface ShellProps {
  children: ReactNode;
}

export function Shell({ children }: ShellProps) {
  const { readonly } = useConfig();

  return (
    <div className="min-h-screen bg-gray-50 text-gray-900">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6">
          <div className="flex items-center justify-between h-14">
            <div className="flex items-center gap-2">
              <Link
                to="/tasks"
                className="text-lg font-semibold tracking-tight"
              >
                taskmd
              </Link>
              {readonly && (
                <span className="px-2 py-0.5 text-xs font-medium rounded-full bg-amber-100 text-amber-800">
                  Read Only
                </span>
              )}
            </div>
            <nav className="flex gap-1">
              {tabs.map((tab) => (
                <NavLink
                  key={tab.path}
                  to={tab.path}
                  className={({ isActive }) =>
                    `px-3 py-1.5 text-sm rounded-md transition-colors ${
                      isActive
                        ? "bg-gray-900 text-white"
                        : "text-gray-600 hover:text-gray-900 hover:bg-gray-100"
                    }`
                  }
                >
                  {tab.label}
                </NavLink>
              ))}
            </nav>
          </div>
        </div>
      </header>
      <main className="max-w-7xl mx-auto px-4 sm:px-6 py-6">{children}</main>
    </div>
  );
}
