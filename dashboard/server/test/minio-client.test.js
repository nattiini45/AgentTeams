'use strict';

const test = require('node:test');
const assert = require('node:assert/strict');
const http = require('node:http');
const { MinioClient } = require('../src/minio-client');
const { signRequest } = require('../src/sigv4');

function startUpstream(handler) {
  const server = http.createServer(handler);
  return new Promise((resolve) => server.listen(0, '127.0.0.1', () => resolve(server)));
}

test('signRequest produces a well-formed Authorization header', () => {
  const { headers } = signRequest({
    method: 'GET',
    host: '127.0.0.1:9000',
    path: '/mybucket/shared/tasks/t1/meta.json',
    query: {},
    accessKey: 'AKIAEXAMPLE',
    secretKey: 'secretkey',
    date: new Date('2026-07-03T00:00:00Z'),
  });
  assert.match(headers.authorization, /^AWS4-HMAC-SHA256 Credential=AKIAEXAMPLE\//);
  assert.match(headers.authorization, /SignedHeaders=host;x-amz-content-sha256;x-amz-date/);
  assert.equal(headers['x-amz-date'], '20260703T000000Z');
});

test('MinioClient.getObject signs the request and returns the upstream body', async () => {
  let seenAuth;
  let seenPath;
  const upstream = await startUpstream((req, res) => {
    seenAuth = req.headers.authorization;
    seenPath = req.url;
    if (req.url === '/hiclaw-storage/shared/tasks/t1/meta.json') {
      res.writeHead(200, { 'content-type': 'application/json' });
      res.end('{"task_id":"t1"}');
    } else {
      res.writeHead(404);
      res.end();
    }
  });
  try {
    const { port } = upstream.address();
    const client = new MinioClient({
      endpoint: `http://127.0.0.1:${port}`,
      accessKey: 'ak',
      secretKey: 'sk',
      bucket: 'hiclaw-storage',
    });
    const res = await client.getObject('shared/tasks/t1/meta.json');
    assert.equal(res.statusCode, 200);
    assert.equal(res.body.toString('utf8'), '{"task_id":"t1"}');
    assert.ok(seenAuth.startsWith('AWS4-HMAC-SHA256'));
    assert.equal(seenPath, '/hiclaw-storage/shared/tasks/t1/meta.json');
  } finally {
    upstream.close();
  }
});

test('MinioClient.listObjects parses ListObjectsV2 XML into prefixes/objects', async () => {
  const xml = `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult>
  <Name>hiclaw-storage</Name>
  <Prefix>shared/projects/</Prefix>
  <CommonPrefixes><Prefix>shared/projects/proj-1/</Prefix></CommonPrefixes>
  <Contents>
    <Key>shared/projects/readme.txt</Key>
    <Size>42</Size>
    <LastModified>2026-01-01T00:00:00.000Z</LastModified>
  </Contents>
</ListBucketResult>`;
  const upstream = await startUpstream((req, res) => {
    res.writeHead(200, { 'content-type': 'application/xml' });
    res.end(xml);
  });
  try {
    const { port } = upstream.address();
    const client = new MinioClient({
      endpoint: `http://127.0.0.1:${port}`,
      accessKey: 'ak',
      secretKey: 'sk',
      bucket: 'hiclaw-storage',
    });
    const listing = await client.listObjects('shared/projects/');
    assert.deepEqual(listing.prefixes, ['shared/projects/proj-1/']);
    assert.equal(listing.objects.length, 1);
    assert.equal(listing.objects[0].key, 'shared/projects/readme.txt');
    assert.equal(listing.objects[0].size, 42);
  } finally {
    upstream.close();
  }
});

test('MinioClient.listObjects throws on an error status', async () => {
  const upstream = await startUpstream((req, res) => {
    res.writeHead(403, { 'content-type': 'application/xml' });
    res.end('<Error><Code>AccessDenied</Code></Error>');
  });
  try {
    const { port } = upstream.address();
    const client = new MinioClient({
      endpoint: `http://127.0.0.1:${port}`,
      accessKey: 'ak',
      secretKey: 'sk',
      bucket: 'hiclaw-storage',
    });
    await assert.rejects(() => client.listObjects('shared/'));
  } finally {
    upstream.close();
  }
});
