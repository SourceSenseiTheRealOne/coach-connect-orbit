import type { ReactNode } from "react";
import { CrossZoneLink } from "./cross-zone-link";

export interface ZoneShellProps {
  eyebrow: string;
  title: string;
  description: string;
  children?: ReactNode;
}

export function ZoneShell({
  eyebrow,
  title,
  description,
  children,
}: ZoneShellProps) {
  return (
    <main className="min-h-screen bg-white px-6 py-10 text-slate-950 dark:bg-slate-950 dark:text-slate-50 sm:px-10 lg:px-16">
      <nav
        aria-label="Primary"
        className="mx-auto flex max-w-6xl items-center justify-between border-b border-slate-200 pb-6 dark:border-white/10"
      >
        <CrossZoneLink className="font-semibold tracking-tight" href="/">
          Coach Connect
        </CrossZoneLink>
        <div className="flex gap-5 text-sm text-slate-500 dark:text-slate-400">
          <CrossZoneLink href="/feed">Social</CrossZoneLink>
          <CrossZoneLink href="/marketplace">Marketplace</CrossZoneLink>
        </div>
      </nav>
      <section className="mx-auto grid min-h-[70vh] max-w-6xl place-items-center py-16">
        <div className="max-w-3xl">
          <p className="mb-4 font-mono text-xs font-semibold uppercase tracking-[0.22em] text-lime-600 dark:text-lime-400">
            {eyebrow}
          </p>
          <h1 className="text-balance text-5xl font-semibold tracking-[-0.04em] sm:text-7xl">
            {title}
          </h1>
          <p className="mt-6 max-w-2xl text-lg leading-8 text-slate-600 dark:text-slate-400">
            {description}
          </p>
          {children ? <div className="mt-10">{children}</div> : null}
        </div>
      </section>
    </main>
  );
}
