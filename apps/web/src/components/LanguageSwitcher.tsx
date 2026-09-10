'use client'

import { useRouter, usePathname } from 'next/navigation'
import { locales, localeNames, type Locale } from '../i18n/config'
import { useI18n } from './I18nProvider'

export default function LanguageSwitcher() {
  const { locale, dict } = useI18n()
  const router = useRouter()
  const pathname = usePathname()

  const switchTo = (next: Locale) => {
    if (next === locale) return

    // 替换路径中的语言段
    const segments = pathname.split('/')
    if (segments[1] === locale) {
      segments[1] = next
    } else {
      segments.splice(1, 0, next)
    }

    document.cookie = `NEXT_LOCALE=${next};path=/;max-age=31536000`
    router.push(segments.join('/') || `/${next}`)
  }

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
      <span style={{ fontSize: 13, color: 'var(--color-muted)' }}>
        {dict.language}
      </span>
      {locales.map((l) => (
        <button
          key={l}
          onClick={() => switchTo(l)}
          style={{
            border: '1px solid var(--color-border)',
            background: l === locale ? 'var(--color-accent)' : 'transparent',
            color: l === locale ? '#fff' : 'var(--color-muted)',
            borderRadius: 6,
            padding: '3px 10px',
            fontSize: 12,
            cursor: 'pointer',
          }}
        >
          {localeNames[l]}
        </button>
      ))}
    </div>
  )
}
