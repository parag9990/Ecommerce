import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, expect, test } from 'vitest';
import { RouterProvider, createMemoryRouter } from 'react-router-dom';

import { routePaths } from '../../../routes/route-paths';
import { createTestQueryWrapper } from '../../../test/create-test-query-wrapper';
import { LoginPage } from './login-page';

afterEach(() => {
  cleanup();
});

function renderLoginPage() {
  const router = createMemoryRouter(
    [{ element: <LoginPage />, path: routePaths.login }],
    {
      initialEntries: [routePaths.login],
    },
  );

  return render(<RouterProvider router={router} />, {
    wrapper: createTestQueryWrapper(),
  });
}

test('shows validation errors for an empty login form', async () => {
  const user = userEvent.setup();

  renderLoginPage();

  await user.click(screen.getByRole('button', { name: 'Login' }));

  expect(await screen.findByText('Email or phone is required.')).toBeTruthy();
  expect(await screen.findByText('Password is required.')).toBeTruthy();
});
