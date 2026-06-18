import { z } from "zod";

const optionalNumber = (schema: z.ZodNumber) =>
  z.preprocess((value) => {
    if (value === "" || value === null || value === undefined) {
      return undefined;
    }

    return Number(value);
  }, schema.optional());

const optionalDateString = z.preprocess((value) => {
  if (typeof value !== "string" || value.trim() === "") {
    return undefined;
  }

  return value;
}, z.string().optional());

const currencySchema = z
  .string()
  .trim()
  .length(3, "Currency must be a 3-letter ISO code.")
  .default("INR");

const optionalMoneyFormSchema = z.object({
  amount: optionalNumber(z.number().int("Amount must be a whole number.").min(0, "Amount cannot be negative.")),
  currency: currencySchema,
});

export const couponFormSchema = z
  .object({
    code: z
      .string()
      .trim()
      .min(3, "Enter a coupon code with at least 3 characters.")
      .max(64, "Coupon code cannot exceed 64 characters.")
      .regex(/^[A-Z0-9_-]+$/, "Use only A-Z, 0-9, underscore, or dash."),
    discount_type: z.enum(["fixed", "percentage"]),
    discount_value: z.coerce
      .number()
      .int("Discount value must be a whole number.")
      .positive("Enter a discount value greater than 0."),
    min_cart_amount: optionalMoneyFormSchema,
    starts_at: optionalDateString,
    ends_at: optionalDateString,
    usage_limit: optionalNumber(
      z.number().int("Usage limit must be a whole number.").positive("Usage limit must be greater than 0."),
    ),
  })
  .superRefine((value, ctx) => {
    if (value.discount_type === "percentage" && value.discount_value > 100) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["discount_value"],
        message: "Percentage discount cannot exceed 100.",
      });
    }

    if (value.starts_at && value.ends_at) {
      const startsAt = new Date(value.starts_at);
      const endsAt = new Date(value.ends_at);

      if (Number.isNaN(startsAt.getTime()) || Number.isNaN(endsAt.getTime())) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ["starts_at"],
          message: "Enter valid start and end dates.",
        });
        return;
      }

      if (endsAt <= startsAt) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ["ends_at"],
          message: "End time must be after start time.",
        });
      }
    }
  });

export const campaignFormSchema = z
  .object({
    name: z
      .string()
      .trim()
      .min(3, "Enter a campaign name with at least 3 characters.")
      .max(120, "Campaign name cannot exceed 120 characters."),
    starts_at: z.string().min(1, "Select a campaign start time."),
    ends_at: z.string().min(1, "Select a campaign end time."),
    budget: optionalMoneyFormSchema,
  })
  .superRefine((value, ctx) => {
    const startsAt = new Date(value.starts_at);
    const endsAt = new Date(value.ends_at);

    if (Number.isNaN(startsAt.getTime()) || Number.isNaN(endsAt.getTime())) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["starts_at"],
        message: "Enter valid campaign dates.",
      });
      return;
    }

    if (endsAt <= startsAt) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ["ends_at"],
        message: "Campaign end time must be after start time.",
      });
    }
  });

export type CouponFormInput = z.input<typeof couponFormSchema>;
export type CouponFormValues = z.output<typeof couponFormSchema>;
export type CampaignFormInput = z.input<typeof campaignFormSchema>;
export type CampaignFormValues = z.output<typeof campaignFormSchema>;

export const emptyCouponFormValues: CouponFormInput = {
  code: "",
  discount_type: "percentage",
  discount_value: 10,
  min_cart_amount: {
    amount: undefined,
    currency: "INR",
  },
  starts_at: undefined,
  ends_at: undefined,
  usage_limit: undefined,
};

export const emptyCampaignFormValues: CampaignFormInput = {
  name: "",
  starts_at: "",
  ends_at: "",
  budget: {
    amount: undefined,
    currency: "INR",
  },
};
