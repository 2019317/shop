// 运行时环境地址解析：
// - 浏览器：必须使用可公开访问的地址（构建期内联的 NEXT_PUBLIC_API_BASE_URL）
// - 服务端渲染：优先使用容器内网地址（API_BASE_URL，如 http://api:8888/api/v1），
//   避免经由公网回环，同时减少对外部网络配置的依赖
export function resolveApiBase(): string {
  const publicBase = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8888/api/v1'
  if (typeof window === 'undefined' && process.env.API_BASE_URL) {
    return process.env.API_BASE_URL
  }
  return publicBase
}
