/**
 * Cloudflare Worker — CORS proxy for the Rackspace Spot API.
 *
 * Routes:
 *   POST /oauth/token          → https://login.spot.rackspace.com/oauth/token
 *   GET  /apis/ngpc.rxt.io/v1/regions        → https://spot.rackspace.com/...
 *   GET  /apis/ngpc.rxt.io/v1/serverclasses  → https://spot.rackspace.com/...
 *
 * Deploy:
 *   npm install -g wrangler
 *   wrangler login
 *   wrangler deploy
 *
 * Then set VITE_SPOT_PROXY_URL (or the UI "API Proxy URL" field) to your
 * worker URL, e.g. https://rackspace-spot-proxy.<your-subdomain>.workers.dev
 */

const CORS_HEADERS = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
  'Access-Control-Allow-Headers': 'Content-Type, Authorization',
  'Access-Control-Max-Age': '86400',
};

export default {
  async fetch(request) {
    if (request.method === 'OPTIONS') {
      return new Response(null, { status: 204, headers: CORS_HEADERS });
    }

    const url = new URL(request.url);
    let target;

    if (url.pathname === '/oauth/token') {
      target = 'https://login.spot.rackspace.com/oauth/token';
    } else if (url.pathname.startsWith('/apis/ngpc.rxt.io/')) {
      target = 'https://spot.rackspace.com' + url.pathname + url.search;
    } else {
      return new Response('Not found', { status: 404, headers: CORS_HEADERS });
    }

    const upstream = new Request(target, {
      method:  request.method,
      headers: request.headers,
      body:    request.method === 'GET' ? undefined : request.body,
    });

    const resp = await fetch(upstream);
    const body = await resp.arrayBuffer();

    const headers = new Headers(resp.headers);
    for (const [k, v] of Object.entries(CORS_HEADERS)) {
      headers.set(k, v);
    }

    return new Response(body, { status: resp.status, headers });
  },
};
