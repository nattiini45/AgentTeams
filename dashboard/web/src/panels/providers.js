import { api } from '../api.js';
import { startPolling } from '../poll.js';
import { escapeHtml, showToast, confirmAction } from '../ui.js';

const POLL_PROVIDERS_MS = 60000; // providers change rarely; avoid hammering Higress Console

/**
 * renderProviders mounts the AI provider management panel into `root`
 * and returns a cleanup function that stops polling.
 */
export function renderProviders(root) {
  root.innerHTML = `
    <div class="section-header">
      <div class="section-title">AI Providers</div>
      <button id="register-provider-btn" class="action-btn primary">Register Provider</button>
    </div>
    <div id="providers-cards" class="card-grid"><div class="empty-state">Loading...</div></div>
  `;

  const cardsEl = root.querySelector('#providers-cards');
  const registerBtn = root.querySelector('#register-provider-btn');

  registerBtn.addEventListener('click', () => openProviderDialog());

  const stop = startPolling(
    async () => {
      const data = await api.listProviders();
      renderProviderCards(cardsEl, data.providers || []);
    },
    POLL_PROVIDERS_MS,
    (err) => {
      cardsEl.innerHTML = `<div class="empty-state error">Failed to load providers: ${escapeHtml(err.message)}</div>`;
    },
  );

  function renderProviderCards(el, providers) {
    if (providers.length === 0) {
      el.innerHTML = '<div class="empty-state">No providers registered. Use the default provider or register a new one.</div>';
      return;
    }
    el.innerHTML = providers
      .map(
        (p) => `
      <div class="card" data-provider="${escapeHtml(p.name)}">
        <div class="card-header">
          <span class="card-name">${escapeHtml(p.name)}</span>
          <span class="badge badge-running">${escapeHtml(p.type || 'openai')}</span>
        </div>
        <div class="card-meta">
          ${p.route ? `Route: ${escapeHtml(p.route)}<br/>` : ''}
          Model prefix: <code>${escapeHtml(p.name)}/</code>
        </div>
        <div class="card-actions">
          <button class="action-btn danger" data-delete="${escapeHtml(p.name)}">Delete</button>
        </div>
      </div>`,
      )
      .join('');

    el.querySelectorAll('button[data-delete]').forEach((btn) => {
      btn.addEventListener('click', () => onDeleteProvider(btn.dataset.delete));
    });
  }

  async function onDeleteProvider(name) {
    const confirmed = await confirmAction(
      `Delete provider "${name}"? This removes its AI route, provider config, and DNS service source. Workers pinned to this provider will lose LLM access.`,
    );
    if (!confirmed) return;

    try {
      await api.deleteProvider(name);
      showToast(`Provider "${name}" deleted`);
      // Immediate re-fetch
      const data = await api.listProviders();
      renderProviderCards(cardsEl, data.providers || []);
    } catch (err) {
      showToast(`Delete failed: ${err.message}`, { error: true });
    }
  }

  function openProviderDialog() {
    const dialog = document.getElementById('provider-dialog');
    if (!dialog) return;
    const form = document.getElementById('provider-form');
    const nameInput = document.getElementById('provider-name');
    const urlInput = document.getElementById('provider-url');
    const keyInput = document.getElementById('provider-key');
    const cancelBtn = document.getElementById('provider-cancel');

    // Reset form
    nameInput.value = '';
    urlInput.value = '';
    keyInput.value = '';

    dialog.showModal();

    cancelBtn.onclick = () => dialog.close();

    form.onsubmit = async (e) => {
      e.preventDefault();
      const name = nameInput.value.trim();
      const url = urlInput.value.trim();
      const key = keyInput.value.trim();

      if (!name || !url || !key) return;

      dialog.close();

      try {
        await api.registerProvider(name, url, key);
        showToast(`Provider "${name}" registered`);
        // Immediate re-fetch
        const data = await api.listProviders();
        renderProviderCards(cardsEl, data.providers || []);
      } catch (err) {
        showToast(`Registration failed: ${err.message}`, { error: true });
      }
    };
  }

  return () => stop();
}
