import { useQuery } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { useEffect } from "react";
import { Navigate, useLocation } from "react-router-dom";

import { getSellerSession } from "../api/seller-session-api";
import { StateShell } from "../components/state/state-shell";
import { useSellerStore } from "../stores/seller-store";

type RequireSellerProps = {
  children: ReactNode;
};

export function RequireSeller({ children }: RequireSellerProps) {
  const location = useLocation();
  const setActiveSeller = useSellerStore((state) => state.setActiveSeller);

  const sessionQuery = useQuery({
    queryKey: ["seller-session"],
    queryFn: getSellerSession,
  });

  const session = sessionQuery.data;

  useEffect(() => {
    if (session?.active_seller?.status === "active") {
      setActiveSeller(session.active_seller);
      return;
    }

    setActiveSeller(null);
  }, [session?.active_seller, setActiveSeller]);

  if (sessionQuery.isPending) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-slate-100 p-6">
        <div className="w-full max-w-lg">
          <StateShell
            kind="loading"
            title="Dashboard loading"
            description="Seller session verify ho raha hai. Dashboard shortly available hoga."
          />
        </div>
      </main>
    );
  }

  if (sessionQuery.isError || !session?.authenticated) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  if (!session.active_seller || session.active_seller.status !== "active") {
    return (
      <Navigate
        to="/seller/permission-denied"
        replace
        state={{
          sellerStatus: session.active_seller?.status,
          sellerName: session.active_seller?.display_name,
        }}
      />
    );
  }

  return <>{children}</>;
}
