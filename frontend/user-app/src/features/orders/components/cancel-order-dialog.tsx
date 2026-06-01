import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';

import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import {
  cancelOrderSchema,
  type CancelOrderFormValues,
} from '../order-schema';

type CancelOrderDialogProps = {
  isSubmitting?: boolean | undefined;
  onClose: () => void;
  onSubmit: (reason: string) => Promise<void>;
};

export function CancelOrderDialog({
  isSubmitting = false,
  onClose,
  onSubmit,
}: CancelOrderDialogProps) {
  const {
    formState: { errors },
    handleSubmit,
    register,
  } = useForm<CancelOrderFormValues>({
    defaultValues: { reason: '' },
    resolver: zodResolver(cancelOrderSchema),
  });

  async function submit(values: CancelOrderFormValues) {
    await onSubmit(values.reason.trim());
  }

  return (
    <div
      aria-modal="true"
      className="fixed inset-0 z-40 grid place-items-center bg-slate-950/30 px-4"
      role="dialog"
    >
      <form
        className="w-full max-w-md rounded-md bg-white p-5 shadow-xl"
        noValidate
        onSubmit={(event) => {
          void handleSubmit(submit)(event);
        }}
      >
        <h2 className="text-lg font-semibold text-slate-950">Cancel order?</h2>
        <p className="mt-2 text-sm leading-6 text-slate-600">
          The backend will confirm whether this order can still be cancelled.
        </p>

        <label className="mt-4 grid gap-1.5">
          <span className="text-sm font-medium text-slate-700">Reason</span>
          <Input
            disabled={isSubmitting}
            invalid={Boolean(errors.reason)}
            placeholder="Changed my mind"
            {...register('reason')}
          />
          {errors.reason ? (
            <span className="text-sm text-red-700">
              {errors.reason.message}
            </span>
          ) : null}
        </label>

        <div className="mt-5 flex flex-wrap justify-end gap-3">
          <Button
            disabled={isSubmitting}
            onClick={onClose}
            type="button"
            variant="secondary"
          >
            Keep order
          </Button>
          <Button
            className="bg-red-600 hover:bg-red-700 focus-visible:outline-red-600"
            disabled={isSubmitting}
            type="submit"
          >
            {isSubmitting ? 'Cancelling' : 'Cancel order'}
          </Button>
        </div>
      </form>
    </div>
  );
}
