import { useState } from "react";
import { Shell, type TabId } from "./components/layout/Shell.tsx";
import { TasksPage } from "./pages/TasksPage.tsx";
import { BoardPage } from "./pages/BoardPage.tsx";
import { GraphPage } from "./pages/GraphPage.tsx";
import { StatsPage } from "./pages/StatsPage.tsx";
import { useLiveReload } from "./hooks/use-live-reload.ts";

export default function App() {
  const [activeTab, setActiveTab] = useState<TabId>("tasks");

  useLiveReload();

  return (
    <Shell activeTab={activeTab} onTabChange={setActiveTab}>
      {activeTab === "tasks" && <TasksPage />}
      {activeTab === "board" && <BoardPage />}
      {activeTab === "graph" && <GraphPage />}
      {activeTab === "stats" && <StatsPage />}
    </Shell>
  );
}
