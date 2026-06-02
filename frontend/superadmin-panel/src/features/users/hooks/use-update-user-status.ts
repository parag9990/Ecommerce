import { useMutation, useQueryClient } from "@tanstack/react-query";

import { updateUserStatus } from "../api/users-api";

export function useUpdateUserStatus() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: updateUserStatus,
    onSuccess: (_data, input) => {
      queryClient.invalidateQueries({ queryKey: ["admin-users"] });
      queryClient.invalidateQueries({ queryKey: ["admin-user", input.userId] });
    }
  });
}
