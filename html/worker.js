/**
 * Cloudflare Worker for SPA routing
 * Handles all requests and serves index.html for non-file requests
 * Based on Cloudflare's recommended pattern for SPA routing
 */
export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    
    // First, try to fetch the requested asset
    const asset = await env.ASSETS.fetch(request);
    
    // If the asset exists (status 200), return it
    if (asset.status === 200) {
      return asset;
    }
    
    // If the asset doesn't exist (404), serve index.html for SPA routing
    // This handles all SPA routes like /login, /admins, etc.
    if (asset.status === 404) {
      return env.ASSETS.fetch(new Request(url.origin + '/index.html', request));
    }
    
    // For any other status, return the asset response
    return asset;
  }
};

