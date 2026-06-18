const CART_COUPON_KEY = 'user_app_cart_coupon_code';

function readSessionValue(key: string): string | null {
  try {
    return window.sessionStorage.getItem(key);
  } catch {
    return null;
  }
}

function writeSessionValue(key: string, value: string) {
  try {
    window.sessionStorage.setItem(key, value);
  } catch {
    // Coupon persistence is a convenience; checkout still accepts manual input.
  }
}

function removeSessionValue(key: string) {
  try {
    window.sessionStorage.removeItem(key);
  } catch {
    // Ignore storage failures.
  }
}

export function readCartCouponCode() {
  return readSessionValue(CART_COUPON_KEY) ?? '';
}

export function storeCartCouponCode(couponCode: string) {
  const normalizedCode = couponCode.trim().toUpperCase();

  if (!normalizedCode) {
    removeSessionValue(CART_COUPON_KEY);
    return;
  }

  writeSessionValue(CART_COUPON_KEY, normalizedCode);
}

export function clearCartCouponCode() {
  removeSessionValue(CART_COUPON_KEY);
}
