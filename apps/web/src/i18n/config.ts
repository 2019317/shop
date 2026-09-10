export const locales = ['en', 'zh'] as const
export type Locale = (typeof locales)[number]

export const defaultLocale: Locale = 'en'

export const localeNames: Record<Locale, string> = {
  en: 'English',
  zh: '中文',
}

export function isLocale(value: string): value is Locale {
  return (locales as readonly string[]).includes(value)
}

// 语言标签，用于 Intl 与 <html lang>
export const localeTags: Record<Locale, string> = {
  en: 'en-US',
  zh: 'zh-CN',
}
