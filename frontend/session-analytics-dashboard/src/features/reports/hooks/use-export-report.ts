import { useMutation } from "@tanstack/react-query";

import {
  exportAnalyticsReport,
  type ExportReportRequest
} from "../../../api/session-api";
import { buildCsvFilename, sanitizeCsvFilename } from "../lib/report-filenames";
import { saveCsvBlob } from "../lib/save-csv-blob";

export function useExportReport() {
  return useMutation({
    mutationFn: (request: ExportReportRequest) => exportAnalyticsReport(request),
    onSuccess: (result, variables: ExportReportRequest) => {
      saveCsvBlob(
        result.blob,
        sanitizeCsvFilename(
          result.filename ??
            buildCsvFilename({
              from: variables.from,
              reportType: variables.reportType,
              to: variables.to
            })
        )
      );
    }
  });
}
