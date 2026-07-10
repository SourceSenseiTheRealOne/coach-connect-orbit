"use client";

import { trpc } from "@coach-connect/trpc-client";

export function ApiHealthStatus() {
  const health = trpc.system.health.useQuery(undefined, {
    retry: false,
    refetchOnWindowFocus: false,
  });

  const label = health.isPending
    ? "Checking API"
    : health.isSuccess
      ? `${health.data.service} online`
      : "API unavailable";

  return (
    <div
      aria-live="polite"
      className="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-white px-4 py-2 text-sm text-slate-600 dark:border-white/10 dark:bg-white/5 dark:text-slate-300"
    >
      <span
        aria-hidden="true"
        className={`size-2 rounded-full ${health.isSuccess ? "bg-lime-400" : "bg-slate-400 dark:bg-slate-600"}`}
      />
      {label}
    </div>
  );
}
