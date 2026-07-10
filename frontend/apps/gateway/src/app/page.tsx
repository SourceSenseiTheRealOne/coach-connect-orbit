import { ZoneShell } from "@coach-connect/ui";
import { ApiHealthStatus } from "@/components/api-health-status";

export default function HomePage() {
  return (
    <ZoneShell
      description="The final public home page and its sections are intentionally deferred. This temporary gateway surface only exposes the verified scaffold."
      eyebrow="MVP v2 · Gateway zone"
      title="Coach Connect foundation"
    >
      <div className="flex flex-wrap items-center gap-3 text-sm">
        <a
          className="rounded-xl bg-slate-950 px-5 py-3 font-semibold text-white transition hover:bg-slate-800 dark:bg-lime-300 dark:text-slate-950 dark:hover:bg-lime-200"
          href="/dashboard"
        >
          Open dashboard
        </a>
        <a
          className="rounded-xl border border-slate-200 px-5 py-3 font-medium transition hover:bg-slate-50 dark:border-white/10 dark:hover:bg-white/5"
          href="/feed"
        >
          Social zone
        </a>
        <a
          className="rounded-xl border border-slate-200 px-5 py-3 font-medium transition hover:bg-slate-50 dark:border-white/10 dark:hover:bg-white/5"
          href="/marketplace"
        >
          Marketplace zone
        </a>
        <ApiHealthStatus />
      </div>
    </ZoneShell>
  );
}
