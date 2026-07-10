import Link from "next/link";

export function AuthConfigurationRequired() {
  return (
    <main className="grid min-h-screen place-items-center bg-slate-50 px-4 py-12 dark:bg-slate-950">
      <section className="w-full max-w-lg rounded-3xl border border-slate-200 bg-white p-8 shadow-xl shadow-slate-200/40 dark:border-white/10 dark:bg-white/5 dark:shadow-none sm:p-10">
        <span className="inline-flex rounded-full bg-amber-100 px-3 py-1 text-xs font-semibold text-amber-800 dark:bg-amber-400/10 dark:text-amber-300">
          Local setup
        </span>
        <h1 className="mt-5 text-2xl font-semibold tracking-tight text-slate-950 dark:text-white">
          Clerk configuration required
        </h1>
        <p className="mt-3 text-sm leading-6 text-slate-600 dark:text-slate-300">
          Add your development Clerk keys before signing in. Route protection
          remains enabled while authentication is unconfigured.
        </p>
        <div className="mt-6 rounded-2xl bg-slate-950 p-4 font-mono text-xs leading-6 text-slate-200 dark:bg-black/40">
          <p>NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY</p>
          <p>CLERK_SECRET_KEY</p>
        </div>
        <p className="mt-4 text-xs leading-5 text-slate-500 dark:text-slate-400">
          Copy{" "}
          <code className="font-semibold text-slate-700 dark:text-slate-200">
            frontend/.env.example
          </code>{" "}
          to a local environment file and replace the placeholders. Never commit
          real keys.
        </p>
        <Link
          className="mt-7 inline-flex rounded-xl bg-slate-950 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-slate-800 dark:bg-lime-300 dark:text-slate-950 dark:hover:bg-lime-200"
          href="/"
        >
          Back to gateway
        </Link>
      </section>
    </main>
  );
}
