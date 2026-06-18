import type { AssignableSellerStaffRole } from "../types";
import {
  ASSIGNABLE_SELLER_STAFF_ROLES,
  ROLE_DESCRIPTIONS,
  ROLE_LABELS,
} from "../utils/seller-permissions";

type RoleSelectProps = {
  value: AssignableSellerStaffRole;
  id?: string;
  disabled?: boolean;
  labelledBy?: string;
  onChange: (role: AssignableSellerStaffRole) => void;
};

export function RoleSelect({
  value,
  id,
  disabled,
  labelledBy,
  onChange,
}: RoleSelectProps) {
  return (
    <select
      id={id}
      aria-labelledby={labelledBy}
      disabled={disabled}
      value={value}
      onChange={(event) => onChange(event.target.value as AssignableSellerStaffRole)}
      className="h-10 w-full rounded-md border border-slate-300 bg-white px-3 text-sm text-slate-900 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-500"
    >
      {ASSIGNABLE_SELLER_STAFF_ROLES.map((role) => (
        <option key={role} value={role}>
          {ROLE_LABELS[role]} - {ROLE_DESCRIPTIONS[role]}
        </option>
      ))}
    </select>
  );
}
