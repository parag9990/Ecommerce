export const routePaths = {
  home: '/',
  account: '/account',
  addresses: '/account/addresses',
  categories: '/categories',
  category: (categoryId: string) => `/category/${encodeURIComponent(categoryId)}`,
  categoryPattern: '/category/:categoryId',
  deals: '/deals',
  notificationPreferences: '/account/notifications',
  orders: '/account/orders',
  orderDetail: (orderId: string) =>
    `/account/orders/${encodeURIComponent(orderId)}`,
  orderDetailPattern: '/account/orders/:orderId',
  profile: '/account/profile',
  productDetail: (productId: string) =>
    `/products/${encodeURIComponent(productId)}`,
  productDetailPattern: '/products/:productId',
  wishlist: '/wishlist',
  cart: '/cart',
  checkout: '/checkout',
  checkoutFailure: '/checkout/failure',
  checkoutSuccess: '/checkout/success',
  forgotPassword: '/forgot-password',
  login: '/login',
  otp: '/auth/otp',
  resetPassword: '/reset-password',
  search: '/search',
  signup: '/signup',
} as const;

export type RoutePath = Extract<(typeof routePaths)[keyof typeof routePaths], string>;
