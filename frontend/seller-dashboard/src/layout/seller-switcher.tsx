import { useQuery } from "@tanstack/react-query";

import { getSellerSession } from "../api/seller-session-api";
import { StatusBadge } from "../components/ui/status-badge";
import { useSellerStore } from "../stores/seller-store";

export function SellerSwitcher() {
  const activeSeller = useSellerStore((state) => state.activeSeller);
  const setActiveSeller = useSellerStore((state) => state.setActiveSeller);

  const { data } = useQuery({
    queryKey: ["seller-session"],
    queryFn: getSellerSession,
  });

  const sellers = data?.sellers ?? [];
  const sellerOptions =
    activeSeller && !sellers.some((seller) => seller.seller_id === activeSeller.seller_id)
      ? [activeSeller, ...sellers]
      : sellers;

  if (!activeSeller) {
    return (
      <span className="shrink-0 rounded-md border border-amber-200 bg-amber-50 px-2 py-1 text-xs font-medium text-amber-700">
        No seller selected
      </span>
    );
  }

  return (
    <div className="flex min-w-0 items-center gap-2">
      <label
        htmlFor="seller-switcher"
        className="hidden text-xs font-medium text-slate-500 sm:inline"
      >
        Seller
      </label>
      <select
        id="seller-switcher"
        value={activeSeller.seller_id}
        onChange={(event) => {
          const nextSeller = sellerOptions.find(
            (seller) => seller.seller_id === event.target.value,
          );

          if (nextSeller?.status === "active") {
            setActiveSeller(nextSeller);
          }
        }}
        className="h-8 max-w-[44vw] rounded-md border border-slate-200 bg-white px-2 text-sm text-slate-900 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100 sm:max-w-56"
      >
        {sellerOptions.map((seller) => (
          <option
            key={seller.seller_id}
            value={seller.seller_id}
            disabled={seller.status !== "active"}
          >
            {seller.display_name}
            {seller.status !== "active" ? ` (${seller.status})` : ""}
          </option>
        ))}
      </select>
      <StatusBadge status={activeSeller.status} className="hidden lg:inline-flex" />
    </div>
  );
}
