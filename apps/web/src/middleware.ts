import { NextResponse, type NextRequest } from 'next/server'
import { defaultLocale, isLocale, locales } from './i18n/config'

// 从请求中推断用户偏好语言
// 优先级：URL 前缀 > Cookie > Accept-Language > 默认
function detectLocale(request: NextRequest): string {
  const cookieLocale = request.cookies.get('NEXT_LOCALE')?.value
  if (cookieLocale && isLocale(cookieLocale)) {
    return cookieLocale
  }

  const acceptLanguage = request.headers.get('accept-language')
  if (acceptLanguage) {
    for (const part of acceptLanguage.split(',')) {
      const tag = part.split(';')[0].trim().toLowerCase()
      if (tag.startsWith('zh')) return 'zh'
      if (tag.startsWith('en')) return 'en'
    }
  }
  return defaultLocale
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // 已带语言前缀则放行
  const hasLocale = locales.some(
    (locale) => pathname === `/${locale}` || pathname.startsWith(`/${locale}/`),
  )
  if (hasLocale) {
    return NextResponse.next()
  }

  // 跳过静态资源与接口路径
  if (
    pathname.startsWith('/_next') ||
    pathname.startsWith('/api') ||
    pathname.includes('.')
  ) {
    return NextResponse.next()
  }

  const locale = detectLocale(request)
  const url = request.nextUrl.clone()
  url.pathname = `/${locale}${pathname === '/' ? '' : pathname}`

  const response = NextResponse.redirect(url)
  response.cookies.set('NEXT_LOCALE', locale, {
    maxAge: 60 * 60 * 24 * 365,
    path: '/',
  })
  return response
}

export const config = {
  matcher: ['/((?!_next|api|.*\\..*).*)'],
}
