import { ThemeToggle } from "@coach-connect/ui";
import { UserButton } from "@clerk/nextjs";
import {
  CircleUserRound,
  LayoutDashboard,
  MessageCircleMore,
  Search,
  ShieldCheck,
  Store,
  UsersRound,
} from "lucide-react";
import Link from "next/link";
import type { ReactNode } from "react";

const navigation = [
  {
    label: "Dashboard",
    icon: LayoutDashboard,
    href: "/dashboard",
    available: true,
  },
  {
    label: "Profile",
    icon: CircleUserRound,
    href: "/profile",
    available: false,
  },
  { label: "Network", icon: UsersRound, href: "/network", available: false },
  {
    label: "Messages",
    icon: MessageCircleMore,
    href: "/messages",
    available: false,
  },
  { label: "Marketplace", icon: Store, href: "/marketplace", available: false },
] as const;

export default function DashboardLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  const isClerkConfigured = Boolean(
    process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY &&
    process.env.CLERK_SECRET_KEY,
  );

  return (
    <div className="min-h-screen bg-slate-50 text-slate-950 dark:bg-slate-950 dark:text-slate-50">
      <aside className="fixed inset-y-0 left-0 z-40 hidden w-72 flex-col border-r border-slate-200 bg-white/90 px-4 py-5 backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/90 lg:flex">
        <Link className="flex items-center gap-3 px-3" href="/dashboard">
          <span className="grid size-10 place-items-center rounded-xl bg-slate-950 text-lime-300 shadow-sm shadow-slate-950/20 dark:bg-lime-300 dark:text-slate-950">
            <ShieldCheck
              aria-hidden="true"
              className="size-5"
              strokeWidth={2.2}
            />
          </span>
          <span>
            <span className="block text-sm font-semibold tracking-tight">
              Coach Connect
            </span>
            <span className="block text-xs text-slate-500 dark:text-slate-400">
              Football workspace
            </span>
          </span>
        </Link>

        <nav
          aria-label="Dashboard"
          className="mt-10 flex flex-1 flex-col gap-1"
        >
          <p className="mb-2 px-3 text-[0.68rem] font-semibold uppercase tracking-[0.18em] text-slate-400 dark:text-slate-500">
            Workspace
          </p>
          {navigation.map(({ label, icon: Icon, href, available }) =>
            available ? (
              <Link
                className="flex items-center gap-3 rounded-xl bg-slate-950 px-3 py-2.5 text-sm font-medium text-white shadow-sm dark:bg-white dark:text-slate-950"
                href={href}
                key={label}
              >
                <Icon aria-hidden="true" className="size-4.5" />
                {label}
              </Link>
            ) : (
              <span
                aria-disabled="true"
                className="flex cursor-not-allowed items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium text-slate-400 dark:text-slate-600"
                key={label}
                title="Coming later"
              >
                <Icon aria-hidden="true" className="size-4.5" />
                {label}
                <span className="ml-auto text-[0.62rem] font-semibold uppercase tracking-wider">
                  Soon
                </span>
              </span>
            ),
          )}
        </nav>

        <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 dark:border-white/10 dark:bg-white/5">
          <p className="text-xs font-medium text-slate-700 dark:text-slate-200">
            Foundation mode
          </p>
          <p className="mt-1 text-xs leading-5 text-slate-500 dark:text-slate-400">
            Product modules will appear here as each verified slice ships.
          </p>
        </div>
      </aside>

      <div className="lg:pl-72">
        <header className="sticky top-0 z-30 flex h-16 items-center gap-3 border-b border-slate-200 bg-white/80 px-4 backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/80 sm:px-6 lg:px-8">
          <Link
            className="mr-auto flex items-center gap-2 font-semibold tracking-tight lg:hidden"
            href="/dashboard"
          >
            <span className="grid size-8 place-items-center rounded-lg bg-slate-950 text-lime-300 dark:bg-lime-300 dark:text-slate-950">
              <ShieldCheck aria-hidden="true" className="size-4" />
            </span>
            Coach Connect
          </Link>

          <div className="relative mr-auto hidden max-w-sm flex-1 md:block lg:mr-0">
            <Search
              aria-hidden="true"
              className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-slate-400"
            />
            <input
              aria-label="Search"
              className="h-9 w-full rounded-xl border border-slate-200 bg-slate-50 pl-9 pr-3 text-sm outline-none transition placeholder:text-slate-400 focus:border-slate-400 focus:ring-2 focus:ring-slate-200 dark:border-white/10 dark:bg-white/5 dark:focus:border-white/20 dark:focus:ring-white/10"
              disabled
              placeholder="Search is coming later"
              type="search"
            />
          </div>

          <div className="ml-auto flex items-center gap-2">
            <ThemeToggle />
            {isClerkConfigured ? (
              <UserButton
                appearance={{
                  elements: {
                    avatarBox:
                      "size-9 ring-1 ring-slate-200 dark:ring-white/15",
                  },
                }}
              />
            ) : (
              <span className="rounded-lg bg-amber-100 px-2.5 py-1.5 text-xs font-semibold text-amber-800 dark:bg-amber-400/10 dark:text-amber-300">
                Auth setup required
              </span>
            )}
          </div>
        </header>
        <main>{children}</main>
      </div>
    </div>
  );
}
