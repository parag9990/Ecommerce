import { describe, expect, it } from 'vitest';

import { useUiStore } from './ui-store';

describe('useUiStore', () => {
  it('stores shared app shell state', () => {
    useUiStore.getState().setMobileNavOpen(true);
    useUiStore.getState().setAccountMenuOpen(true);

    expect(useUiStore.getState().isMobileNavOpen).toBe(true);
    expect(useUiStore.getState().isAccountMenuOpen).toBe(true);
  });
});
