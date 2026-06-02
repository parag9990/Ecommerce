import { EmptyState } from "../../../components/state/empty-state";

type AuditEmptyStateProps = {
  hasFilters: boolean;
};

export function AuditEmptyState({ hasFilters }: AuditEmptyStateProps) {
  return (
    <EmptyState
      title={hasFilters ? "No matching activity found" : "No activity recorded"}
      description={
        hasFilters
          ? "Try a broader action, resource, actor, or date range."
          : "Seller actions hote hi timeline me entries dikhenge."
      }
    />
  );
}
