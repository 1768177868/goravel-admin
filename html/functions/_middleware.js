/**
 * Cloudflare Pages Functions Middleware
 * Handles SPA routing by redirecting all non-file requests to index.html
 * This ensures that routes like /operation-logs work correctly when refreshed
 */
export async function onRequest(context) {
  const { request, next } = context;
  const url = new URL(request.url);
  
  // Check if the request is for a static asset (has file extension)
  const pathname = url.pathname;
  const hasFileExtension = /\.\w+$/.test(pathname);
  
  // If it's not a static asset and not already index.html, rewrite to index.html
  if (!hasFileExtension && pathname !== '/index.html' && pathname !== '/') {
    // Rewrite the request to serve index.html for SPA routing
    const newUrl = new URL('/index.html', url.origin);
    return next(new Request(newUrl, request));
  }
  
  // For static assets or index.html, proceed normally
  return next();
}

