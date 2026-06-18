import { useCallback, useState } from 'react';

import { useCartMutations } from './use-cart-mutations';
import { useCartQuery } from './use-cart-query';

function getErrorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}

export function useCart() {
  const cartQuery = useCartQuery();
  const { removeItem: removeItemMutation, updateItem } = useCartMutations();
  const [actionError, setActionError] = useState<string>();
  const [mutatingItemId, setMutatingItemId] = useState<string>();

  const changeQuantity = useCallback(
    async (itemId: string, quantity: number) => {
      setActionError(undefined);
      setMutatingItemId(itemId);

      try {
        await updateItem.mutateAsync({ itemId, quantity });
      } catch (error) {
        setActionError(getErrorMessage(error, 'Quantity could not be updated.'));
      } finally {
        setMutatingItemId(undefined);
      }
    },
    [updateItem],
  );

  const removeItem = useCallback(async (itemId: string) => {
    setActionError(undefined);
    setMutatingItemId(itemId);

    try {
      await removeItemMutation.mutateAsync(itemId);
    } catch (error) {
      setActionError(getErrorMessage(error, 'Item could not be removed.'));
    } finally {
      setMutatingItemId(undefined);
    }
  }, [removeItemMutation]);

  return {
    actionError,
    cart: cartQuery.data,
    changeQuantity,
    error:
      cartQuery.error instanceof Error
        ? cartQuery.error.message
        : cartQuery.isError
          ? 'Cart could not be loaded.'
          : undefined,
    isLoading: cartQuery.isLoading,
    mutatingItemId,
    reload: cartQuery.refetch,
    removeItem,
  };
}
