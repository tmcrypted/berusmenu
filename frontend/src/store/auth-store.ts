import { create } from "zustand";

type AuthState = {
  user: unknown | null;
  authStatus: "anonymous" | "authenticated";
  setUser: (user: unknown | null) => void;
  clearAuth: () => void;
};

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  authStatus: "anonymous",
  setUser: (user) => set({ user, authStatus: user ? "authenticated" : "anonymous" }),
  clearAuth: () => set({ user: null, authStatus: "anonymous" }),
}));
