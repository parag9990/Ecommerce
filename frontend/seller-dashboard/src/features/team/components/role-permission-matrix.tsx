import { Check, X } from "lucide-react";

import {
  PERMISSION_LABELS,
  ROLE_LABELS,
  ROLE_MATRIX_PERMISSIONS,
  ROLE_PERMISSIONS,
  SELLER_STAFF_ROLES,
} from "../utils/seller-permissions";

export function RolePermissionMatrix() {
  return (
    <div className="overflow-hidden rounded-md border border-slate-200 bg-white shadow-sm">
      <div className="overflow-x-auto">
        <table className="w-full min-w-[760px] border-collapse text-left text-sm">
          <thead className="bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
            <tr>
              <th className="px-4 py-3 font-semibold">Permission</th>
              {SELLER_STAFF_ROLES.map((role) => (
                <th key={role} className="px-4 py-3 font-semibold">
                  {ROLE_LABELS[role]}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {ROLE_MATRIX_PERMISSIONS.map((permission) => (
              <tr key={permission}>
                <td className="px-4 py-3 font-medium text-slate-700">
                  {PERMISSION_LABELS[permission]}
                </td>
                {SELLER_STAFF_ROLES.map((role) => {
                  const allowed = ROLE_PERMISSIONS[role].includes(permission);

                  return (
                    <td key={`${role}-${permission}`} className="px-4 py-3">
                      {allowed ? (
                        <Check
                          aria-label="Allowed"
                          className="h-4 w-4 text-emerald-600"
                        />
                      ) : (
                        <X aria-label="Not allowed" className="h-4 w-4 text-slate-300" />
                      )}
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
