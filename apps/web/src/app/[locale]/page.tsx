import Link from 'next/link'
import ProductCard from '../../components/ProductCard'
import { shopApi } from '../../lib/api'
import { getDictionarySync } from '../../i18n/dictionaries'
import { isLocale, type Locale } from '../../i18n/config'

export const revalidate = 60

export default async function HomePage({
  params,
}: {
  params: { locale: string }
}) {
  const locale = (isLocale(params.locale) ? params.locale : 'en') as Locale
  const dict = getDictionarySync(locale)

  const [productsRes, categories] = await Promise.all([
    shopApi.products({ page: 1, pageSize: 8, sort: 'newest', locale }),
    shopApi.categories(locale),
  ])

  return (
    <>
      <section className="hero">
        <h1>{dict.home.title}</h1>
        <p>{dict.home.subtitle}</p>
      </section>

      <section>
        <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', marginBottom: 8 }}>
          {categories.map((c) => (
            <Link
              key={c.id}
              href={`/${locale}/products?category=${c.slug}`}
              style={{
                padding: '6px 14px',
                border: '1px solid var(--color-border)',
                borderRadius: 999,
                fontSize: 14,
                background: 'var(--color-surface)',
              }}
            >
              {c.name}
            </Link>
          ))}
        </div>
      </section>

      <section className="grid">
        {productsRes.list.map((p) => (
          <ProductCard key={p.id} product={p} locale={locale} />
        ))}
      </section>

      <div style={{ textAlign: 'center', paddingBottom: 64 }}>
        <Link href={`/${locale}/products`}>{dict.home.viewAll} →</Link>
      </div>
    </>
  )
}
