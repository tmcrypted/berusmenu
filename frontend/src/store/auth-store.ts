import { create } from "zustand";
import {
  createAnonymousAuthSnapshot,
  createLoadingAuthSnapshot,
  resolveUserAuthSnapshot,
  type AuthSnapshot,
} from "@/lib/auth-state";
import type { User } from "@/types/auth";

type AuthState = {
  user: AuthSnapshot["user"];
  status: AuthSnapshot["status"];
  initialized: AuthSnapshot["initialized"];
  accessExpiresAt: AuthSnapshot["accessExpiresAt"];
  setLoading: () => void;
  setAnonymous: () => void;
  setUnverified: (user: User, accessExpiresAt?: string | null) => void;
  setIncomplete: (user: User) => void;
  setReady: (user: User) => void;
  hydrate: (snapshot: AuthSnapshot) => void;
  clearAuth: () => void;
};

export const useAuthStore = create<AuthState>((set) => ({
  ...createLoadingAuthSnapshot(),
  setLoading: () => set(createLoadingAuthSnapshot()),
  setAnonymous: () => set(createAnonymousAuthSnapshot()),
  setUnverified: (user, accessExpiresAt) =>
    set(resolveUserAuthSnapshot(user, accessExpiresAt)),
  setIncomplete: (user) =>
    set({
      user,
      status: "authenticated_profile_incomplete",
      initialized: true,
      accessExpiresAt: null,
    }),
  setReady: (user) =>
    set({
      user,
      status: "authenticated_ready",
      initialized: true,
      accessExpiresAt: null,
    }),
  hydrate: (snapshot) => set(snapshot),
  clearAuth: () => set(createAnonymousAuthSnapshot()),
}));
