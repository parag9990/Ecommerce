import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  createDeletionRequest,
  listDeletionRequests,
  previewDeletion,
  type CreateDeletionRequest,
  type DeletionPreviewRequest
} from "../../../api/session-api";

export function useDeletionRequests() {
  return useQuery({
    queryFn: listDeletionRequests,
    queryKey: ["analytics", "privacy", "deletion-requests"],
    staleTime: 30_000
  });
}

export function usePreviewDeletion() {
  return useMutation({
    mutationFn: (request: DeletionPreviewRequest) => previewDeletion(request)
  });
}

export function useCreateDeletionRequest() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateDeletionRequest) =>
      createDeletionRequest(request),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["analytics", "privacy", "deletion-requests"]
      });
    }
  });
}
