'use strict';

const test = require('node:test');
const assert = require('node:assert/strict');
const { classify } = require('../src/allowlist');

test('GET list endpoints map to controller v1 routes', () => {
  for (const kind of ['managers', 'teams', 'workers', 'projects']) {
    const d = classify('GET', `/api/${kind}`);
    assert.equal(d.ok, true);
    assert.equal(d.route.target, 'controller');
    assert.equal(d.route.kind, 'get');
    assert.equal(d.route.controllerPath, `/api/v1/${kind}`);
  }
  const mt = classify('GET', '/api/manager-tasks');
  assert.equal(mt.ok, true);
  assert.equal(mt.route.controllerPath, '/api/v1/manager-tasks');
});

test('GET single-resource paths pass through with name segment', () => {
  const d = classify('GET', '/api/workers/alice');
  assert.equal(d.ok, true);
  assert.equal(d.route.controllerPath, '/api/v1/workers/alice');
});

test('unknown top-level path is 404', () => {
  const d = classify('GET', '/api/nonsense');
  assert.equal(d.ok, false);
  assert.equal(d.status, 404);
});

test('non-GET on a list endpoint is 405', () => {
  for (const method of ['POST', 'PUT', 'DELETE', 'PATCH']) {
    const d = classify(method, '/api/workers');
    assert.equal(d.ok, false, `${method} should be rejected`);
    assert.equal(d.status, 405);
  }
});

test('exactly the three lifecycle POSTs are allowed as writes', () => {
  for (const action of ['wake', 'sleep', 'ensure-ready']) {
    const d = classify('POST', `/api/workers/bob/${action}`);
    assert.equal(d.ok, true, `${action} should be allowed`);
    assert.equal(d.route.kind, 'write');
    assert.equal(d.route.controllerPath, `/api/v1/workers/bob/${action}`);
    assert.equal(d.route.workerName, 'bob');
    assert.equal(d.route.action, action);
  }
});

test('GET on a lifecycle action path is rejected (write-only route)', () => {
  const d = classify('GET', '/api/workers/bob/wake');
  assert.equal(d.ok, false);
  assert.equal(d.status, 405);
});

test('unrelated worker sub-actions are not allow-listed', () => {
  // POST to any /api/workers/{name}/{action} other than wake/sleep/ensure-ready
  // is rejected. It surfaces as 405 (method not allowed) because the path
  // shape still matches the GET-listable "/api/workers/..." passthrough
  // family -- the important, tested property is that it is REJECTED, and
  // that no controller call is ever attempted for it (see handler.test.js).
  for (const action of ['ready', 'status', 'delete', 'restart']) {
    const d = classify('POST', `/api/workers/bob/${action}`);
    assert.equal(d.ok, false, `${action} must not be allowed`);
    assert.equal(d.status, 405);
  }
});

test('/api/tasks/* maps to shared/tasks/* on MinIO, GET only', () => {
  const d = classify('GET', '/api/tasks/task-20260703-000000/meta.json');
  assert.equal(d.ok, true);
  assert.equal(d.route.target, 'minio');
  assert.equal(d.route.minioKey, 'shared/tasks/task-20260703-000000/meta.json');

  const write = classify('POST', '/api/tasks/task-1/meta.json');
  assert.equal(write.ok, false);
  assert.equal(write.status, 405);
});

test('/api/files/* only allows shared/ and agents/ roots', () => {
  const shared = classify('GET', '/api/files/shared/projects/proj-1/manifest.json');
  assert.equal(shared.ok, true);
  assert.equal(shared.route.minioKey, 'shared/projects/proj-1/manifest.json');

  const agents = classify('GET', '/api/files/agents/manager/state.json');
  assert.equal(agents.ok, true);
  assert.equal(agents.route.minioKey, 'agents/manager/state.json');

  const badRoot = classify('GET', '/api/files/etc/passwd');
  assert.equal(badRoot.ok, false);
  assert.equal(badRoot.status, 404);

  const noRoot = classify('GET', '/api/files/');
  assert.equal(noRoot.ok, false);
});

test('path traversal in /api/tasks/* and /api/files/* is rejected', () => {
  const cases = [
    '/api/tasks/../../etc/passwd',
    '/api/files/shared/../../../etc/passwd',
    '/api/files/shared/%2e%2e/%2e%2e/secret',
  ];
  for (const p of cases) {
    const d = classify('GET', p);
    assert.equal(d.ok, false, `${p} should be rejected`);
    assert.equal(d.status, 404);
  }
});

test('/docker/ is never routable', () => {
  const d = classify('GET', '/docker/containers/json');
  assert.equal(d.ok, false);
  assert.equal(d.status, 404);
});
