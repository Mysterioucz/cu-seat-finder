"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { clearToken } from "@/lib/auth";
import { useT } from "@/lib/LangContext";
import type { JwtClaims } from "@/lib/auth";

function NavLink({ href, active, children }: { href: string; active: boolean; children: React.ReactNode }) {
  return (
    <Link
      href={href}
      aria-current={active ? "page" : undefined}
      className={`flex min-h-11 flex-1 flex-col items-center justify-center gap-0.5 rounded-md px-2 py-1.5 text-xs font-medium md:flex-none md:flex-row md:gap-1.5 md:text-sm ${
        active ? "text-brand" : "text-muted-foreground hover:text-brand"
      }`}
    >
      {children}
    </Link>
  );
}

export default function AppShell({
  lang,
  user,
  children,
}: {
  lang: "th" | "en";
  user: JwtClaims | null;
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const router = useRouter();
  const { t, setLang } = useT();

  async function handleLogout() {
    try {
      await api("/api/auth/logout", { method: "POST" });
    } catch {
      // stateless token; logging out client-side is enough even if this call fails
    }
    clearToken();
    router.push("/login");
    router.refresh();
  }

  const links = [
    { href: "/", key: "nav.home" as const, show: true },
    { href: "/me/reservations", key: "nav.reservations" as const, show: !!user },
    { href: "/admin", key: "nav.admin" as const, show: user?.role === "admin" },
  ].filter((l) => l.show);

  return (
    <>
      {/* WCAG 2.4.1 Bypass Blocks: lets keyboard/screen-reader users skip the repeated nav. */}
      <a
        href="#main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-30 focus:rounded-md focus:bg-brand focus:px-3 focus:py-2 focus:text-sm focus:font-medium focus:text-brand-foreground"
      >
        {t("nav.skipToContent")}
      </a>
      <header className="sticky top-0 z-20 flex items-center justify-between border-b border-border bg-surface/95 px-4 py-3 backdrop-blur">
        <Link href="/" className="text-base font-bold text-brand">
          CU SeatFinder
        </Link>
        <nav aria-label="Main" className="hidden items-center gap-1 md:flex">
          {links.map((l) => (
            <NavLink key={l.href} href={l.href} active={pathname === l.href}>
              {t(l.key)}
            </NavLink>
          ))}
        </nav>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => setLang(lang === "th" ? "en" : "th")}
            className="min-h-11 rounded-md border border-border px-2.5 text-xs font-semibold text-foreground hover:border-brand"
            aria-label={lang === "th" ? "Switch to English" : "เปลี่ยนเป็นภาษาไทย"}
          >
            {lang === "th" ? "EN" : "TH"}
          </button>
          {user ? (
            <button
              type="button"
              onClick={handleLogout}
              className="min-h-11 px-1 text-sm font-medium text-muted-foreground hover:text-brand"
            >
              {t("nav.logout")}
            </button>
          ) : (
            <Link href="/login" className="min-h-11 px-1 py-2.5 text-sm font-medium text-brand">
              {t("nav.login")}
            </Link>
          )}
        </div>
      </header>

      <main id="main" className="mx-auto w-full max-w-3xl flex-1 px-4 pb-24 pt-4 md:pb-6">
        {children}
      </main>

      <nav
        aria-label="Main"
        className="fixed inset-x-0 bottom-0 z-20 flex border-t border-border bg-surface/95 backdrop-blur md:hidden"
        style={{ paddingBottom: "env(safe-area-inset-bottom)" }}
      >
        {links.map((l) => (
          <NavLink key={l.href} href={l.href} active={pathname === l.href}>
            {t(l.key)}
          </NavLink>
        ))}
      </nav>
    </>
  );
}
