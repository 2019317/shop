'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useCart } from './CartProvider'
import { useI18n } from './I18nProvider'
import { formatPrice, type ProductDetail } from '../lib/api'

export default function ProductPurchase({
  product,
  locale,
}: {
  product: ProductDetail
  locale: string
}) {
  const [selected, setSelected] = useState(0)
  const [qty, setQty] = useState(1)
  const { add } = useCart()
  const { dict } = useI18n()
  const router = useRouter()

  const variants = product.variants || []
  const variant = variants[selected]

  const handleAddToCart = () => {
    if (!variant) return
    add(
      {
        variantId: variant.id,
        skuCode: variant.sku_code,
        title: `${product.title}${variant.title ? ` — ${variant.title}` : ''}`,
        options: variant.options || {},
        priceCents: variant.price_cents,
        imageUrl: variant.image_url || product.images?.[0]?.url || '',
      },
      qty,
    )
    router.push(`/${locale}/cart`)
  }

  return (
    <div>
      <div style={{ fontSize: 24, fontWeight: 700, color: 'var(--color-accent)' }}>
        {formatPrice(
          variant?.price_cents ?? product.price_cents,
          product.currency,
          locale === 'zh' ? 'zh' : 'en',
        )}
        {variant?.compare_at_cents ? (
          <span
            style={{
              marginLeft: 10,
              fontSize: 15,
              color: 'var(--color-muted)',
              textDecoration: 'line-through',
              fontWeight: 400,
            }}
          >
            {formatPrice(variant.compare_at_cents, product.currency,
              locale === 'zh' ? 'zh' : 'en')}
          </span>
        ) : null}
      </div>

      {variants.length > 0 && (
        <div className="variant-list">
          {variants.map((v, i) => (
            <div
              key={v.id}
              className={`variant-item ${i === selected ? 'selected' : ''}`}
              onClick={() => setSelected(i)}
            >
              <div style={{ fontWeight: 600, fontSize: 14 }}>
                {v.title || v.sku_code}
              </div>
              <div style={{ fontSize: 13, color: 'var(--color-muted)' }}>
                {formatPrice(v.price_cents, product.currency, locale === 'zh' ? 'zh' : 'en')}
                {!v.in_stock && (
                  <span style={{ marginLeft: 8, color: '#c0392b' }}>
                    {dict.product.outOfStock}
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: 16 }}>
        <label style={{ fontSize: 14, color: 'var(--color-muted)' }}>
          {dict.product.qty}
        </label>
        <input
          type="number"
          min={1}
          max={99}
          value={qty}
          onChange={(e) => setQty(Math.max(1, Number(e.target.value) || 1))}
          style={{
            width: 72,
            padding: '8px 10px',
            border: '1px solid var(--color-border)',
            borderRadius: 8,
          }}
        />
      </div>

      <button
        className="btn-primary"
        onClick={handleAddToCart}
        disabled={!variant || !variant.in_stock}
      >
        {variant && !variant.in_stock
          ? dict.product.outOfStock
          : dict.product.addToCart}
      </button>
    </div>
  )
}
