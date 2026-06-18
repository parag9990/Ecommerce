import { useMutation, useQueryClient } from "@tanstack/react-query";

import {
  createProduct,
  publishProduct,
  updateProduct,
} from "../api/seller-product-api";
import type { ProductInput } from "../types";
import { productQueryKeys } from "./query-keys";

export function useCreateProduct() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: ProductInput) => createProduct(input),
    onSuccess: (product) => {
      queryClient.invalidateQueries({ queryKey: productQueryKeys.lists() });
      queryClient.setQueryData(productQueryKeys.detail(product.product_id), product);
    },
  });
}

export function useUpdateProduct(productId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: ProductInput) => updateProduct(productId, input),
    onSuccess: (product) => {
      queryClient.invalidateQueries({ queryKey: productQueryKeys.lists() });
      queryClient.setQueryData(productQueryKeys.detail(product.product_id), product);
    },
  });
}

export function usePublishProduct() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (productId: string) => publishProduct(productId),
    onSuccess: (product) => {
      queryClient.invalidateQueries({ queryKey: productQueryKeys.lists() });
      queryClient.setQueryData(productQueryKeys.detail(product.product_id), product);
    },
  });
}
