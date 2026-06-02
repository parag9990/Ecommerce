import type { ReactNode } from "react";

import { PermissionDeniedState } from "../state/permission-denied-state";
import { useSellerPermissions } from "../../features/team/hooks/use-seller-permissions";
import type { SellerPermission } from "../../features/team/types";
import { PERMISSION_LABELS } from "../../features/team/utils/seller-permissions";

type PermissionGateProps = {
  permission: SellerPermission;
  children: ReactNode;
  fallback?: ReactNode;
};

export function PermissionGate({
  permission,
  children,
  fallback,
}: PermissionGateProps) {
  const permissions = useSellerPermissions();

  if (!permissions.can(permission)) {
    return (
      <>
        {fallback ?? (
          <PermissionDeniedState
            title="Permission required"
            description={`Aapke current seller role ke paas ${PERMISSION_LABELS[
              permission
            ].toLowerCase()} permission nahi hai.`}
          />
        )}
      </>
    );
  }

  return <>{children}</>;
}
