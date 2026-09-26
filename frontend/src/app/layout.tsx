import type { Metadata } from "next";
import { Anuphan, Geist_Mono } from "next/font/google";
import { cookies } from "next/headers";
import "./globals.css";
import { getLang } from "@/lib/langServer";
import { LangProvider } from "@/lib/LangContext";
import { decodeToken, TOKEN_COOKIE } from "@/lib/auth";
import AppShell from "@/components/AppShell";

// Anuphan: one variable font with native Thai shapes + full Latin coverage, so th/en
// text sits at the same visual weight instead of pairing a Thai fallback with a Latin webfont.
const anuphan = Anuphan({
  variable: "--font-anuphan",
  subsets: ["thai", "latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "CU SeatFinder",
  description: "Live library occupancy, meeting room booking, and reviews for CU libraries.",
};

export default async function RootLayout({ children }: { children: React.ReactNode }) {
  const lang = await getLang();
  const store = await cookies();
  const token = store.get(TOKEN_COOKIE)?.value;
  const user = token ? decodeToken(token) : null;

  return (
    <html lang={lang} className={`${anuphan.variable} ${geistMono.variable} h-full antialiased`}>
      <body className="min-h-full flex flex-col font-sans">
        <LangProvider lang={lang}>
          <AppShell lang={lang} user={user}>
            {children}
          </AppShell>
        </LangProvider>
      </body>
    </html>
  );
}
