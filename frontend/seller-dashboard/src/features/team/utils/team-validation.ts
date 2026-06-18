import { z } from "zod";

import { ASSIGNABLE_SELLER_STAFF_ROLES } from "./seller-permissions";

export const inviteStaffSchema = z.object({
  email: z
    .string()
    .trim()
    .email("Enter a valid email address.")
    .transform((value) => value.toLowerCase()),
  role: z.enum(ASSIGNABLE_SELLER_STAFF_ROLES),
});

export type InviteStaffFormInput = z.input<typeof inviteStaffSchema>;
export type InviteStaffFormValues = z.output<typeof inviteStaffSchema>;

export const emptyInviteStaffFormValues: InviteStaffFormInput = {
  email: "",
  role: "seller_catalog_editor",
};
