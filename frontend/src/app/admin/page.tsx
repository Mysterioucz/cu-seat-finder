import Link from "next/link";
import { getAuthUser } from "@/lib/authServer";
import { t } from "@/lib/i18n";
import { getLang } from "@/lib/langServer";
import AdminDashboard from "@/components/AdminDashboard";
import Card from "@/components/Card";

export default async function AdminPage() {
  const [lang, user] = await Promise.all([getLang(), getAuthUser()]);

  if (user?.role !== "admin") {
    return (
      <Card className="mx-auto mt-8 flex max-w-sm flex-col items-start gap-3 text-center sm:items-center">
        <p className="text-sm text-muted-foreground">{t(lang, "error.forbidden")}</p>
        <Link
          href="/login"
          className="inline-flex min-h-11 items-center rounded-lg bg-brand px-4 text-sm font-medium text-brand-foreground hover:opacity-90"
        >
          {t(lang, "nav.login")}
        </Link>
      </Card>
    );
  }

  return <AdminDashboard />;
}
