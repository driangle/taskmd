import type { ReactNode } from "react";

const tabs = [
  { id: "tasks", label: "Tasks" },
  { id: "board", label: "Board" },
  { id: "graph", label: "Graph" },
  { id: "stats", label: "Stats" },
] as const;

export type TabId = (typeof tabs)[number]["id"];

interface ShellProps {
  activeTab: TabId;
  onTabChange: (tab: TabId) => void;
  children: ReactNode;
}

export function Shell({ activeTab, onTabChange, children }: ShellProps) {
  return (
    <div className="min-h-screen bg-gray-50 text-gray-900">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6">
          <div className="flex items-center justify-between h-14">
            <h1 className="text-lg font-semibold tracking-tight">
              taskmd
            </h1>
            <nav className="flex gap-1">
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => onTabChange(tab.id)}
                  className={`px-3 py-1.5 text-sm rounded-md transition-colors ${
                    activeTab === tab.id
                      ? "bg-gray-900 text-white"
                      : "text-gray-600 hover:text-gray-900 hover:bg-gray-100"
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </nav>
          </div>
        </div>
      </header>
      <main className="max-w-7xl mx-auto px-4 sm:px-6 py-6">{children}</main>
    </div>
  );
}
