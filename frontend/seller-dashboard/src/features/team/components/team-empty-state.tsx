import { EmptyState } from "../../../components/state/empty-state";

type TeamEmptyStateProps = {
  canInvite: boolean;
  onInvite: () => void;
};

export function TeamEmptyState({ canInvite, onInvite }: TeamEmptyStateProps) {
  return (
    <EmptyState
      title="No staff members found"
      description="Invite trusted staff with scoped access for catalog, orders, offers, or daily operations."
      action={{
        label: "Invite staff",
        onClick: onInvite,
        disabled: !canInvite,
        title: !canInvite ? "You do not have permission to invite staff." : undefined,
      }}
    />
  );
}
