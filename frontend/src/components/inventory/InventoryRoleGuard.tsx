"use client";

import type { ReactNode } from "react";
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useResolvedAuthSnapshot } from "@/components/auth/AuthStoreHydrator";
import { Spinner } from "@/components/ui/Spinner";
import { profileCompletionPath } from "@/lib/profile";
import type { UserRole } from "@/types/auth";

type Props = {
  children: ReactNode;
  role: Exclude<UserRole, "unspecified">;
};

export function InventoryRoleGuard({ children, role }: Props) {
  const router = useRouter();
  const { initialized, status, user } = useResolvedAuthSnapshot();

  useEffect(() => {
    if (!initialized || status === "loading") {
      return;
    }
    if (status === "anonymous") {
      router.replace("/login");
      return;
    }
    if (status === "authenticated_unverified") {
      router.replace("/verify-email");
      return;
    }
    if (status === "authenticated_profile_incomplete" && user) {
      router.replace(profileCompletionPath(user.role));
      return;
    }
    if (!user || user.role !== role) {
      router.replace("/");
    }
  }, [initialized, role, router, status, user]);

  if (!initialized || status === "loading") {
    return (
      <div className="flex min-h-[calc(100vh_-_var(--header-height))] items-center justify-center">
        <Spinner />
      </div>
    );
  }

  if (status === "authenticated_ready" && user?.role === role) {
    return <>{children}</>;
  }

  return (
    <div className="flex min-h-[calc(100vh_-_var(--header-height))] items-center justify-center">
      <Spinner />
    </div>
  );
}
