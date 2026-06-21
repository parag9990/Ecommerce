import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  createReportSchedule,
  deleteReportSchedule,
  getReportSchedules,
  updateReportScheduleStatus,
  type CreateReportScheduleInput,
  type UpdateReportScheduleStatusInput
} from "../../../api/session-api";

export const reportSchedulesQueryKey = ["analytics", "report-schedules"];

export function useReportSchedules() {
  return useQuery({
    queryFn: ({ signal }) => getReportSchedules({ signal }),
    queryKey: reportSchedulesQueryKey,
    staleTime: 30_000
  });
}

export function useCreateReportSchedule() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CreateReportScheduleInput) => createReportSchedule(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: reportSchedulesQueryKey });
    }
  });
}

export function useUpdateReportScheduleStatus() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: UpdateReportScheduleStatusInput) =>
      updateReportScheduleStatus(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: reportSchedulesQueryKey });
    }
  });
}

export function useDeleteReportSchedule() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => deleteReportSchedule(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: reportSchedulesQueryKey });
    }
  });
}
