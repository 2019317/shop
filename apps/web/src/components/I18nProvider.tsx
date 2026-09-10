'use client'

import { createContext, useContext } from 'react'
import type { Locale } from '../i18n/config'
import type { Dictionary } from '../i18n/dictionaries'

interface I18nContextValue {
  locale: Locale
  dict: Dictionary
  // 生成带语言前缀的链接，避免各处手工拼接
  href: (path: string) => string
}

const I18nContext = createContext<I18nContextValue | null>(null)

export function I18nProvider({
  locale,
  dict,
  children,
}: {
  locale: Locale
  dict: Dictionary
  children: React.ReactNode
}) {
  const value: I18nContextValue = {
    locale,
    dict,
    href: (path: string) => `/${locale}${path === '/' ? '' : path}`,
  }
  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>
}

export function useI18n() {
  const ctx = useContext(I18nContext)
  if (!ctx) {
    throw new Error('useI18n must be used within I18nProvider')
  }
  return ctx
}
