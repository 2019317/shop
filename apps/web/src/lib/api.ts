import type { Locale } from '../i18n/config'

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8888/api/v1'
const IMG_BASE = process.env.NEXT_PUBLIC_IMG_BASE_URL || ''

export function imgUrl(keyOrUrl: string): string {
  if (!keyOrUrl) return ''
  if (keyOrUrl.startsWith('http')) return keyOrUrl
  return `${IMG_BASE}/${keyOrUrl}`
}

interface ApiResult<T> {
  code: number
  msg: string
  data: T
}

async function request<T>(path: string, init?: RequestInit & { revalidate?: number }): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...(init?.headers || {}) },
    // ISR：商品数据变动不频繁，缓存 60 秒
    next: { revalidate: init?.revalidate ?? 60 },
  })
  if (!res.ok) {
    throw new Error(`API error: ${res.status}`)
  }
  const json: ApiResult<T> = await res.json()
  if (json.code !== 0) {
    throw new Error(json.msg)
  }
  return json.data
}

export interface ProductListItem {
  id: string
  title: string
  slug: string
  subtitle: string
  price_cents: number
  currency: string
  cover_url: string
}

export interface Variant {
  id: string
  sku_code: string
  title: string
  options: Record<string, string>
  price_cents: number
  compare_at_cents: number
  weight_g: number
  image_url: string
  in_stock: boolean
}

export interface ProductDetail {
  id: string
  title: string
  slug: string
  subtitle: string
  description: string
  price_cents: number
  currency: string
  category_slug: string
  category_name: string
  attributes: Record<string, unknown>
  tags: string[]
  seo_title: string
  seo_description: string
  images: { url: string; alt: string; sort: number }[]
  variants: Variant[]
}

export interface Category {
  id: string
  name: string
  slug: string
  description: string
  image_url: string
  sort_order: number
}

export interface Paged<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export function formatPrice(
  cents: number,
  currency = 'USD',
  locale: Locale = 'en',
): string {
  return new Intl.NumberFormat(locale === 'zh' ? 'zh-CN' : 'en-US', {
    style: 'currency',
    currency,
  }).format(cents / 100)
}

export const shopApi = {
  products: (params: {
    category?: string
    q?: string
    sort?: string
    page?: number
    pageSize?: number
    locale?: Locale
  }) => {
    const sp = new URLSearchParams()
    if (params.category) sp.set('category', params.category)
    if (params.q) sp.set('q', params.q)
    if (params.sort) sp.set('sort', params.sort)
    if (params.locale) sp.set('locale', params.locale)
    sp.set('page', String(params.page || 1))
    sp.set('page_size', String(params.pageSize || 24))
    return request<Paged<ProductListItem>>(`/products?${sp.toString()}`)
  },

  product: (slug: string, locale?: Locale) =>
    request<ProductDetail>(
      `/products/${slug}${locale ? `?locale=${locale}` : ''}`,
    ),

  categories: (locale?: Locale) =>
    request<Category[]>(`/categories${locale ? `?locale=${locale}` : ''}`),
}
