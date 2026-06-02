import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  disableSellerStaff,
  inviteSellerStaff,
  listSellerTeam,
  resendSellerStaffInvite,
  updateSellerStaffRole,
} from "../api/seller-team-api";
import type {
  DisableStaffInput,
  InviteStaffInput,
  SellerTeamFilters,
  UpdateStaffRoleInput,
} from "../types";
import { teamQueryKeys } from "./query-keys";

export function useSellerTeam(sellerId: string | undefined, filters: SellerTeamFilters) {
  return useQuery({
    queryKey: teamQueryKeys.list(sellerId ?? "", filters),
    queryFn: () => listSellerTeam(filters),
    enabled: Boolean(sellerId),
    staleTime: 30_000,
  });
}

export function useInviteSellerStaff() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: InviteStaffInput) => inviteSellerStaff(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: teamQueryKeys.lists() });
    },
  });
}

export function useUpdateSellerStaffRole() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: UpdateStaffRoleInput) => updateSellerStaffRole(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: teamQueryKeys.lists() });
    },
  });
}

export function useDisableSellerStaff() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: DisableStaffInput) => disableSellerStaff(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: teamQueryKeys.lists() });
    },
  });
}

export function useResendSellerStaffInvite() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (staffId: string) => resendSellerStaffInvite(staffId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: teamQueryKeys.lists() });
    },
  });
}
