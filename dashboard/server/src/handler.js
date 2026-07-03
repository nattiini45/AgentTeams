'use strict';

const { classify } = require('./allowlist');

/**
 * createRequestHandler builds a Node http(s) request listener.
 *
 * All upstream I/O is via the injected `controllerClient` / `minioClient`
 * so this module (and its tests) never make a real network call.
 *
 * @param {Object} deps
 * @param {import('./controller-client').ControllerClient} deps.controllerClient
 * @param {import('./minio-client').MinioClient} deps.minioClient
 * @param {(entry: object) => void} deps.logWrite  called once per allow-listed write
 * @param {Buffer} [deps.spaIndex]   optional static SPA index.html to serve for non-/api paths
 * @param {(pathname: string) => {contentType: string, body: Buffer} | null} [deps.staticFile]
 *        optional resolver for serving the built SPA's static assets
 */
function createRequestHandler(deps) {
  const { controllerClient, minioClient, logWrite, staticFile } = deps;

  return async function handleRequest(req, res) {
    let url;
    try {
      url = new URL(req.url, 'http://internal');
    } catch {
      sendJson(res, 400, { error: 'bad request' });
      return;
    }

    if (!url.pathname.startsWith('/api/')) {
      if (staticFile) {
        const asset = staticFile(url.pathname);
        if (asset) {
          res.writeHead(200, { 'content-type': asset.contentType });
          res.end(asset.body);
          return;
        }
      }
      sendJson(res, 404, { error: 'not found' });
      return;
    }

    const method = (req.method || 'GET').toUpperCase();
    const decision = classify(method, url.pathname);

    if (!decision.ok) {
      sendJson(res, decision.status, { error: decision.status === 405 ? 'method not allowed' : 'not found' });
      return;
    }

    const route = decision.route;

    try {
      if (route.target === 'controller') {
        let body;
        if (route.kind === 'write') {
          body = await readBody(req);
        }
        const controllerPath = route.controllerPath + (url.search || '');
        const upstream = await controllerClient.request(method, controllerPath, body);

        if (route.kind === 'write') {
          logWrite({
            action: route.action,
            worker: route.workerName,
            status: upstream.statusCode,
            remoteAddr: req.socket ? req.socket.remoteAddress : undefined,
          });
        }

        relayJson(res, upstream);
        return;
      }

      if (route.target === 'minio') {
        await handleMinio(res, minioClient, route.minioKey, url.pathname);
        return;
      }

      sendJson(res, 500, { error: 'unhandled route target' });
    } catch (err) {
      sendJson(res, 502, { error: 'upstream request failed', detail: String((err && err.message) || err) });
    }
  };
}

async function handleMinio(res, minioClient, key, requestedPath) {
  // /api/files/<root>/ (trailing slash implied by no extension AND the
  // caller wanting a directory listing) is ambiguous from the path alone,
  // so we try GetObject first (covers meta.json/plan.md/state files) and
  // fall back to a prefix listing on 404 -- this covers both the file
  // browser (directories) and direct file reads (objects) through one route.
  const getRes = await minioClient.getObject(key);
  if (getRes.statusCode === 200) {
    res.writeHead(200, { 'content-type': guessContentType(key) });
    res.end(getRes.body);
    return;
  }
  if (getRes.statusCode !== 404) {
    sendJson(res, 502, { error: 'minio object read failed', status: getRes.statusCode });
    return;
  }

  // Fall back to a listing under key + '/'.
  const prefix = key.endsWith('/') ? key : key + '/';
  try {
    const listing = await minioClient.listObjects(prefix);
    if (listing.prefixes.length === 0 && listing.objects.length === 0) {
      sendJson(res, 404, { error: 'not found', path: requestedPath });
      return;
    }
    sendJson(res, 200, {
      prefix,
      directories: listing.prefixes.map((p) => p.slice(prefix.length).replace(/\/$/, '')),
      files: listing.objects
        .filter((o) => o.key !== prefix)
        .map((o) => ({ key: o.key.slice(prefix.length), size: o.size, lastModified: o.lastModified })),
    });
  } catch (err) {
    sendJson(res, 502, { error: 'minio list failed', detail: String((err && err.message) || err) });
  }
}

function guessContentType(key) {
  if (key.endsWith('.json')) return 'application/json';
  if (key.endsWith('.md')) return 'text/markdown; charset=utf-8';
  return 'application/octet-stream';
}

function readBody(req) {
  return new Promise((resolve, reject) => {
    const chunks = [];
    req.on('data', (c) => chunks.push(c));
    req.on('end', () => resolve(Buffer.concat(chunks)));
    req.on('error', reject);
  });
}

function relayJson(res, upstream) {
  const headers = { ...upstream.headers };
  if (!headers['content-type']) headers['content-type'] = 'application/json';
  res.writeHead(upstream.statusCode, headers);
  res.end(upstream.body);
}

function sendJson(res, status, obj) {
  const body = Buffer.from(JSON.stringify(obj));
  res.writeHead(status, { 'content-type': 'application/json', 'content-length': String(body.length) });
  res.end(body);
}

module.exports = { createRequestHandler };
