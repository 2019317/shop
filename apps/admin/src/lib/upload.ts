import { api, type PresignResult } from '../api/client'

const ALLOWED = ['image/jpeg', 'image/png', 'image/webp', 'image/avif']
const MAX_SIZE = 5 * 1024 * 1024

export interface UploadedAsset {
  objectKey: string
  publicUrl: string
}

/**
 * 图片直传 Cloudflare R2：
 * 1) 向后端申请预签名地址
 * 2) 浏览器直接 PUT 到 R2（不经过服务器，2GB 机型无压力）
 */
export async function uploadImage(file: File): Promise<UploadedAsset> {
  if (!ALLOWED.includes(file.type)) {
    throw new Error('Only JPG / PNG / WEBP / AVIF images are allowed')
  }
  if (file.size > MAX_SIZE) {
    throw new Error('Image must be smaller than 5MB')
  }

  const presign = await api.post<PresignResult>('/admin/assets/presign', {
    filename: file.name,
    content_type: file.type,
    size: file.size,
  })

  const res = await fetch(presign.upload_url, {
    method: 'PUT',
    headers: { 'Content-Type': file.type },
    body: file,
  })

  if (!res.ok) {
    throw new Error(`Upload failed: ${res.status}`)
  }

  return { objectKey: presign.object_key, publicUrl: presign.public_url }
}
