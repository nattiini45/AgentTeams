import { api } from '../api.js';
import { startPolling } from '../poll.js';
import { escapeHtml, badgeClass } from '../ui.js';
import { freshnessFromAge, healthDotClass, formatAge, parseRFC3339 } from './overview-health.js';

const POLL_LIST_MS = 30000; // match the overview worker cadence (live Status() per member)
const POLL_DETAIL_MS = 15000;

const HEALTH_PROBE_ORDER = [
  ['container', 'Container'],
  ['heartbeat', 'Heartbeat'],
  ['llm', 'LLM'],
  ['git', 'Git'],
  ['sync', 'Sync'],
];

/**
 * renderWorkers mounts the per-worker overview: a list of all workers where
 * clicking one opens a drill-down detail view with current health, an
 * event/failure timeline (controller event history), and the worker's
 * better-harness findings. Returns a cleanup that stops all polling.
 */
export function renderWorkers(root) {
  root.innerHTML = `
    <div id="workers-list-view">
      <div class="section-title">Workers</div>
      <div id="workers-rows" class="card-grid"><div class="empty-state">Loading...</div></div>
    </div>
    <div id="worker-detail-view" hidden>
      <div class="breadcrumbs"><button id="worker-back">&larr; all workers</button></div>
      <div id="worker-detail-body"><div class="empty-state">Loading...</div></div>
    </div>
  `;

  const listView = root.querySelector('#workers-list-view');
  const detailView = root.querySelector('#worker-detail-view');
  const rowsEl = root.querySelector('#workers-rows');
  const detailBody = root.querySelector('#worker-detail-body');
  const backBtn = root.querySelector('#worker-back');

  let selectedWorker = null;
  let stopDetail = null;

  const stopList = startPolling(
    async () => {
      const data = await api.listWorkers();
      renderWorkerList(rowsEl, data.workers || [], openDetail);
    },
    POLL_LIST_MS,
    (err) => {
      rowsEl.innerHTML = `<div class="error-state">Failed to load: ${escapeHtml(err.message)}</div>`;
    },
  );

  function openDetail(name) {
    selectedWorker = name;
    listView.hidden = true;
    detailView.hidden = false;
    if (stopDetail) stopDetail();
    stopDetail = startPolling(
      () => loadAndRenderDetail(detailBody, name),
      POLL_DETAIL_MS,
      (err) => {
        detailBody.innerHTML = `<div class="error-state">Failed to load: ${escapeHtml(err.message)}</div>`;
      },
    );
  }

  function closeDetail() {
    selectedWorker = null;
    if (stopDetail) {
      stopDetail();
      stopDetail = null;
    }
    detailView.hidden = true;
    listView.hidden = false;
  }

  backBtn.addEventListener('click', closeDetail);

  return () => {
    stopList();
    if (stopDetail) stopDetail();
    backBtn.removeEventListener('click', closeDetail);
  };
}

// --- List view ---

function renderWorkerList(el, workers, onOpen) {
  if (workers.length === 0) {
    el.innerHTML = '<div class="empty-state">No workers</div>';
    return;
  }
  el.innerHTML = workers.map((w) => workerRow(w)).join('');
  el.querySelectorAll('[data-worker-name]').forEach((card) => {
    card.addEventListener('click', () => onOpen(card.dataset.workerName));
  });
}

function workerRow(w) {
  const health = w.healthState
    ? `<span class="badge badge-health-${escapeHtml(w.healthState)}">${escapeHtml(w.healthState)}</span>`
    : '';
  return `
    <div class="card worker-row" data-worker-name="${escapeHtml(w.name)}" role="button" tabindex="0">
      <div class="card-header">
        <span class="card-name">${escapeHtml(w.name)}</span>
        <span class="badge ${badgeClass(w.phase)}">${escapeHtml(w.phase)}</span>
        ${health}
      </div>
      <div class="card-meta">
        ${w.team ? `Team: ${escapeHtml(w.team)}<br/>` : ''}
        ${w.runtime ? `Runtime: ${escapeHtml(w.runtime)}<br/>` : ''}
        ${renderHealthStrip(w.healthChecks)}
      </div>
    </div>`;
}

function renderHealthStrip(checks) {
  if (!checks) {
    return 'Health: <span class="muted">unknown</span>';
  }
  const dots = HEALTH_PROBE_ORDER.map(([key, label]) => {
    const check = checks[key] || {};
    const status = check.status || 'unknown';
    const detail = check.detail ? ` — ${escapeHtml(check.detail)}` : '';
    const title = `${label}: ${status}${detail}`.replace(/"/g, '&quot;');
    return `<span class="health-dot ${healthDotClass(status)}" title="${title}" aria-label="${label} ${status}"></span>`;
  }).join('');
  return `<span class="health-strip-label">Health</span><span class="health-strip" aria-label="Worker health probes">${dots}</span>`;
}

// --- Detail view ---

async function loadAndRenderDetail(body, name) {
  // Status/events are independent reads; better-harness reports come from MinIO.
  const [status, events, harness] = await Promise.all([
    fetchOrNull(() => api.workerStatus(name)),
    fetchOrNull(() => api.workerEvents(name)),
    loadHarnessReports(name),
  ]);
  renderDetail(body, name, { status, events, harness });
}

async function fetchOrNull(fn) {
  try {
    return await fn();
  } catch (err) {
    if (err && err.status === 404) return null;
    throw err;
  }
}

/**
 * loadHarnessReports lists shared/harness-reports/<name>/ and reads the most
 * recent findings files. 404-tolerant: a worker with no reports yet yields an
 * empty list (rendered as an empty state, not an error).
 */
async function loadHarnessReports(name) {
  const listing = await fetchOrNull(() => api.fileBrowse(['shared', 'harness-reports', name]));
  if (!listing || !Array.isArray(listing.files)) return [];
  // Findings files are named <YYYYMMDD>.json; take the 5 most recent by name.
  const names = listing.files
    .map((f) => f.key)
    .filter((k) => /^\d{8}\.json$/.test(k))
    .sort()
    .reverse()
    .slice(0, 5);
  const reports = await Promise.all(
    names.map(async (fname) => {
      const findings = await fetchOrNull(() => api.fileGetJson(['shared', 'harness-reports', name, fname]));
      return { date: fname.replace(/\.json$/, ''), findings: normalizeFindings(findings) };
    }),
  );
  return reports;
}

function normalizeFindings(raw) {
  if (Array.isArray(raw)) return raw;
  if (raw && Array.isArray(raw.findings)) return raw.findings;
  return [];
}

function renderDetail(body, name, { status, events, harness }) {
  body.innerHTML = `
    <div class="section-title">${escapeHtml(name)}</div>
    ${renderStatusSection(status)}
    <div class="section-title">Event timeline</div>
    ${renderEventsSection(events)}
    <div class="section-title">Better-harness changes</div>
    ${renderHarnessSection(harness)}
  `;
}

function renderStatusSection(status) {
  if (!status) {
    return '<div class="empty-state">Status not available.</div>';
  }
  const heartbeat = freshnessFromAge(status.lastHeartbeat);
  const active = freshnessFromAge(status.lastActiveAt);
  return `
    <div class="card">
      <div class="card-header">
        <span class="badge ${badgeClass(status.phase)}">${escapeHtml(status.phase || 'unknown')}</span>
        ${status.healthState ? `<span class="badge badge-health-${escapeHtml(status.healthState)}">${escapeHtml(status.healthState)}</span>` : ''}
      </div>
      <div class="card-meta">
        Container: ${escapeHtml(status.containerState || 'unknown')}<br/>
        Heartbeat: ${heartbeat.age ?? 'never'} <span class="badge ${heartbeat.cls}">${heartbeat.label}</span><br/>
        Last active: ${active.age ?? 'never'} <span class="badge ${active.cls}">${active.label}</span><br/>
        ${renderHealthStrip(status.healthChecks)}<br/>
        ${status.message ? `Message: ${escapeHtml(status.message)}` : ''}
      </div>
    </div>`;
}

function renderEventsSection(events) {
  const list = (events && events.events) || [];
  if (list.length === 0) {
    return '<div class="empty-state">No events recorded yet.</div>';
  }
  return `
    <div class="event-list">
      ${list.map((ev) => renderEvent(ev)).join('')}
    </div>`;
}

function renderEvent(ev) {
  const when = formatEventTime(ev.timestamp);
  const isFailure = ev.type === 'failure';
  const cls = isFailure ? 'event-row event-failure' : 'event-row';
  const badge = isFailure
    ? '<span class="badge badge-failed">failure</span>'
    : `<span class="badge badge-unknown">${escapeHtml(ev.type || 'event')}</span>`;
  return `
    <div class="${cls}">
      <span class="event-time">${escapeHtml(when)}</span>
      ${badge}
      <span class="event-reason">${escapeHtml(ev.reason || '')}</span>
      ${ev.message ? `<span class="event-message">${escapeHtml(ev.message)}</span>` : ''}
    </div>`;
}

function renderHarnessSection(harness) {
  if (!harness || harness.length === 0) {
    return '<div class="empty-state">No better-harness reports yet. The weekly self-review has not reported for this worker.</div>';
  }
  return harness
    .map((report) => {
      const items = report.findings.length
        ? report.findings
            .map(
              (f) => `
          <li class="harness-finding">
            <span class="harness-finding-title">${escapeHtml(f.title || 'finding')}</span>
            ${f.summary ? `<div class="muted">${escapeHtml(f.summary)}</div>` : ''}
          </li>`,
            )
            .join('')
        : '<li class="muted">No findings recorded.</li>';
      return `
      <div class="card harness-report">
        <div class="card-header"><span class="card-name">Report ${escapeHtml(report.date)}</span></div>
        <ul class="harness-findings">${items}</ul>
      </div>`;
    })
    .join('');
}

function formatEventTime(ts) {
  const ms = parseRFC3339(ts);
  if (ms === null) return ts || 'unknown';
  return formatAge(ms);
}
