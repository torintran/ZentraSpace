import { AsyncLocalStorage } from 'node:async_hooks';
import { randomBytes } from 'node:crypto';
import { createServer } from 'node:http';
import next from 'next';

const port = Number(process.env.PORT || 3000);
const hostname = '0.0.0.0';
const app = next({ dev: false, hostname, port });
const contextualView = new AsyncLocalStorage();
globalThis[Symbol.for('zentraspace.request-view')] = contextualView;
await app.prepare();
const handle = app.getRequestHandler();

async function render(req, res) {
  let size = 0;
  const chunks = [];
  for await (const chunk of req) {
    size += chunk.length;
    if (size > 4 * 1024 * 1024) {
      res.writeHead(413).end();
      return;
    }
    chunks.push(chunk);
  }
  const { path, view } = JSON.parse(Buffer.concat(chunks).toString('utf8'));
  if (
    !view ||
    typeof view.Screen !== 'string' ||
    typeof path !== 'string' ||
    !path.startsWith('/') ||
    path.startsWith('//') ||
    path.includes('\\') ||
    path.startsWith('/_next/') ||
    path.startsWith('/__render')
  ) {
    res.writeHead(400).end();
    return;
  }

  const nonce = randomBytes(18).toString('base64');
  const policy = [
    "default-src 'none'",
    "script-src 'self' 'nonce-" + nonce + "'",
    "style-src 'self' 'unsafe-inline'",
    "img-src 'self'",
    "connect-src 'self'",
    "frame-ancestors 'none'",
    "base-uri 'none'",
    "form-action 'self'",
  ].join('; ');
  req.method = 'GET';
  req.url = path;
  // Do not forward form or client-supplied framework headers to the renderer.
  req.headers = {
    host: 'localhost:' + port,
    'content-security-policy': policy,
  };
  res.setHeader('Content-Security-Policy', policy);
  res.setHeader('Cache-Control', 'private, no-store');
  await contextualView.run({ view }, () => handle(req, res));
}

const server = createServer(async (req, res) => {
  try {
    if (req.method === 'GET' && req.url === '/__ready') {
      res.writeHead(200).end('ok');
    } else if (req.method === 'POST' && req.url === '/__render') {
      await render(req, res);
    } else if (
      (req.method === 'GET' || req.method === 'HEAD') &&
      req.url?.startsWith('/_next/static/')
    ) {
      await handle(req, res);
    } else {
      res.writeHead(404).end();
    }
  } catch {
    if (!res.headersSent) res.writeHead(500);
    res.end();
  }
});
server.requestTimeout = 15000;
server.headersTimeout = 5000;
server.listen(port, hostname);
