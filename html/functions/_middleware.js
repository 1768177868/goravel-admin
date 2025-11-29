export function onRequest(context) {
  const url = new URL(context.request.url)
  
  // 如果是静态资源或 API 请求，直接返回
  if (url.pathname.startsWith('/assets/') || 
      url.pathname.startsWith('/api/') ||
      url.pathname.startsWith('/ws/') ||
      url.pathname.startsWith('/favicon.ico') ||
      url.pathname.match(/\.(js|css|png|jpg|jpeg|gif|svg|ico|woff|woff2|ttf|eot)$/)) {
    return context.next()
  }
  
  // 其他所有请求都返回 index.html（SPA 路由支持）
  return context.next({
    request: new Request(new URL('/index.html', context.request.url), context.request)
  })
}

