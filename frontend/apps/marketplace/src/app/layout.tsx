import { CoachConnectAuthProvider } from "@coach-connect/auth";
import { ThemeProvider } from "@coach-connect/ui";
import type { Metadata } from "next";
import type { ReactNode } from "react";
import "./globals.css";

export const metadata: Metadata = {
  title: "Coach Connect — Marketplace",
  description: "Football-first social network and marketplace",
};

export default function RootLayout({
  children,
}: Readonly<{ children: ReactNode }>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className="min-h-screen bg-white font-sans text-slate-950 antialiased dark:bg-slate-950 dark:text-slate-50">
        <CoachConnectAuthProvider
          publishableKey={process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY}
        >
          <ThemeProvider>{children}</ThemeProvider>
        </CoachConnectAuthProvider>
      </body>
    </html>
  );
}
