import { useMemo, useState } from "react";
import { CheckSquare } from "lucide-react";

import type { AdminSeller } from "../types";

const CHECKLIST_ITEMS = [
  "Store name and support email look valid",
  "GST or business document matches seller profile",
  "KYC documents are readable and not expired",
  "Catalog has no prohibited or counterfeit products",
  "Seller policies follow marketplace rules"
] as const;

export function SellerReviewChecklist({ seller }: { seller: AdminSeller }) {
  const [checkedItems, setCheckedItems] = useState<Set<string>>(() => new Set());
  const completedCount = checkedItems.size;
  const progressLabel = useMemo(
    () => `${completedCount} of ${CHECKLIST_ITEMS.length} checks complete`,
    [completedCount]
  );

  return (
    <aside className="rounded-lg border border-slate-200 bg-white p-4">
      <div className="flex items-center gap-2">
        <CheckSquare className="h-4 w-4 text-slate-500" aria-hidden="true" />
        <h2 className="text-base font-semibold text-slate-950">Review Checklist</h2>
      </div>
      <p className="mt-1 text-sm text-slate-600">{progressLabel}</p>

      <div className="mt-4 space-y-3">
        {CHECKLIST_ITEMS.map((item) => (
          <label key={item} className="flex items-start gap-3 text-sm text-slate-700">
            <input
              type="checkbox"
              checked={checkedItems.has(item)}
              onChange={(event) => {
                setCheckedItems((current) => {
                  const next = new Set(current);

                  if (event.target.checked) {
                    next.add(item);
                  } else {
                    next.delete(item);
                  }

                  return next;
                });
              }}
              className="mt-0.5 h-4 w-4 rounded border-slate-300"
            />
            <span>{item}</span>
          </label>
        ))}
      </div>

      <p className="mt-4 break-all rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-xs text-slate-500">
        Local checklist for {seller.seller_id}. Mutation audit is still enforced by the reason dialog.
      </p>
    </aside>
  );
}
