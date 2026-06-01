import { create } from 'zustand';

export type ToastTone = 'error' | 'info' | 'success';

export type Toast = {
  id: string;
  message: string;
  tone: ToastTone;
};

type UiState = {
  isAccountMenuOpen: boolean;
  isMobileNavOpen: boolean;
  pushToast: (toast: Omit<Toast, 'id'>) => void;
  removeToast: (toastId: string) => void;
  setAccountMenuOpen: (isOpen: boolean) => void;
  setMobileNavOpen: (isOpen: boolean) => void;
  toasts: Toast[];
};

export const useUiStore = create<UiState>((set) => ({
  isAccountMenuOpen: false,
  isMobileNavOpen: false,
  pushToast: (toast) => {
    set((state) => ({
      toasts: [
        ...state.toasts,
        {
          ...toast,
          id: crypto.randomUUID(),
        },
      ],
    }));
  },
  removeToast: (toastId) => {
    set((state) => ({
      toasts: state.toasts.filter((toast) => toast.id !== toastId),
    }));
  },
  setAccountMenuOpen: (isAccountMenuOpen) => {
    set({ isAccountMenuOpen });
  },
  setMobileNavOpen: (isMobileNavOpen) => {
    set({ isMobileNavOpen });
  },
  toasts: [],
}));
