// 零依赖 HTML 清洗：商品描述为后台富文本，前端渲染前做二次清洗（纵深防御）。
// 覆盖常见 XSS 向量：脚本/样式、危险标签、事件处理器、javascript: 协议、内联样式。

const htmlComment = /<!--[\s\S]*?-->/g
const scriptStyle = /<(script|style)\b[^>]*>[\s\S]*?<\/\1>/gi
const dangerousTag =
  /<\/?(iframe|object|embed|link|meta|base|form|input|button|textarea|select|option|svg|math|noscript|template|frame|frameset|applet)\b[^>]*>/gi
const eventAttr = /\son[a-z]+\s*=\s*(?:"[^"]*"|'[^']*'|[^\s>]+)/gi
const jsProtocol =
  /(href|src|xlink:href|action|formaction|poster|background)\s*=\s*(["']?)\s*javascript:[^\s"'>]*/gi
const inlineStyle = /\sstyle\s*=\s*(?:"[^"]*"|'[^']*'|[^\s>]+)/gi

export function sanitizeHtml(html: string): string {
  if (!html) return ''
  return html
    .replace(htmlComment, '')
    .replace(scriptStyle, '')
    .replace(dangerousTag, '')
    .replace(eventAttr, '')
    .replace(jsProtocol, '$1=$2#')
    .replace(inlineStyle, '')
}
