/**
 * Cloudflare Worker for SPA routing
 * Handles all requests and serves index.html for non-file requests
 * This ensures that routes like /operation-logs work correctly when refreshed
 */
export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const pathname = url.pathname;
    
    // Direct requests to index.html or root should be served as-is
    if (pathname === '/index.html' || pathname === '/') {
      const indexRequest = new Request(new URL('/index.html', url.origin).toString(), {
        method: request.method,
        headers: request.headers,
      });
      return env.ASSETS.fetch(indexRequest);
    }
    
    // Check if the request is for a static asset (has file extension)
    const hasFileExtension = /\.\w+$/.test(pathname);
    
    // If it's a static asset request, try to fetch it
    if (hasFileExtension) {
      const asset = await env.ASSETS.fetch(request);
      // If asset exists, return it; otherwise continue to serve index.html
      if (asset.status === 200) {
        return asset;
      }
    }
    
    // For non-file requests (SPA routes), serve index.html
    // This handles all SPA routes like /operation-logs, /login, /admins, etc.
    const indexRequest = new Request(new URL('/index.html', url.origin).toString(), {
      method: request.method,
      headers: request.headers,
    });
    
    return env.ASSETS.fetch(indexRequest);
  }
};

