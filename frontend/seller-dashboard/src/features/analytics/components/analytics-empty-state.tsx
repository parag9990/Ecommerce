import { EmptyState } from "../../../components/state/empty-state";
import { UnavailableState } from "../../../components/state/unavailable-state";

type AnalyticsEmptyStateProps = {
  title: string;
  description: string;
  kind?: "empty" | "unavailable";
  className?: string;
};

export function AnalyticsEmptyState({
  title,
  description,
  kind = "unavailable",
  className,
}: AnalyticsEmptyStateProps) {
  if (kind === "empty") {
    return <EmptyState title={title} description={description} className={className} />;
  }

  return (
    <UnavailableState
      title={title}
      description={description}
      className={className}
    />
  );
}
