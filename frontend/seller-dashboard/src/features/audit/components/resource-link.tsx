import { Link } from "react-router-dom";

import type { SellerAuditLog } from "../types";

function getResourcePath(log: SellerAuditLog) {
  switch (log.resource_type) {
    case "product":
      return `/seller/products/${log.resource_id}/edit`;
    case "order":
      return `/seller/orders/${log.resource_id}`;
    case "coupon":
      return `/seller/offers/coupons/${log.resource_id}/edit`;
    case "campaign":
      return "/seller/offers";
    case "team":
      return "/seller/team";
    default:
      return null;
  }
}

type ResourceLinkProps = {
  log: SellerAuditLog;
};

export function ResourceLink({ log }: ResourceLinkProps) {
  const label = log.resource_title || log.resource_id;
  const path = getResourcePath(log);

  if (!path) {
    return <span className="font-mono text-xs text-slate-700">{label}</span>;
  }

  return (
    <Link
      className="font-medium text-blue-700 transition hover:text-blue-900"
      to={path}
    >
      {label}
    </Link>
  );
}
