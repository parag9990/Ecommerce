import type { SellerAuditLog } from "../types";

type ActorSummaryProps = {
  log: SellerAuditLog;
};

export function ActorSummary({ log }: ActorSummaryProps) {
  const label = log.actor_name || log.actor_email || log.actor_user_id;

  return (
    <span className="font-medium text-slate-800" title={log.actor_email}>
      {label}
    </span>
  );
}
