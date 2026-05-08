"use client";

import { useEffect } from "react";
import { refreshSession } from "@/lib/auth";
import { ApiError } from "@/lib/api";
import { createAnonymousAuthSnapshot, resolveApiAuthSnapshot } from "@/lib/auth-state";
import { useAuthStore } from "@/store/auth-store";

export function AuthBootstrap() {
  const initialized = useAuthStore((state) => state.initialized);
  const setLoading = useAuthStore((state) => state.setLoading);
  const hydrate = useAuthStore((state) => state.hydrate);

  useEffect(() => {
    if (initialized) {
      return;
    }

    let isCancelled = false;

    async function bootstrap() {
      setLoading();

      try {
        const auth = await refreshSession();
        if (isCancelled) {
          return;
        }
        hydrate(resolveApiAuthSnapshot(auth.user, auth.access_expires_at));
      } catch (error) {
        if (isCancelled) {
          return;
        }

        if (error instanceof ApiError && (error.status === 401 || error.status === 403)) {
          hydrate(createAnonymousAuthSnapshot());
          return;
        }

        hydrate(createAnonymousAuthSnapshot());
      }
    }

    void bootstrap();

    return () => {
      isCancelled = true;
    };
  }, [hydrate, initialized, setLoading]);

  return null;
}
