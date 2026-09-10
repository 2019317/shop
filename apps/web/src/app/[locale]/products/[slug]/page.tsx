import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import ProductGallery from '../../../../components/ProductGallery'
import ProductPurchase from '../../../../components/ProductPurchase'
import { shopApi, type ProductDetail } from '../../../../lib/api'
import { sanitizeHtml } from '../../../../lib/sanitize'
import { getDictionarySync } from '../../../../i18n/dictionaries'
import { isLocale, locales, type Locale } from '../../../../i18n/config'

export const revalidate = 60

async function getProduct(
  slug: string,
  locale: Locale,
): Promise<ProductDetail | null> {
  try {
    return await shopApi.product(slug, locale)
  } catch {
    return null
  }
}

// 面向海外：为每个语言生成对应的 SEO 元数据
export async function generateMetadata({
  params,
}: {
  params: { slug: string; locale: string }
}): Promise<Metadata> {
  const locale = (isLocale(params.locale) ? params.locale : 'en') as Locale
  const product = await getProduct(params.slug, locale)
  if (!product) return { title: 'Product not found' }

  return {
    title: product.seo_title || `${product.title} — Stationery`,
    description: product.seo_description || product.subtitle || product.title,
    openGraph: {
      title: product.title,
      description: product.subtitle,
      images: product.images?.[0]?.url ? [product.images[0].url] : [],
    },
    // 告知搜索引擎存在其他语言版本，避免重复内容
    alternates: {
      languages: Object.fromEntries(
        locales.map((l) => [l, `/${l}/products/${params.slug}`]),
      ),
    },
  }
}

export default async function ProductDetailPage({
  params,
}: {
  params: { slug: string; locale: string }
}) {
  const locale = (isLocale(params.locale) ? params.locale : 'en') as Locale
  const dict = getDictionarySync(locale)

  const product = await getProduct(params.slug, locale)
  if (!product) notFound()

  return (
    <div className="product-detail">
      <ProductGallery images={product.images || []} title={product.title} />

      <div>
        <p style={{ color: 'var(--color-muted)', fontSize: 13, margin: 0 }}>
          {product.category_name}
        </p>
        <h1 style={{ fontSize: 28, margin: '6px 0' }}>{product.title}</h1>
        {product.subtitle && (
          <p style={{ color: 'var(--color-muted)', margin: '0 0 16px' }}>
            {product.subtitle}
          </p>
        )}

        <ProductPurchase product={product} locale={locale} />

        {product.description && (
          <div
            style={{ marginTop: 32, lineHeight: 1.7 }}
            dangerouslySetInnerHTML={{ __html: sanitizeHtml(product.description) }}
          />
        )}

        {!!product.tags?.length && (
          <div style={{ marginTop: 24, display: 'flex', gap: 8, flexWrap: 'wrap' }}>
            {product.tags.map((t) => (
              <span
                key={t}
                style={{
                  fontSize: 12,
                  padding: '4px 10px',
                  border: '1px solid var(--color-border)',
                  borderRadius: 999,
                  color: 'var(--color-muted)',
                }}
              >
                {t}
              </span>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
