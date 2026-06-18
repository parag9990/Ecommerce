const cancellableStatuses = new Set([
  'created',
  'pending_payment',
  'paid',
  'packed',
]);

export function canCancelOrder(status?: string) {
  return status ? cancellableStatuses.has(status) : false;
}

export function formatOrderStatus(status?: string) {
  return status ? status.replaceAll('_', ' ') : 'unknown';
}
