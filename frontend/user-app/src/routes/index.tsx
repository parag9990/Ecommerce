import { createBrowserRouter } from 'react-router-dom';

import { AppShell } from '../app-shell/app-shell';
import { AccountLayout } from '../features/account/components/account-layout';
import { ProtectedRoute } from './protected-route';
import {
  AccountOverviewPage,
  AddressesPage,
  CartPage,
  CategoriesPage,
  CategoryPage,
  CheckoutPage,
	DealsPage,
  ForgotPasswordPage,
  HomePage,
  LoginPage,
  NotificationPreferencesPage,
  OrderDetailPage,
  OrderListPage,
  OtpPage,
  PaymentResultPage,
  ProductDetailPage,
  ProfilePage,
  ResetPasswordPage,
  SearchPage,
  SignupPage,
  WishlistPage,
} from './lazy-pages';
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
      { path: routePaths.deals, element: <DealsPage /> },
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
