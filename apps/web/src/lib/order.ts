const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8888/api/v1'

interface ApiResult<T> {
  code: number
  msg: string
  data: T
}

async function post<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
    cache: 'no-store',
  })
  const json: ApiResult<T> = await res.json()
  if (json.code !== 0) {
    throw new Error(json.msg || 'request failed')
  }
  return json.data
}

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, { cache: 'no-store' })
  const json: ApiResult<T> = await res.json()
  if (json.code !== 0) {
    throw new Error(json.msg || 'request failed')
  }
  return json.data
}

export interface OrderItemVO {
  sku_code: string
  title: string
  options: Record<string, string>
  image_url: string
  unit_price_cents: number
  qty: number
  total_cents: number
}

export interface OrderVO {
  order_no: string
  email: string
  status: string
  currency: string
  subtotal_cents: number
  shipping_cents: number
  discount_cents: number
  total_cents: number
  customer_note: string
  shipping_address: Record<string, string>
  items: OrderItemVO[]
  payment_status?: string
  provider?: string
  carrier?: string
  tracking_no?: string
  tracking_url?: string
  created_at: string
}

export interface CreateOrderInput {
  email: string
  currency: string
  items: { sku_code: string; qty: number }[]
  shipping_address: Record<string, string>
  customer_note?: string
  coupon_code?: string
}

export interface ValidateCouponResult {
  valid: boolean
  code: string
  discount_cents: number
  message: string
}

export interface PublicCoupon {
  code: string
  type: 'percent' | 'fixed'
  value: number
  min_amount_cents: number
  ends_at?: string
}

export const orderApi = {
  create: (input: CreateOrderInput) => post<OrderVO>('/orders', input),
  // 试算优惠券：服务端校验，返回可抵扣金额（不占用名额）
  validateCoupon: (code: string, subtotalCents: number) =>
    post<ValidateCouponResult>('/coupons/validate', {
      code,
      subtotal_cents: subtotalCents,
    }),
  // 可用券列表
  activeCoupons: () => get<PublicCoupon[]>('/coupons'),
  // 查询订单需同时提供下单邮箱，服务端据此校验归属，防止订单号被枚举泄露隐私
  detail: (orderNo: string, email?: string) =>
    get<OrderVO>(
      `/orders/${orderNo}${email ? `?email=${encodeURIComponent(email)}` : ''}`,
    ),
  // 仅供本地/测试环境的 mock 渠道手动标记支付；生产环境由支付服务商 Webhook 回调
  notifyPaid: (orderNo: string, amountCents: number) =>
    post<{ handled: boolean }>('/payments/notify', {
      provider: 'mock',
      event_id: `web_${orderNo}_${Date.now()}`,
      order_no: orderNo,
      status: 'succeeded',
      amount_cents: amountCents,
    }),
}

export const createOrder = orderApi.create
