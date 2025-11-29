export function onRequest(context) {
  const url = new URL(context.request.url)
  
  // 如果是静态资源、API 请求或 index.html 本身，直接返回
  if (url.pathname.startsWith('/assets/') || 
      url.pathname.startsWith('/api/') ||
      url.pathname.startsWith('/ws/') ||
      url.pathname === '/index.html' ||
      url.pathname === '/favicon.ico' ||
      url.pathname.match(/\.(js|css|png|jpg|jpeg|gif|svg|ico|woff|woff2|ttf|eot)$/)) {
    return context.next()
  }
  
  // 其他所有请求都返回 index.html（SPA 路由支持）
  // 使用 fetch 获取 index.html 并返回
  return context.next({
    request: new Request(new URL('/index.html', context.request.url), context.request)
  })
}

