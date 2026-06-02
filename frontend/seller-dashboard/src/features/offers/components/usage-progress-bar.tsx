type UsageProgressBarProps = {
  used?: number;
  limit?: number;
};

export function UsageProgressBar({ used, limit }: UsageProgressBarProps) {
  if (!limit) {
    return <span className="text-xs text-slate-500">No limit</span>;
  }

  if (used === undefined) {
    return <span className="text-xs text-slate-500">Limit {limit}</span>;
  }

  const percentage = Math.min(100, Math.round((used / limit) * 100));

  return (
    <div className="w-36">
      <div className="mb-1 flex items-center justify-between text-xs text-slate-500">
        <span>
          {used}/{limit}
        </span>
        <span>{percentage}%</span>
      </div>
      <div className="h-1.5 rounded-full bg-slate-100">
        <div
          className="h-1.5 rounded-full bg-emerald-500"
          style={{ width: `${percentage}%` }}
        />
      </div>
    </div>
  );
}
