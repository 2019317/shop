import { createContext, useContext, useEffect, useMemo, useState } from 'react'
import {
  detectInitialLocale,
  getDict,
  type Dictionary,
  type Locale,
} from './index'

interface I18nContextValue {
  locale: Locale
  dict: Dictionary
  setLocale: (l: Locale) => void
}

const I18nContext = createContext<I18nContextValue | null>(null)

export function I18nProvider({ children }: { children: React.ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>(() => detectInitialLocale())

  useEffect(() => {
    localStorage.setItem('admin_locale', locale)
    document.documentElement.lang = locale === 'zh' ? 'zh-CN' : 'en-US'
  }, [locale])

  const value = useMemo<I18nContextValue>(
    () => ({
      locale,
      dict: getDict(locale),
      setLocale: (l: Locale) => setLocaleState(l),
    }),
    [locale],
  )

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>
}

export function useI18n() {
  const ctx = useContext(I18nContext)
  if (!ctx) {
    throw new Error('useI18n must be used within I18nProvider')
  }
  return ctx
}
