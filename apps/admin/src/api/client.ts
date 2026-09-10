const API_BASE = import.meta.env.VITE_API_BASE_URL || '/api/v1'

const TOKEN_KEY = 'shop_admin_token'

export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

interface ApiResult<T> {
  code: number
  msg: string
  data: T
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = getToken()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(init.headers as Record<string, string>),
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers })

  if (res.status === 401) {
    clearToken()
    window.location.href = '/login'
    throw new Error('unauthorized')
  }

  const json: ApiResult<T> = await res.json()
  if (json.code !== 0) {
    throw new Error(json.msg || 'request failed')
  }
  return json.data
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'POST', body: JSON.stringify(body) }),
  put: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'PUT', body: JSON.stringify(body) }),
  del: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
}

// ---------- 类型定义（与后端 types 对齐）----------
export interface AdminInfo {
  id: string
  email: string
  name: string
  role: string
}

export interface LoginResult {
  token: string
  expires_in: number
  admin: AdminInfo
}

export interface Category {
  id: string
  name: string
  slug: string
  description: string
  image_url: string
  sort_order: number
  translations?: Record<string, CategoryTranslationInput>
}

export interface VariantInput {
  sku_code: string
  title: string
  options: Record<string, unknown>
  price_cents: number
  compare_at_cents: number
  weight_g: number
  image_key: string
  stock: number
  sort_order: number
}

export interface ImageInput {
  object_key: string
  alt: string
  sort_order: number
}

export interface TranslationInput {
  title?: string
  subtitle?: string
  description?: string
  seo_title?: string
  seo_description?: string
}

// 类目翻译输入（中文覆盖 name/description）
export interface CategoryTranslationInput {
  name?: string
  description?: string
}

export interface ProductInput {
  title: string
  slug: string
  subtitle: string
  description: string
  category_id: string
  status: 'draft' | 'published' | 'archived'
  currency: string
  attributes: Record<string, unknown>
  tags: string[]
  seo_title: string
  seo_description: string
  variants: VariantInput[]
  images: ImageInput[]
  translations?: Record<string, TranslationInput>
}

export interface ProductRow {
  id: string
  title: string
  slug: string
  status: string
  price_cents: number
  currency: string
  updated_at: string
}

export interface Paged<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export interface PresignResult {
  upload_url: string
  object_key: string
  public_url: string
}
