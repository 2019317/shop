import { orderApi } from '../../../../lib/order'
import { formatPrice } from '../../../../lib/api'
import { getDictionarySync } from '../../../../i18n/dictionaries'
import { isLocale, type Locale } from '../../../../i18n/config'

export const dynamic = 'force-dynamic'

const statusKey = {
  pending: 'pending',
  paid: 'paid',
  fulfilled: 'fulfilled',
  cancelled: 'cancelled',
  refunded: 'refunded',
} as const

export default async function OrderPage({
  params,
  searchParams,
}: {
  params: { orderNo: string; locale: string }
  searchParams: { email?: string }
}) {
  const locale = (isLocale(params.locale) ? params.locale : 'en') as Locale
  const dict = getDictionarySync(locale)

  let order
  try {
    // 需同时提供下单邮箱，服务端校验归属后再返回订单详情
    order = await orderApi.detail(params.orderNo, searchParams.email)
  } catch {
    order = null
  }

  if (!order) {
    return <div className="empty">{dict.order.notFound}</div>
  }

  const status =
    dict.order.status[statusKey[order.status as keyof typeof statusKey] ?? 'pending']

  return (
    <div style={{ padding: '32px 0 64px', maxWidth: 800, margin: '0 auto' }}>
      <p style={{ color: 'var(--color-muted)', margin: 0 }}>{dict.order.title}</p>
      <h1 style={{ fontSize: 26, margin: '4px 0 4px' }}>{order.order_no}</h1>
      <p style={{ color: 'var(--color-accent)', fontWeight: 600 }}>{status}</p>

      <div
        style={{
          border: '1px solid var(--color-border)',
          borderRadius: 12,
          padding: 20,
          marginTop: 24,
          background: 'var(--color-surface)',
        }}
      >
        <h3 style={{ marginTop: 0 }}>{dict.order.items}</h3>
        {order.items.map((it) => (
          <div
            key={it.sku_code}
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              padding: '10px 0',
              borderBottom: '1px solid var(--color-border)',
              fontSize: 14,
            }}
          >
            <span>
              {it.title} <span style={{ color: 'var(--color-muted)' }}>× {it.qty}</span>
            </span>
            <span>{formatPrice(it.total_cents, order.currency, locale)}</span>
          </div>
        ))}

        <Row label={dict.order.subtotal} value={formatPrice(order.subtotal_cents, order.currency, locale)} />
        <Row label={dict.order.shipping} value={formatPrice(order.shipping_cents, order.currency, locale)} />
        {order.discount_cents > 0 && (
          <Row
            label={dict.order.discount}
            value={`-${formatPrice(order.discount_cents, order.currency, locale)}`}
          />
        )}
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            marginTop: 12,
            fontWeight: 700,
            fontSize: 16,
          }}
        >
          <span>{dict.order.total}</span>
          <span>{formatPrice(order.total_cents, order.currency, locale)}</span>
        </div>
      </div>

      <div
        style={{
          border: '1px solid var(--color-border)',
          borderRadius: 12,
          padding: 20,
          marginTop: 16,
          background: 'var(--color-surface)',
        }}
      >
        <h3 style={{ marginTop: 0 }}>{dict.order.shipTo}</h3>
        <p style={{ margin: 0, lineHeight: 1.7, fontSize: 14 }}>
          {order.shipping_address?.name}
          <br />
          {order.shipping_address?.address1} {order.shipping_address?.address2}
          <br />
          {order.shipping_address?.city} {order.shipping_address?.state}{' '}
          {order.shipping_address?.postal_code}
          <br />
          {order.shipping_address?.country}
          {order.shipping_address?.phone && (
            <>
              <br />
              {order.shipping_address.phone}
            </>
          )}
        </p>
      </div>

      {order.status === 'fulfilled' && order.tracking_no && (
        <div
          style={{
            border: '1px solid var(--color-border)',
            borderRadius: 12,
            padding: 20,
            marginTop: 16,
            background: 'var(--color-surface)',
          }}
        >
          <h3 style={{ marginTop: 0 }}>
            {dict.order.tracking}
            {order.shipment_status && (
              <span
                style={{
                  marginLeft: 8,
                  fontSize: 12,
                  fontWeight: 500,
                  padding: '2px 8px',
                  borderRadius: 10,
                  background: 'var(--color-accent)',
                  color: '#fff',
                }}
              >
                {order.shipment_status === 'delivered'
                  ? dict.order.shipStatusDelivered
                  : dict.order.shipStatusShipped}
              </span>
            )}
          </h3>
          <p style={{ margin: 0, fontSize: 14 }}>
            {order.carrier} — {order.tracking_no}
            {order.tracking_url && (
              <>
                {' · '}
                <a
                  href={order.tracking_url}
                  target="_blank"
                  rel="noreferrer"
                  style={{ color: 'var(--color-accent)' }}
                >
                  {dict.order.trackPackage}
                </a>
              </>
            )}
          </p>
        </div>
      )}

      <p style={{ fontSize: 12, color: 'var(--color-muted)', marginTop: 24 }}>
        {dict.order.placedAt} {order.created_at}. {dict.order.confirmEmail}
      </p>
    </div>
  )
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'space-between',
        marginTop: 8,
        fontSize: 14,
        color: 'var(--color-muted)',
      }}
    >
      <span>{label}</span>
      <span>{value}</span>
    </div>
  )
}
