type Variant = "table" | "board" | "graph" | "cards" | "detail" | "default";

interface LoadingStateProps {
  variant?: Variant;
}

export function LoadingState({ variant = "default" }: LoadingStateProps) {
  switch (variant) {
    case "table":
      return <TableSkeleton />;
    case "board":
      return <BoardSkeleton />;
    case "graph":
      return <GraphSkeleton />;
    case "cards":
      return <CardsSkeleton />;
    case "detail":
      return <DetailSkeleton />;
    default:
      return <DefaultSkeleton />;
  }
}

function TableSkeleton() {
  return (
    <div className="animate-pulse">
      <div className="h-8 bg-gray-100 rounded mb-3" />
      {Array.from({ length: 6 }, (_, i) => (
        <div key={i} className="flex gap-4 mb-2">
          <div className="h-6 bg-gray-100 rounded w-16" />
          <div className="h-6 bg-gray-100 rounded flex-1" />
          <div className="h-6 bg-gray-100 rounded w-20" />
          <div className="h-6 bg-gray-100 rounded w-16" />
        </div>
      ))}
    </div>
  );
}

function BoardSkeleton() {
  return (
    <div className="animate-pulse flex gap-4">
      {Array.from({ length: 4 }, (_, col) => (
        <div key={col} className="flex-1 min-w-[200px]">
          <div className="h-6 bg-gray-100 rounded mb-3 w-24" />
          {Array.from({ length: 3 - col % 2 }, (_, row) => (
            <div
              key={row}
              className="h-20 bg-gray-100 rounded-lg mb-2"
            />
          ))}
        </div>
      ))}
    </div>
  );
}

function GraphSkeleton() {
  return (
    <div className="animate-pulse bg-white rounded-lg border border-gray-200 h-[calc(100vh-200px)] min-h-[500px] flex items-center justify-center">
      <div className="text-center">
        <div className="h-8 w-8 mx-auto mb-3 rounded-full bg-gray-100" />
        <div className="h-4 bg-gray-100 rounded w-32 mx-auto" />
      </div>
    </div>
  );
}

function CardsSkeleton() {
  return (
    <div className="animate-pulse">
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-6">
        {Array.from({ length: 4 }, (_, i) => (
          <div key={i} className="bg-gray-100 rounded-lg h-20" />
        ))}
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div className="bg-gray-100 rounded-lg h-48" />
        <div className="bg-gray-100 rounded-lg h-48" />
      </div>
    </div>
  );
}

function DetailSkeleton() {
  return (
    <div className="animate-pulse bg-white border border-gray-200 rounded-lg p-6">
      <div className="flex items-start justify-between mb-4">
        <div>
          <div className="h-3 bg-gray-100 rounded w-16 mb-2" />
          <div className="h-6 bg-gray-100 rounded w-64" />
        </div>
        <div className="h-6 bg-gray-100 rounded-full w-20" />
      </div>
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-6">
        {Array.from({ length: 3 }, (_, i) => (
          <div key={i}>
            <div className="h-3 bg-gray-100 rounded w-12 mb-1" />
            <div className="h-5 bg-gray-100 rounded w-20" />
          </div>
        ))}
      </div>
      <div className="border-t border-gray-200 pt-4 space-y-2">
        <div className="h-4 bg-gray-100 rounded w-full" />
        <div className="h-4 bg-gray-100 rounded w-3/4" />
        <div className="h-4 bg-gray-100 rounded w-5/6" />
      </div>
    </div>
  );
}

function DefaultSkeleton() {
  return (
    <div className="animate-pulse flex items-center justify-center py-12">
      <div className="text-center">
        <div className="h-8 w-8 mx-auto mb-3 rounded-full bg-gray-200" />
        <div className="h-4 bg-gray-200 rounded w-24 mx-auto" />
      </div>
    </div>
  );
}
