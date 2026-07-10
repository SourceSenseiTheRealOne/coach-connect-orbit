import { auth } from "@clerk/nextjs/server";
import { redirect } from "next/navigation";

export const dynamic = "force-dynamic";

export default async function DashboardPage() {
  const isClerkConfigured = Boolean(
    process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY &&
    process.env.CLERK_SECRET_KEY,
  );

  if (!isClerkConfigured) {
    redirect("/sign-in?redirect_url=/dashboard");
  }

  const { isAuthenticated } = await auth();

  if (!isAuthenticated) {
    redirect("/sign-in?redirect_url=/dashboard");
  }

  return (
    <section className="mx-auto min-h-[calc(100vh-4rem)] max-w-7xl px-5 py-8 sm:px-8 lg:px-10">
      <p className="text-xs font-semibold uppercase tracking-[0.18em] text-lime-600 dark:text-lime-400">
        Workspace
      </p>
      <h1 className="mt-2 text-2xl font-semibold tracking-tight text-slate-950 dark:text-white">
        Dashboard
      </h1>
    </section>
  );
}
