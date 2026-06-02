import { create } from "zustand";

import type { SellerSummary } from "../api/seller-session-api";

type SellerStore = {
  activeSeller: SellerSummary | null;
  sidebarCollapsed: boolean;
  mobileSidebarOpen: boolean;
  setActiveSeller: (seller: SellerSummary | null) => void;
  toggleSidebar: () => void;
  openMobileSidebar: () => void;
  closeMobileSidebar: () => void;
  reset: () => void;
};

const initialState = {
  activeSeller: null,
  sidebarCollapsed: false,
  mobileSidebarOpen: false,
};

export const useSellerStore = create<SellerStore>((set) => ({
  ...initialState,
  setActiveSeller: (seller) => set({ activeSeller: seller }),
  toggleSidebar: () =>
    set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
  openMobileSidebar: () => set({ mobileSidebarOpen: true }),
  closeMobileSidebar: () => set({ mobileSidebarOpen: false }),
  reset: () => set(initialState),
}));
