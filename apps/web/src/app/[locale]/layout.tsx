import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import Link from 'next/link'
import CartProvider from '../../components/CartProvider'
import { I18nProvider } from '../../components/I18nProvider'
import LanguageSwitcher from '../../components/LanguageSwitcher'
import { getDictionarySync } from '../../i18n/dictionaries'
import { isLocale, localeTags, locales, type Locale } from '../../i18n/config'
import '../../app/globals.css'

export function generateStaticParams() {
  return locales.map((locale) => ({ locale }))
}

export const metadata: Metadata = {
  title: 'Stationery Shop — Handcrafted Journals & Paper Goods',
  description:
    'Thoughtfully designed journals, planner inserts, stickers and paper goods, shipped worldwide.',
}

export default function LocaleLayout({
  children,
  params,
}: {
  children: React.ReactNode
  params: { locale: string }
}) {
  if (!isLocale(params.locale)) {
    notFound()
  }

  const locale = params.locale as Locale
  const dict = getDictionarySync(locale)

  return (
    <html lang={localeTags[locale]}>
      <body>
        <I18nProvider locale={locale} dict={dict}>
          <CartProvider>
            <SiteHeader locale={locale} />
            <main className="container">{children}</main>
            <footer className="site-footer">
              <div className="container">
                © {new Date().getFullYear()} Stationery. {dict.footer.rights}
              </div>
            </footer>
          </CartProvider>
        </I18nProvider>
      </body>
    </html>
  )
}

function SiteHeader({ locale }: { locale: Locale }) {
  const dict = getDictionarySync(locale)

  return (
    <header className="site-header">
      <div className="container">
        <Link href={`/${locale}`} className="logo">
          Stationery
        </Link>
        <nav className="nav">
          <Link href={`/${locale}/products`}>{dict.nav.shopAll}</Link>
          <Link href={`/${locale}/products?category=journals-notebooks`}>
            {dict.nav.journals}
          </Link>
          <Link href={`/${locale}/products?category=stickers-washi`}>
            {dict.nav.stickers}
          </Link>
          <Link href={`/${locale}/cart`}>{dict.nav.cart}</Link>
          <LanguageSwitcher />
        </nav>
      </div>
    </header>
  )
}
