'use client'

import Link from 'next/link'
import { useCart } from '../../../components/CartProvider'
import { useI18n } from '../../../components/I18nProvider'
import { formatPrice } from '../../../lib/api'

export default function CartPage() {
  const { items, remove, updateQty, subtotalCents, count } = useCart()
  const { dict, locale } = useI18n()

  if (items.length === 0) {
    return (
      <div className="empty">
        <p>{dict.cart.empty}</p>
        <Link href={`/${locale}/products`} style={{ color: 'var(--color-accent)' }}>
          {dict.cart.startShopping} →
        </Link>
      </div>
    )
  }

  return (
    <div style={{ padding: '32px 0 64px' }}>
      <h1 style={{ fontSize: 28, margin: '0 0 24px' }}>
        {dict.cart.title} ({count})
      </h1>

      {items.map((item) => (
        <div
          key={item.variantId}
          style={{
            display: 'flex',
            gap: 16,
            padding: '16px 0',
            borderBottom: '1px solid var(--color-border)',
            alignItems: 'center',
          }}
        >
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={item.imageUrl}
            alt={item.title}
            style={{ width: 96, height: 96, objectFit: 'cover', borderRadius: 8 }}
          />
          <div style={{ flex: 1 }}>
            <div style={{ fontWeight: 600 }}>{item.title}</div>
            <div style={{ fontSize: 13, color: 'var(--color-muted)' }}>
              {item.skuCode}
            </div>
          </div>
          <input
            type="number"
            min={1}
            value={item.qty}
            onChange={(e) => updateQty(item.variantId, Number(e.target.value))}
            style={{
              width: 64,
              padding: '6px 8px',
              border: '1px solid var(--color-border)',
              borderRadius: 8,
            }}
          />
          <div style={{ width: 100, textAlign: 'right', fontWeight: 600 }}>
            {formatPrice(item.priceCents * item.qty, 'USD', locale)}
          </div>
          <button
            onClick={() => remove(item.variantId)}
            style={{
              border: 'none',
              background: 'none',
              color: 'var(--color-muted)',
              cursor: 'pointer',
            }}
          >
            {dict.cart.remove}
          </button>
        </div>
      ))}

      <div style={{ marginTop: 24, textAlign: 'right' }}>
        <div style={{ fontSize: 18, marginBottom: 16 }}>
          {dict.cart.subtotal}:{' '}
          <strong>{formatPrice(subtotalCents, 'USD', locale)}</strong>
        </div>
        <Link href={`/${locale}/checkout`}>
          <button className="btn-primary">{dict.cart.checkout}</button>
        </Link>
        <p style={{ fontSize: 12, color: 'var(--color-muted)', marginTop: 8 }}>
          {dict.cart.shippingNote}
        </p>
      </div>
    </div>
  )
}
