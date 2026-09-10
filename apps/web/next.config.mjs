/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // standalone 输出：仅打包运行所需文件，便于用精简镜像部署（Docker）
  output: 'standalone',
  // 图片全部托管在 Cloudflare R2，通过自定义域名访问
  images: {
    remotePatterns: [
      { protocol: 'https', hostname: 'img.yourbrand.com' },
      { protocol: 'https', hostname: '*.r2.cloudflarestorage.com' },
    ],
    formats: ['image/avif', 'image/webp'],
  },
  // 面向海外，SEO 优先：商品详情用 ISR
  experimental: {
    optimizePackageImports: [],
  },
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          { key: 'X-Content-Type-Options', value: 'nosniff' },
          { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
        ],
      },
    ]
  },
}

export default nextConfig
