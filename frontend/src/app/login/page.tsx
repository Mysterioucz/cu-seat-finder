"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import { setToken } from "@/lib/auth";
import { useT } from "@/lib/LangContext";
import Button from "@/components/Button";
import Card from "@/components/Card";

type Tab = "cu" | "admin";

function fieldClass() {
  return "min-h-11 rounded-lg border border-border bg-surface px-3 py-2 text-sm outline-none focus-visible:border-brand";
}

export default function LoginPage() {
  const { t } = useT();
  const router = useRouter();
  const [tab, setTab] = useState<Tab>("cu");
  const [studentId, setStudentId] = useState("");
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function submitCu(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const res = await api<{ token: string }>("/api/auth/login", {
        method: "POST",
        body: JSON.stringify({ student_id: studentId, name: name || undefined }),
      });
      setToken(res.token);
      router.push("/");
      router.refresh();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("error.network"));
    } finally {
      setLoading(false);
    }
  }

  async function submitAdmin(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const res = await api<{ token: string }>("/api/auth/admin/login", {
        method: "POST",
        body: JSON.stringify({ email, password }),
      });
      setToken(res.token);
      router.push("/admin");
      router.refresh();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("error.network"));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="mx-auto max-w-sm">
      <h1 className="mb-4 text-xl font-bold">{t("auth.title")}</h1>
      <Card>
        <div role="tablist" aria-label={t("auth.title")} className="mb-4 flex gap-1 rounded-lg bg-surface-muted p-1">
          {(["cu", "admin"] as Tab[]).map((tb) => (
            <button
              key={tb}
              id={`tab-${tb}`}
              role="tab"
              type="button"
              aria-selected={tab === tb}
              aria-controls={`panel-${tb}`}
              tabIndex={tab === tb ? 0 : -1}
              onClick={() => {
                setTab(tb);
                setError(null);
              }}
              onKeyDown={(e) => {
                if (e.key !== "ArrowRight" && e.key !== "ArrowLeft") return;
                e.preventDefault();
                const next: Tab = tb === "cu" ? "admin" : "cu";
                setTab(next);
                setError(null);
                document.getElementById(`tab-${next}`)?.focus();
              }}
              className={`min-h-11 flex-1 rounded-md px-3 text-sm font-medium ${
                tab === tb ? "bg-surface shadow-sm text-foreground" : "text-muted-foreground"
              }`}
            >
              {t(tb === "cu" ? "auth.tab.cu" : "auth.tab.admin")}
            </button>
          ))}
        </div>

        {tab === "cu" ? (
          <form
            id="panel-cu"
            role="tabpanel"
            aria-labelledby="tab-cu"
            onSubmit={submitCu}
            className="flex flex-col gap-3"
          >
            <p className="text-xs text-muted-foreground">{t("auth.mockNote")}</p>
            <label className="flex flex-col gap-1 text-sm">
              {t("auth.studentId")}
              <input
                required
                value={studentId}
                onChange={(e) => setStudentId(e.target.value)}
                placeholder={t("auth.studentIdPlaceholder")}
                className={fieldClass()}
              />
            </label>
            <label className="flex flex-col gap-1 text-sm">
              {t("auth.name")}
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder={t("auth.namePlaceholder")}
                className={fieldClass()}
              />
            </label>
            {error && (
              <p role="alert" className="text-sm text-red-600">
                {error}
              </p>
            )}
            <Button type="submit" loading={loading}>
              {loading ? t("auth.loggingIn") : t("auth.loginCu")}
            </Button>
          </form>
        ) : (
          <form
            id="panel-admin"
            role="tabpanel"
            aria-labelledby="tab-admin"
            onSubmit={submitAdmin}
            className="flex flex-col gap-3"
          >
            <label className="flex flex-col gap-1 text-sm">
              {t("auth.email")}
              <input
                required
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className={fieldClass()}
              />
            </label>
            <label className="flex flex-col gap-1 text-sm">
              {t("auth.password")}
              <input
                required
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className={fieldClass()}
              />
            </label>
            {error && (
              <p role="alert" className="text-sm text-red-600">
                {error}
              </p>
            )}
            <Button type="submit" loading={loading}>
              {loading ? t("auth.loggingIn") : t("auth.loginAdmin")}
            </Button>
          </form>
        )}
      </Card>
    </div>
  );
}
