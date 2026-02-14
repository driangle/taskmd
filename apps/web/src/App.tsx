import { Routes, Route, Navigate } from "react-router-dom";
import { Shell } from "./components/layout/Shell.tsx";
import { TasksPage } from "./pages/TasksPage.tsx";
import { TaskDetailPage } from "./pages/TaskDetailPage.tsx";
import { BoardPage } from "./pages/BoardPage.tsx";
import { GraphPage } from "./pages/GraphPage.tsx";
import { StatsPage } from "./pages/StatsPage.tsx";
import { ValidatePage } from "./pages/ValidatePage.tsx";
import { useLiveReload } from "./hooks/use-live-reload.ts";

export default function App() {
  useLiveReload();

  return (
    <Shell>
      <Routes>
        <Route path="/" element={<Navigate to="/tasks" replace />} />
        <Route path="/tasks" element={<TasksPage />} />
        <Route path="/tasks/:id" element={<TaskDetailPage />} />
        <Route path="/board" element={<BoardPage />} />
        <Route path="/graph" element={<GraphPage />} />
        <Route path="/stats" element={<StatsPage />} />
        <Route path="/validate" element={<ValidatePage />} />
      </Routes>
    </Shell>
  );
}
