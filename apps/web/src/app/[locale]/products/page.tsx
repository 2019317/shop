import ProductCard from '../../../components/ProductCard'
import { shopApi } from '../../../lib/api'
import { getDictionarySync } from '../../../i18n/dictionaries'
import { isLocale, type Locale } from '../../../i18n/config'

export const revalidate = 60

export default async function ProductsPage({
  params,
  searchParams,
}: {
  params: { locale: string }
  searchParams: { category?: string; q?: string; sort?: string; page?: string }
}) {
  const locale = (isLocale(params.locale) ? params.locale : 'en') as Locale
  const dict = getDictionarySync(locale)

  const page = Number(searchParams.page || 1)
  const { list, total } = await shopApi.products({
    category: searchParams.category,
    q: searchParams.q,
    sort: searchParams.sort || 'newest',
    page,
    pageSize: 24,
    locale,
  })

  const buildHref = (p: number) => {
    const sp = new URLSearchParams()
    if (searchParams.category) sp.set('category', searchParams.category)
    if (searchParams.q) sp.set('q', searchParams.q)
    if (searchParams.sort) sp.set('sort', searchParams.sort)
    sp.set('page', String(p))
    return `/${locale}/products?${sp.toString()}`
  }

  const totalPages = Math.max(1, Math.ceil(total / 24))

  return (
    <>
      <div style={{ padding: '32px 0 8px' }}>
        <h1 style={{ fontSize: 28, margin: '0 0 6px' }}>
          {searchParams.q
            ? `${dict.products.searchPrefix}: ${searchParams.q}`
            : dict.products.shopAll}
        </h1>
        <p style={{ color: 'var(--color-muted)', margin: 0 }}>
          {total} {dict.products.count}
        </p>
      </div>

      {list.length === 0 ? (
        <div className="empty">{dict.products.empty}</div>
      ) : (
        <section className="grid">
          {list.map((p) => (
            <ProductCard key={p.id} product={p} locale={locale} />
          ))}
        </section>
      )}

      {totalPages > 1 && (
        <div
          style={{
            display: 'flex',
            gap: 8,
            justifyContent: 'center',
            paddingBottom: 64,
          }}
        >
          {Array.from({ length: totalPages }).map((_, i) => (
            <a
              key={i}
              href={buildHref(i + 1)}
              style={{
                padding: '8px 14px',
                border: '1px solid var(--color-border)',
                borderRadius: 8,
                background: page === i + 1 ? 'var(--color-accent)' : '#fff',
                color: page === i + 1 ? '#fff' : 'inherit',
              }}
            >
              {i + 1}
            </a>
          ))}
        </div>
      )}
    </>
  )
}
