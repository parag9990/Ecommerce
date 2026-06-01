import { zodResolver } from '@hookform/resolvers/zod';
import { useEffect } from 'react';
import { useForm } from 'react-hook-form';

import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import { addressSchema, type AddressFormValues } from '../address-schema';
import type { Address } from '../types';

type AddressFormProps = {
  address?: Address | undefined;
  isSaving?: boolean | undefined;
  onCancel: () => void;
  onSubmit: (values: AddressFormValues) => Promise<void>;
};

const blankAddressValues: AddressFormValues = {
  city: '',
  country: 'India',
  is_default: false,
  line1: '',
  line2: '',
  name: '',
  phone: '',
  postal_code: '',
  state: '',
};

function getDefaultValues(address?: Address): AddressFormValues {
  if (!address) {
    return blankAddressValues;
  }

  return {
    city: address.city,
    country: address.country,
    is_default: address.is_default ?? false,
    line1: address.line1,
    line2: address.line2 ?? '',
    name: address.name,
    phone: address.phone ?? '',
    postal_code: address.postal_code,
    state: address.state,
  };
}

export function AddressForm({
  address,
  isSaving = false,
  onCancel,
  onSubmit,
}: AddressFormProps) {
  const {
    formState: { errors },
    handleSubmit,
    register,
    reset,
  } = useForm<AddressFormValues>({
    defaultValues: getDefaultValues(address),
    resolver: zodResolver(addressSchema),
  });

  useEffect(() => {
    reset(getDefaultValues(address));
  }, [address, reset]);

  return (
    <form
      className="space-y-4 rounded-md border border-slate-200 bg-white p-4"
      noValidate
      onSubmit={(event) => {
        void handleSubmit(onSubmit)(event);
      }}
    >
      <div>
        <h2 className="text-lg font-semibold text-slate-950">
          {address ? 'Edit address' : 'Add address'}
        </h2>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <label className="grid gap-1.5">
          <span className="text-sm font-medium text-slate-700">Name</span>
          <Input
            autoComplete="name"
            invalid={Boolean(errors.name)}
            {...register('name')}
          />
          {errors.name ? (
            <span className="text-sm text-red-700">{errors.name.message}</span>
          ) : null}
        </label>

        <label className="grid gap-1.5">
          <span className="text-sm font-medium text-slate-700">Phone</span>
          <Input
            autoComplete="tel"
            invalid={Boolean(errors.phone)}
            {...register('phone')}
          />
          {errors.phone ? (
            <span className="text-sm text-red-700">{errors.phone.message}</span>
          ) : null}
        </label>
      </div>

      <label className="grid gap-1.5">
        <span className="text-sm font-medium text-slate-700">Address line 1</span>
        <Input
          autoComplete="address-line1"
          invalid={Boolean(errors.line1)}
          {...register('line1')}
        />
        {errors.line1 ? (
          <span className="text-sm text-red-700">{errors.line1.message}</span>
        ) : null}
      </label>

      <label className="grid gap-1.5">
        <span className="text-sm font-medium text-slate-700">Address line 2</span>
        <Input autoComplete="address-line2" {...register('line2')} />
      </label>

      <div className="grid gap-4 sm:grid-cols-2">
        <label className="grid gap-1.5">
          <span className="text-sm font-medium text-slate-700">City</span>
          <Input
            autoComplete="address-level2"
            invalid={Boolean(errors.city)}
            {...register('city')}
          />
          {errors.city ? (
            <span className="text-sm text-red-700">{errors.city.message}</span>
          ) : null}
        </label>

        <label className="grid gap-1.5">
          <span className="text-sm font-medium text-slate-700">State</span>
          <Input
            autoComplete="address-level1"
            invalid={Boolean(errors.state)}
            {...register('state')}
          />
          {errors.state ? (
            <span className="text-sm text-red-700">{errors.state.message}</span>
          ) : null}
        </label>

        <label className="grid gap-1.5">
          <span className="text-sm font-medium text-slate-700">Postal code</span>
          <Input
            autoComplete="postal-code"
            invalid={Boolean(errors.postal_code)}
            {...register('postal_code')}
          />
          {errors.postal_code ? (
            <span className="text-sm text-red-700">
              {errors.postal_code.message}
            </span>
          ) : null}
        </label>

        <label className="grid gap-1.5">
          <span className="text-sm font-medium text-slate-700">Country</span>
          <Input
            autoComplete="country-name"
            invalid={Boolean(errors.country)}
            {...register('country')}
          />
          {errors.country ? (
            <span className="text-sm text-red-700">
              {errors.country.message}
            </span>
          ) : null}
        </label>
      </div>

      <label className="flex items-center gap-2 text-sm font-medium text-slate-700">
        <input
          className="h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-600"
          type="checkbox"
          {...register('is_default')}
        />
        Use as default address
      </label>

      <div className="flex flex-wrap gap-3">
        <Button disabled={isSaving} type="submit">
          {isSaving ? 'Saving' : address ? 'Save address' : 'Add address'}
        </Button>
        <Button
          disabled={isSaving}
          onClick={onCancel}
          type="button"
          variant="secondary"
        >
          Cancel
        </Button>
      </div>
    </form>
  );
}
