import { createBrowserRouter } from 'react-router-dom';

import { AppShell } from '../app-shell/app-shell';
import { AccountLayout } from '../features/account/components/account-layout';
import { AccountOverviewPage } from '../features/account/pages/account-overview-page';
import { AddressesPage } from '../features/addresses/pages/addresses-page';
import { ForgotPasswordPage } from '../features/auth/pages/forgot-password-page';
import { LoginPage } from '../features/auth/pages/login-page';
import { OtpPage } from '../features/auth/pages/otp-page';
import { ResetPasswordPage } from '../features/auth/pages/reset-password-page';
import { SignupPage } from '../features/auth/pages/signup-page';
import { CartPage } from '../features/cart/pages/cart-page';
import { CheckoutPage } from '../features/checkout/pages/checkout-page';
import { PaymentResultPage } from '../features/checkout/pages/payment-result-page';
import { OrderDetailPage } from '../features/orders/pages/order-detail-page';
import { OrderListPage } from '../features/orders/pages/order-list-page';
import { CategoriesPage } from '../features/product/pages/categories-page';
import { CategoryPage } from '../features/product/pages/category-page';
import { HomePage } from '../features/product/pages/home-page';
import { ProductDetailPage } from '../features/product/pages/product-detail-page';
import { NotificationPreferencesPage } from '../features/profile/pages/notification-preferences-page';
import { ProfilePage } from '../features/profile/pages/profile-page';
import { SearchPage } from '../features/search/pages/search-page';
import { WishlistPage } from '../features/wishlist/pages/wishlist-page';
import { PlaceholderPage } from './placeholder-page';
import { ProtectedRoute } from './protected-route';
import { routePaths } from './route-paths';

export const router = createBrowserRouter([
  {
    element: <AppShell />,
    children: [
      { path: routePaths.home, element: <HomePage /> },
      {
        path: routePaths.categories,
        element: <CategoriesPage />,
      },
      { path: routePaths.categoryPattern, element: <CategoryPage /> },
      { path: routePaths.deals, element: <PlaceholderPage title="Deals" /> },
      {
        path: routePaths.wishlist,
        element: (
          <ProtectedRoute>
            <WishlistPage />
          </ProtectedRoute>
        ),
      },
      {
        path: routePaths.cart,
        element: (
          <ProtectedRoute>
            <CartPage />
          </ProtectedRoute>
        ),
      },
      {
        path: routePaths.checkout,
        element: (
          <ProtectedRoute>
            <CheckoutPage />
          </ProtectedRoute>
        ),
      },
      {
        path: routePaths.checkoutSuccess,
        element: (
          <ProtectedRoute>
            <PaymentResultPage result="success" />
          </ProtectedRoute>
        ),
      },
      {
        path: routePaths.checkoutFailure,
        element: (
          <ProtectedRoute>
            <PaymentResultPage result="failure" />
          </ProtectedRoute>
        ),
      },
      { path: routePaths.search, element: <SearchPage /> },
      {
        path: routePaths.productDetailPattern,
        element: <ProductDetailPage />,
      },
      { path: routePaths.login, element: <LoginPage /> },
      { path: routePaths.signup, element: <SignupPage /> },
      { path: routePaths.otp, element: <OtpPage /> },
      {
        path: routePaths.forgotPassword,
        element: <ForgotPasswordPage />,
      },
      {
        path: routePaths.resetPassword,
        element: <ResetPasswordPage />,
      },
      {
        path: routePaths.account,
        element: (
          <ProtectedRoute>
            <AccountLayout />
          </ProtectedRoute>
        ),
        children: [
          { index: true, element: <AccountOverviewPage /> },
          { path: 'profile', element: <ProfilePage /> },
          { path: 'addresses', element: <AddressesPage /> },
          { path: 'orders', element: <OrderListPage /> },
          { path: 'orders/:orderId', element: <OrderDetailPage /> },
          {
            path: 'notifications',
            element: <NotificationPreferencesPage />,
          },
        ],
      },
    ],
  },
]);
