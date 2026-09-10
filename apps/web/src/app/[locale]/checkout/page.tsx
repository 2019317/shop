'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useCart } from '../../../components/CartProvider'
import { useI18n } from '../../../components/I18nProvider'
import { formatPrice } from '../../../lib/api'
import { createOrder, type OrderVO } from '../../../lib/order'

export default function CheckoutPage() {
  const { items, subtotalCents, clear } = useCart()
  const { dict, locale } = useI18n()
  const router = useRouter()
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [form, setForm] = useState({
    email: '',
    name: '',
    country: 'US',
    state: '',
    city: '',
    address1: '',
    address2: '',
    postal_code: '',
    phone: '',
    note: '',
  })

  const update = (key: string, value: string) =>
    setForm((prev) => ({ ...prev, [key]: value }))

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (items.length === 0) {
      setError(dict.checkout.emptyCart)
      return
    }
    if (!form.email || !form.name || !form.address1 || !form.city) {
      setError(dict.checkout.fillRequired)
      return
    }

    setSubmitting(true)
    try {
      const order: OrderVO = await createOrder({
        email: form.email,
        currency: 'USD',
        items: items.map((i) => ({ sku_code: i.skuCode, qty: i.qty })),
        shipping_address: {
          name: form.name,
          country: form.country,
          state: form.state,
          city: form.city,
          address1: form.address1,
          address2: form.address2,
          postal_code: form.postal_code,
          phone: form.phone,
        },
        customer_note: form.note,
      })

      clear()
      router.push(`/${locale}/orders/${order.order_no}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Checkout failed')
    } finally {
      setSubmitting(false)
    }
  }

  if (items.length === 0) {
    return (
      <div className="empty">
        <p>{dict.cart.empty}</p>
        <a href={`/${locale}/products`} style={{ color: 'var(--color-accent)' }}>
          {dict.cart.startShopping} →
        </a>
      </div>
    )
  }

  return (
    <div style={{ padding: '32px 0 64px', maxWidth: 960, margin: '0 auto' }}>
      <h1 style={{ fontSize: 28, margin: '0 0 24px' }}>{dict.checkout.title}</h1>

      <div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: 40 }}>
        <form onSubmit={submit}>
          <h3 style={{ marginTop: 0 }}>{dict.checkout.contact}</h3>
          <Field
            label={`${dict.checkout.email} *`}
            value={form.email}
            onChange={(v) => update('email', v)}
            type="email"
          />

          <h3>{dict.checkout.shippingAddress}</h3>
          <Field label={`${dict.checkout.recipient} *`} value={form.name} onChange={(v) => update('name', v)} />
          <Field label={`${dict.checkout.country} *`} value={form.country} onChange={(v) => update('country', v)} />
          <Field label={dict.checkout.state} value={form.state} onChange={(v) => update('state', v)} />
          <Field label={`${dict.checkout.city} *`} value={form.city} onChange={(v) => update('city', v)} />
          <Field label={`${dict.checkout.address} *`} value={form.address1} onChange={(v) => update('address1', v)} />
          <Field label={dict.checkout.apartment} value={form.address2} onChange={(v) => update('address2', v)} />
          <Field label={dict.checkout.postalCode} value={form.postal_code} onChange={(v) => update('postal_code', v)} />
          <Field label={dict.checkout.phone} value={form.phone} onChange={(v) => update('phone', v)} />
          <Field label={dict.checkout.orderNote} value={form.note} onChange={(v) => update('note', v)} />

          {error && <p style={{ color: '#c0392b', marginTop: 12 }}>{error}</p>}

          <button
            className="btn-primary"
            type="submit"
            disabled={submitting}
            style={{ marginTop: 20, width: '100%' }}
          >
            {submitting ? dict.checkout.placing : dict.checkout.placeOrder}
          </button>
        </form>

        <aside>
          <div
            style={{
              border: '1px solid var(--color-border)',
              borderRadius: 12,
              padding: 20,
              background: 'var(--color-surface)',
            }}
          >
            <h3 style={{ marginTop: 0 }}>{dict.checkout.summary}</h3>
            {items.map((i) => (
              <div
                key={i.variantId}
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  padding: '10px 0',
                  borderBottom: '1px solid var(--color-border)',
                  fontSize: 14,
                }}
              >
                <span>
                  {i.title} <span style={{ color: 'var(--color-muted)' }}>× {i.qty}</span>
                </span>
                <span>{formatPrice(i.priceCents * i.qty, 'USD', locale)}</span>
              </div>
            ))}
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                marginTop: 16,
                fontWeight: 700,
                fontSize: 16,
              }}
            >
              <span>{dict.cart.subtotal}</span>
              <span>{formatPrice(subtotalCents, 'USD', locale)}</span>
            </div>
            <p style={{ fontSize: 12, color: 'var(--color-muted)', marginTop: 8 }}>
              {dict.checkout.serverShipping}
            </p>
          </div>
        </aside>
      </div>
    </div>
  )
}

function Field({
  label,
  value,
  onChange,
  type = 'text',
}: {
  label: string
  value: string
  onChange: (v: string) => void
  type?: string
}) {
  return (
    <label style={{ display: 'block', marginBottom: 12 }}>
      <span style={{ fontSize: 13, color: 'var(--color-muted)' }}>{label}</span>
      <input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        style={{
          display: 'block',
          width: '100%',
          marginTop: 4,
          padding: '10px 12px',
          border: '1px solid var(--color-border)',
          borderRadius: 8,
          fontSize: 14,
        }}
      />
    </label>
  )
}
