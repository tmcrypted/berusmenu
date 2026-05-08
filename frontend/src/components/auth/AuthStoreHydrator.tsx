"use client";

import type { ReactNode } from "react";
import { createContext, useContext, useRef } from "react";
import { useShallow } from "zustand/react/shallow";
import type { AuthSnapshot } from "@/lib/auth-state";
import { useAuthStore } from "@/store/auth-store";

type Props = {
  snapshot: AuthSnapshot;
  children: ReactNode;
};

const AuthSnapshotContext = createContext<AuthSnapshot | null>(null);

export function AuthStoreHydrator({ snapshot, children }: Props) {
  const signature = JSON.stringify(snapshot);
  const appliedSignatureRef = useRef<string | null>(null);

  if (appliedSignatureRef.current !== signature) {
    useAuthStore.setState(snapshot);
    appliedSignatureRef.current = signature;
  }

  return <AuthSnapshotContext.Provider value={snapshot}>{children}</AuthSnapshotContext.Provider>;
}

export function useResolvedAuthSnapshot(): AuthSnapshot {
  const fallback = useContext(AuthSnapshotContext);
  const state = useAuthStore(
    useShallow((store) => ({
      user: store.user,
      status: store.status,
      initialized: store.initialized,
      accessExpiresAt: store.accessExpiresAt,
    })),
  );

  if ((!state.initialized || state.status === "loading") && fallback) {
    return fallback;
  }

  return state;
}
