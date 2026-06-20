import { useMutation } from "@tanstack/react-query";

import { exportAdminAuditLogs } from "../api/admin-audit-api";

export function useExportAuditLogs() {
  return useMutation({
    mutationFn: exportAdminAuditLogs
  });
}
