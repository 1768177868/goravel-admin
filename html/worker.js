/**
 * Cloudflare Worker for SPA routing
 * Handles all requests and serves index.html for non-file requests
 */
export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const pathname = url.pathname;

    // First, try to serve the request as-is (for static assets)
    const response = await env.ASSETS.fetch(request);
    
    // If the file exists (status 200), return it
    if (response.status === 200) {
      return response;
    }

    // If the file doesn't exist and it's not index.html, serve index.html for SPA routing
    // This handles all SPA routes (e.g., /admins, /roles, etc.)
    if (pathname !== '/index.html') {
      const indexRequest = new Request(new URL('/index.html', request.url), request);
      return env.ASSETS.fetch(indexRequest);
    }

    // If it's already index.html, return the response (should be 404 if not found)
    return response;
  }
};

