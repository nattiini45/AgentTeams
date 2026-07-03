import { api } from '../api.js';
import { startPolling } from '../poll.js';
import { escapeHtml, badgeClass } from '../ui.js';

const POLL_MS = 15000;

/**
 * renderProjects mounts the Project browser: one card per Project CRD
 * (from GET /api/projects, Step 1), joined by id with the chat-flow layer
 * under shared/projects/{id}/meta.json + plan.md (decision #16 -- federated,
 * never schema-merged; we only ever render them side by side here). Task
 * progress is summarized as [ ]/[~]/[x] counts parsed out of plan.md, no
 * DAG/kanban (that's v2, explicitly deferred).
 */
export function renderProjects(root) {
  root.innerHTML = `
    <div class="section-title">Projects</div>
    <div id="projects-body"><div class="empty-state">Loading...</div></div>
  `;
  const body = root.querySelector('#projects-body');

  const stop = startPolling(
    async () => {
      const data = await api.listProjects();
      await renderList(body, data.projects || []);
    },
    POLL_MS,
    (err) => {
      body.innerHTML = `<div class="error-state">Failed to load: ${escapeHtml(err.message)}</div>`;
    },
  );

  return stop;
}

async function renderList(body, projects) {
  if (projects.length === 0) {
    body.innerHTML = '<div class="empty-state">No projects</div>';
    return;
  }

  const cards = await Promise.all(projects.map((p) => buildCard(p)));
  body.innerHTML = `<div class="card-grid">${cards.join('')}</div>`;
}

async function buildCard(project) {
  const id = project.name;
  let planCounts = null;
  let chatStatus = null;

  const [metaResult, planResult] = await Promise.allSettled([
    api.fileGetJson(['shared', 'projects', id, 'meta.json']),
    api.fileGetText(['shared', 'projects', id, 'plan.md']),
  ]);

  if (metaResult.status === 'fulfilled') {
    chatStatus = metaResult.value.status;
  }
  if (planResult.status === 'fulfilled') {
    planCounts = countTaskMarkers(typeof planResult.value === 'string' ? planResult.value : '');
  }

  const repoLines = (project.repos || [])
    .map((r) => `<li>${escapeHtml(r.name || r.url)} <span class="muted">(${escapeHtml(r.access)})</span></li>`)
    .join('');

  return `
    <div class="card">
      <div class="card-header">
        <span class="card-name">${escapeHtml(project.projectName || id)}</span>
        <span class="badge ${badgeClass(project.phase)}">${escapeHtml(project.phase)}</span>
      </div>
      <div class="card-meta">
        Team: ${escapeHtml(project.team)}<br/>
        Repos: ${project.repoCount ?? (project.repos || []).length}<br/>
        ${chatStatus ? `Chat-flow status: ${escapeHtml(chatStatus)}<br/>` : ''}
        ${planCounts ? renderProgress(planCounts) : '<span class="muted">plan.md not yet available</span>'}
      </div>
      ${repoLines ? `<ul class="card-meta">${repoLines}</ul>` : ''}
    </div>
  `;
}

function renderProgress(counts) {
  return `
    <div class="progress-counts">
      <span>[ ] ${counts.pending}</span>
      <span>[~] ${counts.inProgress}</span>
      <span>[x] ${counts.done}</span>
      ${counts.blocked ? `<span>[!] ${counts.blocked}</span>` : ''}
    </div>
  `;
}

/** countTaskMarkers scans plan.md lines for the task-status markers documented in
 * manager/agent/skills/project-management/references/plan-format.md. */
export function countTaskMarkers(markdown) {
  const counts = { pending: 0, inProgress: 0, done: 0, blocked: 0 };
  const lines = markdown.split('\n');
  for (const line of lines) {
    const m = line.match(/^\s*-\s*\[([ x~!→])\]/);
    if (!m) continue;
    switch (m[1]) {
      case ' ':
        counts.pending++;
        break;
      case '~':
        counts.inProgress++;
        break;
      case 'x':
        counts.done++;
        break;
      case '!':
        counts.blocked++;
        break;
      default:
        break;
    }
  }
  return counts;
}
