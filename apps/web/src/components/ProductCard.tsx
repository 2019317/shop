import Link from 'next/link'
import { formatPrice, imgUrl, type ProductListItem } from '../lib/api'
import type { Locale } from '../i18n/config'

export default function ProductCard({
  product,
  locale,
}: {
  product: ProductListItem
  locale: Locale
}) {
  return (
    <Link href={`/${locale}/products/${product.slug}`} className="card">
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        className="card-image"
        src={imgUrl(product.cover_url)}
        alt={product.title}
      />
      <div className="card-body">
        <h3 className="card-title">{product.title}</h3>
        <p className="card-subtitle">{product.subtitle}</p>
        <div className="price">
          {formatPrice(product.price_cents, product.currency, locale)}
        </div>
      </div>
    </Link>
  )
}
