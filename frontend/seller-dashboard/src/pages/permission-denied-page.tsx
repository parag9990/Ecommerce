import { useLocation } from "react-router-dom";

import type { SellerStatus } from "../api/seller-session-api";
import { PermissionDeniedState } from "../components/state/permission-denied-state";

type PermissionLocationState = {
  sellerStatus?: SellerStatus;
  sellerName?: string;
};

function getCopy(status?: SellerStatus) {
  if (status === "pending") {
    return {
      title: "Seller approval pending",
      description:
        "Aapka seller profile review me hai. Approval complete hote hi dashboard access enable ho jayega.",
    };
  }

  if (status === "suspended") {
    return {
      title: "Seller account suspended",
      description:
        "Is seller account ka dashboard access temporarily disabled hai. Support team se contact karein.",
    };
  }

  return {
    title: "Seller access unavailable",
    description:
      "Your seller profile is not active yet. Check approval, KYC, or suspension status with support before opening the dashboard.",
  };
}

export function PermissionDeniedPage() {
  const location = useLocation();
  const state = location.state as PermissionLocationState | null;
  const copy = getCopy(state?.sellerStatus);

  return (
    <main className="flex min-h-screen items-center justify-center bg-slate-100 p-6">
      <div className="w-full max-w-lg">
        <PermissionDeniedState
          title={copy.title}
          description={
            state?.sellerName
              ? `${copy.description} Active seller: ${state.sellerName}.`
              : copy.description
          }
        />
      </div>
    </main>
  );
}
