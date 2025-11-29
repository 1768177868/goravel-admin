export async function onRequest(context) {
  const url = new URL(context.request.url)
  const pathname = url.pathname
  
  // 如果是静态资源、API 请求或 index.html 本身，直接返回
  if (pathname.startsWith('/assets/') || 
      pathname.startsWith('/api/') ||
      pathname.startsWith('/ws/') ||
      pathname === '/index.html' ||
      pathname === '/favicon.ico' ||
      pathname.match(/\.(js|css|png|jpg|jpeg|gif|svg|ico|woff|woff2|ttf|eot)$/)) {
    return context.next()
  }
  
  // 其他所有请求都返回 index.html（SPA 路由支持）
  // 重写请求 URL 为 /index.html
  const indexUrl = new URL('/index.html', url.origin)
  return context.next({
    request: new Request(indexUrl, context.request)
  })
}

