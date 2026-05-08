"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useResolvedAuthSnapshot } from "@/components/auth/AuthStoreHydrator";
import {
  GuestHomeContent,
  OwnerHomeContent,
  WorkerHomeContent,
} from "@/components/home/HomePanels";
import { profileCompletionPath } from "@/lib/profile";

export default function Home() {
  const router = useRouter();
  const { status, user } = useResolvedAuthSnapshot();

  useEffect(() => {
    if (status === "authenticated_unverified") {
      router.replace("/verify-email");
      return;
    }

    if (status === "authenticated_profile_incomplete" && user) {
      router.replace(profileCompletionPath(user.role));
    }
  }, [router, status, user]);

  const readyOwner = status === "authenticated_ready" && user?.role === "owner";
  const readyWorker =
    status === "authenticated_ready" && user && user.role !== "owner";
  const isBootstrapping = status === "loading";

  return (
    <div className="app-shell">
      <section className="mx-auto max-w-7xl px-4 py-6 sm:px-6 sm:py-8 lg:px-8">
        {isBootstrapping ? (
          <div className="grid gap-4 lg:grid-cols-[minmax(0,1.25fr)_minmax(20rem,0.75fr)]">
            <div className="min-h-[18rem] rounded-[2rem] border border-[var(--border)] bg-[var(--panel)] shadow-[var(--shadow-soft)]" />
            <div className="min-h-[18rem] rounded-[2rem] border border-[var(--border)] bg-[var(--card)] shadow-[var(--shadow-soft)]" />
          </div>
        ) : readyOwner && user ? (
          <OwnerHomeContent user={user} />
        ) : readyWorker && user ? (
          <WorkerHomeContent user={user} />
        ) : (
          <GuestHomeContent status={status} user={user} />
        )}
      </section>
    </div>
  );
}
