"use client";

import { createContext, useCallback, useContext, useMemo } from "react";
import { DEFAULT_LANG, LANG_COOKIE, t as translate, type DictKey, type Lang } from "./i18n";

const LangContext = createContext<Lang>(DEFAULT_LANG);

export function LangProvider({ lang, children }: { lang: Lang; children: React.ReactNode }) {
  return <LangContext.Provider value={lang}>{children}</LangContext.Provider>;
}

/** Returns [lang, t, setLang]. setLang writes the cookie and reloads so both
 * server- and client-rendered text pick up the new language. */
export function useT() {
  const lang = useContext(LangContext);
  const t = useCallback((key: DictKey, vars?: Record<string, string | number>) => translate(lang, key, vars), [lang]);
  const setLang = useMemo(
    () => (next: Lang) => {
      document.cookie = `${LANG_COOKIE}=${next}; path=/; max-age=31536000`;
      window.location.reload();
    },
    [],
  );
  return { lang, t, setLang };
}
