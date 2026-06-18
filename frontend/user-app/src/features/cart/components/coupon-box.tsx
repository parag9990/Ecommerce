import { zodResolver } from '@hookform/resolvers/zod';
import { TicketPercent } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { Alert } from '../../../components/ui/alert';
import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { useCartMutations } from '../hooks/use-cart-mutations';
import { formatMoney } from './price';

const couponSchema = z.object({
  coupon_code: z
    .string()
    .trim()
    .min(2, 'Enter a coupon code.')
    .max(40, 'Coupon code must be 40 characters or fewer.'),
});

type CouponFormValues = z.infer<typeof couponSchema>;

type CouponBoxProps = {
  cartId?: string | undefined;
  initialCouponCode?: string | undefined;
  onCouponAccepted: (couponCode: string) => void;
  onCouponRejected?: (() => void) | undefined;
};

export function CouponBox({
  cartId,
  initialCouponCode = '',
  onCouponAccepted,
  onCouponRejected,
}: CouponBoxProps) {
  const { previewCouponCode } = useCartMutations();
  const preview = previewCouponCode.data;
  const {
    formState: { errors, isSubmitting },
    handleSubmit,
    register,
  } = useForm<CouponFormValues>({
    defaultValues: { coupon_code: initialCouponCode },
    resolver: zodResolver(couponSchema),
  });

  async function applyCoupon(values: CouponFormValues) {
    const couponCode = values.coupon_code.trim().toUpperCase();

    previewCouponCode.reset();

    try {
      const result = await previewCouponCode.mutateAsync({
        cart_id: cartId,
        coupon_code: couponCode,
      });

      if (result.valid) {
        onCouponAccepted(couponCode);
      } else {
        onCouponRejected?.();
      }
    } catch {
      onCouponRejected?.();
    }
  }

  const serverError =
    previewCouponCode.error instanceof Error
      ? previewCouponCode.error.message
      : previewCouponCode.isError
        ? 'Coupon could not be checked.'
        : undefined;

  return (
    <section className="rounded-md border border-slate-200 bg-white p-4">
      <div className="mb-3 flex items-center gap-2">
        <TicketPercent aria-hidden="true" className="h-5 w-5 text-blue-700" />
        <h2 className="text-sm font-semibold text-slate-950">Coupon</h2>
      </div>

      <form
        className="flex gap-2"
        noValidate
        onSubmit={(event) => {
          void handleSubmit(applyCoupon)(event);
        }}
      >
        <Input
          aria-describedby={errors.coupon_code ? 'coupon-code-error' : undefined}
          aria-invalid={Boolean(errors.coupon_code)}
          autoComplete="off"
          className="min-w-0 uppercase"
          invalid={Boolean(errors.coupon_code)}
          placeholder="SAVE10"
          {...register('coupon_code')}
        />
        <Button
          disabled={isSubmitting || previewCouponCode.isPending}
          type="submit"
        >
          {isSubmitting || previewCouponCode.isPending ? 'Checking' : 'Apply'}
        </Button>
      </form>

      {errors.coupon_code ? (
        <p className="mt-2 text-sm text-red-700" id="coupon-code-error">
          {errors.coupon_code.message}
        </p>
      ) : null}

      {serverError ? (
        <div className="mt-3">
          <Alert variant="error">{serverError}</Alert>
        </div>
      ) : null}

      {preview?.valid ? (
        <p className="mt-3 text-sm font-medium text-emerald-700">
          Coupon applied. Discount preview: {formatMoney(preview.discount)}
        </p>
      ) : null}

      {preview && !preview.valid ? (
        <p className="mt-3 text-sm font-medium text-amber-700">
          {preview.reason ?? 'This coupon is not applicable.'}
        </p>
      ) : null}
    </section>
  );
}
