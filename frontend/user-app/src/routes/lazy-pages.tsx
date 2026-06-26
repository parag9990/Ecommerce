import { lazy } from 'react';

export const AccountOverviewPage = lazy(() =>
  import('../features/account/pages/account-overview-page').then((module) => ({
    default: module.AccountOverviewPage,
  })),
);
export const AddressesPage = lazy(() =>
  import('../features/addresses/pages/addresses-page').then((module) => ({
    default: module.AddressesPage,
  })),
);
export const ForgotPasswordPage = lazy(() =>
  import('../features/auth/pages/forgot-password-page').then((module) => ({
    default: module.ForgotPasswordPage,
  })),
);
export const LoginPage = lazy(() =>
  import('../features/auth/pages/login-page').then((module) => ({
    default: module.LoginPage,
  })),
);
export const OtpPage = lazy(() =>
  import('../features/auth/pages/otp-page').then((module) => ({
    default: module.OtpPage,
  })),
);
export const ResetPasswordPage = lazy(() =>
  import('../features/auth/pages/reset-password-page').then((module) => ({
    default: module.ResetPasswordPage,
  })),
);
export const SignupPage = lazy(() =>
  import('../features/auth/pages/signup-page').then((module) => ({
    default: module.SignupPage,
  })),
);
export const CartPage = lazy(() =>
  import('../features/cart/pages/cart-page').then((module) => ({
    default: module.CartPage,
  })),
);
export const CheckoutPage = lazy(() =>
  import('../features/checkout/pages/checkout-page').then((module) => ({
    default: module.CheckoutPage,
  })),
);
export const PaymentResultPage = lazy(() =>
  import('../features/checkout/pages/payment-result-page').then((module) => ({
    default: module.PaymentResultPage,
  })),
);
export const OrderDetailPage = lazy(() =>
  import('../features/orders/pages/order-detail-page').then((module) => ({
    default: module.OrderDetailPage,
  })),
);
export const OrderListPage = lazy(() =>
  import('../features/orders/pages/order-list-page').then((module) => ({
    default: module.OrderListPage,
  })),
);
export const CategoriesPage = lazy(() =>
  import('../features/product/pages/categories-page').then((module) => ({
    default: module.CategoriesPage,
  })),
);
export const DealsPage = lazy(() =>
  import('../features/product/pages/deals-page').then((module) => ({
    default: module.DealsPage,
  })),
);
export const CategoryPage = lazy(() =>
  import('../features/product/pages/category-page').then((module) => ({
    default: module.CategoryPage,
  })),
);
export const HomePage = lazy(() =>
  import('../features/product/pages/home-page').then((module) => ({
    default: module.HomePage,
  })),
);
export const ProductDetailPage = lazy(() =>
  import('../features/product/pages/product-detail-page').then((module) => ({
    default: module.ProductDetailPage,
  })),
);
export const NotificationPreferencesPage = lazy(() =>
  import('../features/profile/pages/notification-preferences-page').then(
    (module) => ({ default: module.NotificationPreferencesPage }),
  ),
);
export const ProfilePage = lazy(() =>
  import('../features/profile/pages/profile-page').then((module) => ({
    default: module.ProfilePage,
  })),
);
export const SearchPage = lazy(() =>
  import('../features/search/pages/search-page').then((module) => ({
    default: module.SearchPage,
  })),
);
export const WishlistPage = lazy(() =>
  import('../features/wishlist/pages/wishlist-page').then((module) => ({
    default: module.WishlistPage,
  })),
);
