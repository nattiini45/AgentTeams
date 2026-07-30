import { api } from '../api.js';
import { startPolling } from '../poll.js';
import { escapeHtml, showToast, confirmAction, promptMessage, badgeClass } from '../ui.js';
import { freshnessFromAge, healthDotClass } from './overview-health.js';

const POLL_MANAGERS_TEAMS_MS = 15000;
const POLL_WORKERS_MS = 30000; // workers list does a live backend Status() call per member -- slower cadence

/**
 * renderOverview mounts the Managers/Teams/Workers card grids into `root`
 * and returns a cleanup function that stops all polling (call it when the
 * tab is switched away from).
 */
export function renderOverview(root) {
  root.innerHTML = `
    <div id="managers-section">
      <div class="section-title">Managers</div>
      <div id="managers-cards" class="card-grid"><div class="empty-state">Loading...</div></div>
    </div>
    <div id="teams-section">
      <div class="section-title">Teams</div>
      <div id="teams-cards" class="card-grid"><div class="empty-state">Loading...</div></div>
    </div>
    <div id="workers-section">
      <div class="section-title">Workers</div>
      <div id="workers-cards" class="card-grid"><div class="empty-state">Loading...</div></div>
    </div>
  `;

  const managersEl = root.querySelector('#managers-cards');
  const teamsEl = root.querySelector('#teams-cards');
  const workersEl = root.querySelector('#workers-cards');

  const stopManagers = startPolling(
    async () => {
      const data = await api.listManagers();
      renderManagerCards(managersEl, data.managers || []);
    },
    POLL_MANAGERS_TEAMS_MS,
    (err) => renderError(managersEl, err),
  );

  const stopTeams = startPolling(
    async () => {
      const data = await api.listTeams();
      renderTeamCards(teamsEl, data.teams || []);
    },
    POLL_MANAGERS_TEAMS_MS,
    (err) => renderError(teamsEl, err),
  );

  const stopWorkers = startPolling(
    async () => {
      const data = await api.listWorkers();
      renderWorkerCards(workersEl, data.workers || []);
    },
    POLL_WORKERS_MS,
    (err) => renderError(workersEl, err),
  );

  return () => {
    stopManagers();
    stopTeams();
    stopWorkers();
  };
}

function renderError(el, err) {
  el.innerHTML = `<div class="error-state">Failed to load: ${escapeHtml(err.message)}</div>`;
}

function renderManagerCards(el, managers) {
  if (managers.length === 0) {
    el.innerHTML = '<div class="empty-state">No managers</div>';
    return;
  }
  el.innerHTML = managers
    .map(
      (m) => `
    <div class="card" data-manager="${escapeHtml(m.name)}">
      <div class="card-header">
        <span class="card-name">${escapeHtml(m.name)}</span>
        <span class="badge ${badgeClass(m.phase)}">${escapeHtml(m.phase)}</span>
      </div>
      <div class="card-meta">
        ${m.model ? `Model: ${escapeHtml(m.model)}<br/>` : ''}
        ${m.modelProvider ? `Provider: ${escapeHtml(m.modelProvider)}<br/>` : ''}
        ${m.runtime ? `Runtime: ${escapeHtml(m.runtime)}<br/>` : ''}
        State: ${escapeHtml(m.state || 'unknown')}<br/>
        Welcome sent: ${m.welcomeSent ? 'yes' : 'no'}
      </div>
      <div class="card-actions">
        <button class="action-btn" data-message-kind="manager" data-name="${escapeHtml(m.name)}">Message</button>
        <button class="action-btn" data-set-provider-kind="managers" data-name="${escapeHtml(m.name)}" data-current-model="${escapeHtml(m.model || '')}" data-current-provider="${escapeHtml(m.modelProvider || '')}">Set Provider</button>
      </div>
    </div>`,
    )
    .join('');

  el.querySelectorAll('button[data-message-kind]').forEach((btn) => {
    btn.addEventListener('click', () => onMessageAction(btn));
  });
  el.querySelectorAll('button[data-set-provider-kind]').forEach((btn) => {
    btn.addEventListener('click', () => onSetProvider(btn));
  });
}

function renderTeamCards(el, teams) {
  if (teams.length === 0) {
    el.innerHTML = '<div class="empty-state">No teams</div>';
    return;
  }
  el.innerHTML = teams
    .map(
      (t) => `
    <div class="card" data-team="${escapeHtml(t.name)}">
      <div class="card-header">
        <span class="card-name">${escapeHtml(t.name)}</span>
        <span class="badge ${badgeClass(t.phase)}">${escapeHtml(t.phase)}</span>
      </div>
      <div class="card-meta">
        Leader: ${escapeHtml(t.leaderName || 'n/a')}<br/>
        ${t.modelProvider ? `Provider: ${escapeHtml(t.modelProvider)}<br/>` : ''}
        Workers ready: ${t.readyWorkers ?? 0} / ${t.totalWorkers ?? 0}<br/>
        Leader ready: ${t.leaderReady ? 'yes' : 'no'}
      </div>
      <div class="card-actions">
        <button class="action-btn" data-message-kind="team" data-name="${escapeHtml(t.name)}">Message</button>
        <button class="action-btn" data-set-provider-kind="teams" data-name="${escapeHtml(t.name)}" data-current-provider="${escapeHtml(t.modelProvider || '')}">Set Provider</button>
      </div>
    </div>`,
    )
    .join('');

  el.querySelectorAll('button[data-message-kind]').forEach((btn) => {
    btn.addEventListener('click', () => onMessageAction(btn));
  });
  el.querySelectorAll('button[data-set-provider-kind]').forEach((btn) => {
    btn.addEventListener('click', () => onSetProvider(btn));
  });
}

function renderWorkerCards(el, workers) {
  if (workers.length === 0) {
    el.innerHTML = '<div class="empty-state">No workers</div>';
    return;
  }
  el.innerHTML = workers
    .map(
      (w) => `
    <div class="card" data-worker="${escapeHtml(w.name)}">
      <div class="card-header">
        <span class="card-name">${escapeHtml(w.name)}</span>
        <span class="badge ${badgeClass(w.phase)}">${escapeHtml(w.phase)}</span>
        ${w.healthState ? `<span class="badge badge-health-${escapeHtml(w.healthState)}">${escapeHtml(w.healthState)}</span>` : ''}
      </div>
      <div class="card-meta">
        ${w.team ? `Team: ${escapeHtml(w.team)}<br/>` : ''}
        ${w.modelProvider ? `Provider: ${escapeHtml(w.modelProvider)}<br/>` : ''}
        State: ${escapeHtml(w.state || 'unknown')}<br/>
        Container: ${escapeHtml(w.containerState || 'unknown')}<br/>
        ${renderFreshnessLine('Heartbeat', w.lastHeartbeat)}<br/>
        ${renderFreshnessLine('Last active', w.lastActiveAt)}<br/>
        ${renderLLMUsageLine(w.llmCallsLastHeartbeat, w.llmCallsTotal)}<br/>
        ${renderHealthStrip(w.healthChecks)}
      </div>
      <div class="card-actions">
        <button class="action-btn" data-action="wake" data-name="${escapeHtml(w.name)}">Wake</button>
        <button class="action-btn" data-action="sleep" data-name="${escapeHtml(w.name)}">Sleep</button>
        <button class="action-btn" data-action="ensure-ready" data-name="${escapeHtml(w.name)}">Ensure Ready</button>
        <button class="action-btn" data-set-provider-kind="workers" data-name="${escapeHtml(w.name)}" data-current-model="${escapeHtml(w.model || '')}" data-current-provider="${escapeHtml(w.modelProvider || '')}">Set Provider</button>
      </div>
    </div>`,
    )
    .join('');

  el.querySelectorAll('button.action-btn[data-action]').forEach((btn) => {
    btn.addEventListener('click', () => onWorkerAction(btn));
  });
  el.querySelectorAll('button[data-set-provider-kind]').forEach((btn) => {
    btn.addEventListener('click', () => onSetProvider(btn));
  });
}

function renderFreshnessLine(label, ts) {
  const health = freshnessFromAge(ts);
  const ageText = health.age ?? 'never';
  return `${label}: ${ageText} <span class="badge ${health.cls}">${health.label}</span>`;
}

function renderLLMUsageLine(lastWindow, total) {
  if (lastWindow == null && (total == null || total === 0)) {
    return 'LLM calls: <span class="muted">unknown</span>';
  }
  const windowText = lastWindow == null ? 'unknown' : String(lastWindow);
  const totalText = total == null || total === 0 ? '' : ` (total ${total})`;
  return `LLM calls (last window): ${windowText}${totalText}`;
}

const HEALTH_PROBE_ORDER = [
  ['container', 'Container'],
  ['heartbeat', 'Heartbeat'],
  ['llm', 'LLM'],
  ['git', 'Git'],
  ['sync', 'Sync'],
];

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

async function onMessageAction(btn) {
  const kind = btn.dataset.messageKind; // 'manager' | 'team'
  const name = btn.dataset.name;
  const body = await promptMessage(`Send a message to ${kind} "${name}":`);
  if (!body) return;

  btn.disabled = true;
  try {
    const res = kind === 'manager' ? await api.messageManager(name, body) : await api.messageTeam(name, body);
    showToast(`${name}: message sent to ${res.roomID}`);
  } catch (err) {
    if (err.status === 409) {
      showToast(`${name}: room not provisioned yet`, { error: true });
    } else {
      showToast(`${name}: message failed -- ${err.message}`, { error: true });
    }
  } finally {
    btn.disabled = false;
  }
}

async function onWorkerAction(btn) {
  const action = btn.dataset.action;
  const name = btn.dataset.name;
  const verbs = { wake: 'wake', sleep: 'sleep', 'ensure-ready': 'ensure-ready for' };
  const confirmed = await confirmAction(`${verbs[action]} worker "${name}"?`);
  if (!confirmed) return;

  const allButtons = btn.closest('.card-actions').querySelectorAll('button');
  allButtons.forEach((b) => (b.disabled = true));
  try {
    if (action === 'wake') await api.wake(name);
    else if (action === 'sleep') await api.sleep(name);
    else if (action === 'ensure-ready') await api.ensureReady(name);
    showToast(`${name}: ${action} sent`);
  } catch (err) {
    showToast(`${name}: ${action} failed -- ${err.message}`, { error: true });
  } finally {
    allButtons.forEach((b) => (b.disabled = false));
  }
}

// --- Model/Provider dialog ---

const mpDialog = document.getElementById('model-provider-dialog');
const mpForm = document.getElementById('model-provider-form');
const mpTitle = document.getElementById('model-provider-title');
const mpModel = document.getElementById('mp-model');
const mpCustomRow = document.getElementById('mp-custom-row');
const mpCustomModel = document.getElementById('mp-custom-model');
const mpProvider = document.getElementById('mp-provider');
const mpCancel = document.getElementById('mp-cancel');

let mpResolve = null; // resolves { model, modelProvider } or null (cancelled)

mpCancel.addEventListener('click', () => {
  if (mpResolve) mpResolve(null);
  mpResolve = null;
  mpDialog.close();
});

mpForm.addEventListener('submit', () => {
  if (!mpResolve) return;
  const provider = mpProvider.value; // "" for default
  const isCustom = mpModel.value === '__custom__';
  const bareModel = isCustom ? mpCustomModel.value.trim() : mpModel.value;
  // Prefix logic: default provider sends bare model; extras get provider/model
  const finalModel = provider ? `${provider}/${bareModel}` : bareModel;
  mpResolve({ model: finalModel, modelProvider: provider });
  mpResolve = null;
});

mpDialog.addEventListener('cancel', () => {
  if (mpResolve) mpResolve(null);
  mpResolve = null;
});

// Toggle custom input visibility based on model select value.
mpModel.addEventListener('change', () => {
  mpCustomRow.style.display = mpModel.value === '__custom__' ? '' : 'none';
});

// Re-fetch models when provider changes.
mpProvider.addEventListener('change', () => {
  loadModelsForProvider(mpProvider.value, '', '');
});

/**
 * Fetch models for the given provider value and populate #mp-model.
 * providerValue: "" = default (maps to openai-compat), else the provider name.
 * currentModel: bare model id to pre-select ("" = none).
 */
async function loadModelsForProvider(providerValue, currentModel, currentProvider) {
  const fetchName = providerValue || 'openai-compat';
  mpModel.innerHTML = '<option value="">Loading…</option>';
  mpModel.disabled = true;
  mpCustomRow.style.display = 'none';

  try {
    const data = await api.listProviderModels(fetchName);
    const models = data.models || [];
    mpModel.innerHTML = '';
    for (const id of models) {
      const opt = document.createElement('option');
      opt.value = id;
      opt.textContent = id;
      mpModel.appendChild(opt);
    }
    // Add Custom… option
    const customOpt = document.createElement('option');
    customOpt.value = '__custom__';
    customOpt.textContent = 'Custom…';
    mpModel.appendChild(customOpt);

    // Pre-select current model if it's in the list
    if (currentModel) {
      const found = models.includes(currentModel);
      if (found) {
        mpModel.value = currentModel;
      } else {
        mpModel.value = '__custom__';
        mpCustomModel.value = currentModel;
        mpCustomRow.style.display = '';
      }
    }
  } catch {
    // Fetch failed — fall back to custom-only
    mpModel.innerHTML = '';
    const customOpt = document.createElement('option');
    customOpt.value = '__custom__';
    customOpt.textContent = 'Custom…';
    mpModel.appendChild(customOpt);
    mpModel.value = '__custom__';
    if (currentModel) {
      mpCustomModel.value = currentModel;
      mpCustomRow.style.display = '';
    }
    showToast('Could not load model list — enter model manually.', { error: true });
  } finally {
    mpModel.disabled = false;
  }
}

function openModelProviderDialog(title, currentModel, currentProvider, providers) {
  mpTitle.textContent = title;

  // Populate provider select
  mpProvider.innerHTML = '<option value="">(default)</option>';
  for (const p of providers) {
    const opt = document.createElement('option');
    opt.value = p.name;
    opt.textContent = p.name;
    if (p.name === currentProvider) opt.selected = true;
    mpProvider.appendChild(opt);
  }

  // Strip provider prefix from currentModel to get the bare id
  let bareModel = currentModel || '';
  if (currentProvider && bareModel.startsWith(currentProvider + '/')) {
    bareModel = bareModel.slice(currentProvider.length + 1);
  }

  // Reset custom input
  mpCustomModel.value = '';
  mpCustomRow.style.display = 'none';

  // Load models for the selected provider
  const selectedProvider = currentProvider || '';
  loadModelsForProvider(selectedProvider, bareModel, currentProvider);

  return new Promise((resolve) => {
    mpResolve = resolve;
    mpDialog.showModal();
  });
}

async function onSetProvider(btn) {
  const kind = btn.dataset.setProviderKind; // 'workers' | 'teams' | 'managers'
  const name = btn.dataset.name;
  const currentModel = btn.dataset.currentModel || '';
  const currentProvider = btn.dataset.currentProvider || '';

  let providers;
  try {
    const data = await api.listProviders();
    providers = data.providers || [];
  } catch (err) {
    showToast(`Failed to load providers: ${err.message}`, { error: true });
    return;
  }

  const kindLabel = kind === 'workers' ? 'worker' : kind === 'teams' ? 'team' : 'manager';
  const result = await openModelProviderDialog(
    `Set model / provider for ${kindLabel} "${name}"`,
    currentModel,
    currentProvider,
    providers,
  );
  if (!result) return; // cancelled

  const { model, modelProvider } = result;
  if (!model && !modelProvider) {
    showToast('Nothing to update — both fields empty.', { error: true });
    return;
  }

  btn.disabled = true;
  try {
    if (kind === 'teams') {
      // Teams: modelProvider is applied to the leader via the backend.
      const patch = {};
      if (modelProvider) patch.modelProvider = modelProvider;
      await api.updateTeam(name, patch);
    } else {
      const patch = {};
      if (model) patch.model = model;
      if (modelProvider) patch.modelProvider = modelProvider;
      if (kind === 'workers') await api.updateWorker(name, patch);
      else await api.updateManager(name, patch);
    }
    const parts = [];
    if (model) parts.push(`model=${model}`);
    if (modelProvider) parts.push(`provider=${modelProvider}`);
    showToast(`${name}: updated (${parts.join(', ')})`);
  } catch (err) {
    showToast(`${name}: update failed -- ${err.message}`, { error: true });
  } finally {
    btn.disabled = false;
  }
}
